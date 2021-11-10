package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ghodss/yaml"

	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	_ "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-tools/pkg/crd"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
	"sigs.k8s.io/controller-tools/pkg/version"

	_ "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

func main() {
	roots, err := loader.LoadRoots(
		"k8s.io/apimachinery/pkg/runtime/schema", // Needed to parse generated register functions.
		"sigs.k8s.io/gateway-api/apis/v1alpha2",
	)

	if err != nil {
		log.Fatalf("failed to load package roots: %s", err)
	}

	// Cribbed from https://github.com/kubernetes-sigs/controller-tools/blob/master/pkg/crd/gen.go.

	generator := &crd.Generator{}

	parser := &crd.Parser{
		Collector: &markers.Collector{Registry: &markers.Registry{}},
		Checker: &loader.TypeChecker{
			NodeFilters: []loader.NodeFilter{generator.CheckFilter()},
		},
	}

	generator.RegisterMarkers(parser.Collector.Registry)

	for _, r := range roots {
		parser.NeedPackage(r)
	}

	metav1Pkg := crd.FindMetav1(roots)
	if metav1Pkg == nil {
		log.Fatalf("no objects in the roots, since nothing imported metav1")
	}

	kubeKinds := crd.FindKubeKinds(parser, metav1Pkg)
	if len(kubeKinds) == 0 {
		log.Fatalf("no objects in the roots")
	}

	for pkg := range parser.GroupVersions {
		for _, err := range pkg.Errors {
			log.Printf("%s: %s\n", pkg.PkgPath, err)
		}
	}

	for groupKind := range kubeKinds {
		log.Printf("generating CRD for %v\n", groupKind)

		parser.NeedCRDFor(groupKind, nil)
		crdRaw := parser.CustomResourceDefinitions[groupKind]

		// Inline version of "addAttribution(&crdRaw)" ...
		if crdRaw.ObjectMeta.Annotations == nil {
			crdRaw.ObjectMeta.Annotations = map[string]string{}
		}
		crdRaw.ObjectMeta.Annotations["controller-gen.kubebuilder.io/version"] = version.Version()

		// Prevent the top level metadata for the CRD to be generated regardless of the intention in the arguments
		crd.FixTopLevelMetadata(crdRaw)

		conv, err := crd.AsVersion(crdRaw, apiext.SchemeGroupVersion)
		if err != nil {
			log.Fatalf("failed to convert CRD: %s", err)
		}

		if def, ok := conv.(*apiext.CustomResourceDefinition); ok {
			def.Annotations[apiext.KubeAPIApprovedAnnotation] = "https://github.com/kubernetes-sigs/gateway-api/pull/891"
		}

		// XXX(jpeach) removeDescriptionFromMetadata(crd.(*apiext.CustomResourceDefinition))
		out, err := yaml.Marshal(conv)
		if err != nil {
			log.Fatalf("failed to marhal CRD: %s", err)
		}

		fmt.Println("---")
		os.Stdout.Write(out)
	}
}

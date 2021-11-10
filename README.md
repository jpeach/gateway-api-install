# gateway-api-install

This reppository is a litle experiment with generating CRDs for
[Gateway API](https://gateway-api.sigs.k8s.io/).

Any controller that needs to support Gateway API might need to install
the CRDs. They probably want to install the CRDs at a supported version,
and they also probably already import the corresponding module with the Go
types. So why should they also have to carry around a big bundle of YAML
to feed to Helm or kubectl? It seems like we should be able to generate
the right YAML on demand from the Go types - after all that's what
[crd-gen](https://github.com/kubernetes-sigs/controller-tools/blob/master/pkg/crd/gen.go)
is doing.

So this repo janks just enough code out of crd-gen to generate the v1 CRD
definitions. This is good enough to pass YAML to a tool, or to directly
create the API definitions if you have a Kubernetes API client handy.

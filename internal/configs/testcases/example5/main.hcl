kube_resource "namespace" {
  apiVersion = "v1"
  kind       = "Namespace"
  metadata = {
    name = "foo"
    labels = {
      name = "bar"
    }
  }

}
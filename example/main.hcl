resource "namespace" {
  apiVersion = "v1"
  kind       = "Namespace"
  metadata = {
    name = "bla"
    labels = {
      name = "bla"
    }
  }

}
resource "foo" {
  for_each = {
    "foo" = "bar",
    "bar" = "foo"
  }
  apiVersion = "apps/v1"
  kind       = "Deployment"
  metadata = {
    name = each.key
    labels = {
      app = each.value
    }
  }
  spec = {
    replicas = 3
    selector = {
      matchLabels = {
        app = each.value
      }
    }
    template = {
      metadata = {
        labels = {
          app = each.value
        }
      }
      spec = {
        containers = [{
          name  = each.value
          image = "nginx:1.14.2"
          ports = var.foo
        }]
      }
    }
  }
  depends_on = [module.test, resource.namespace]
}

module "test" {

  source = "./modules/starter"
  foo    = ["service1", "service2"]
  ports  = var.foo
  # depends_on = [resource.t,resource.a]
  depends_on = [resource.namespace]
}

default_annotations {
  foo = "bar"
}
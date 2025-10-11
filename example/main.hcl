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

kube_resource "foo" {
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
    replicas = 2
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
  depends_on = [module.test, kube_resource.namespace]
}

module "test" {
  source     = "./modules/starter"
  foo        = ["service1", "service2"]
  ports      = var.foo
  depends_on = [kube_resource.namespace]
}

default_annotations {
  foo = "bar"
}

default_annotations {
  bar = "foo"
}

# backend_storage {
##   stateless {}
#   kube_secret {}
# }
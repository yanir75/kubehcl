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
  depends_on = [module.test]
}

module "test" {
    source = "./modules/starter"
    foo = ["service1","service2"]
    ports = var.foo
    # depends_on = [resource.t,resource.a]
}

default_annotations {
  foo = "bar"
}

resource "bar" {
  count = 0
  apiVersion = "apps/v1"
  kind       = "Deployment"
  metadata = {
    name = "shouldntbecreated"
    labels = {
      app = "shouldntbecreated"
    }
  }
  spec = {
    replicas = 3
    selector = {
      matchLabels = {
        app = "shouldntbecreated"
      }
    }
    template = {
      metadata = {
        labels = {
          app = "shouldntbecreated"
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
}
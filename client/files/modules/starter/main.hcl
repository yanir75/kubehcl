locals {
  service_ports = {
    for i in range(length(var.foo)) : "${i}" => {
      name       = var.foo[i]
      targetPort = 80
    }
  }

  other_option = {
    for name in var.foo : name => {
      targetPort = 80
    }
  }
}

kube_resource "service" {
  for_each   = local.service_ports
  apiVersion = "v1"
  kind       = "Service"
  metadata = {
    name = each.value["name"]
  }
  spec = {
    selector = {
      "app.kubernetes.io/name" = each.value["name"]
    }
    ports = [merge(each.value, { port = 9367 })]
  }
  depends_on = [kube_resource.foo]
}

kube_resource "foo" {
  for_each   = local.service_ports
  apiVersion = "apps/v1"
  kind       = "Deployment"
  metadata = {
    name = each.value["name"]
    labels = {
      "app.kubernetes.io/name" = each.value["name"]
    }
  }
  spec = {
    replicas = 3
    selector = {
      matchLabels = {
        "app.kubernetes.io/name" = each.value["name"]
      }
    }
    template = {
      metadata = {
        labels = {
          "app.kubernetes.io/name" = each.value["name"]
        }
      }
      spec = {
        containers = [{
          name  = each.value["name"]
          image = "nginx:1.14.2"
          ports = var.ports
        }]
      }
    }
  }
}

module "secret" {
  source     = "./modules/secret"
  depends_on = [kube_resource.bar]
}

kube_resource "bar" {
  apiVersion = "apps/v1"
  kind       = "Deployment"
  metadata = {
    name = "foobar"
    labels = {
      "app.kubernetes.io/name" = "foobar"
    }
  }
  spec = {
    replicas = 3
    selector = {
      matchLabels = {
        "app.kubernetes.io/name" = "foobar"
      }
    }
    template = {
      metadata = {
        labels = {
          "app.kubernetes.io/name" = "foobar"
        }
      }
      spec = {
        containers = [{
          name  = "foobar"
          image = "nginx:1.14.2"
          ports = var.ports
        }]
      }
    }
  }
}


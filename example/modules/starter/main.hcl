locals {
    service_ports = {
        for i in range(length(var.foo)) : "${i}" => {
            name = var.foo[i]
            targetPort = 9376
            }
    }

    other_option = {
        for name in var.foo :   name => {
                targetPort = 9376
            }
    }
}

resource "service" {
    for_each = local.service_ports
    apiVersion= "v1"
    kind= "Service"
    metadata= {
        name = each.value["name"]
    }
    spec= {
        selector = {
            "app.kubernetes.io/name" = each.value["name"]
        }
        ports= [merge(each.value,{port = 80})]
    }
}

module "secret" {
    source = "./modules/secret"
}
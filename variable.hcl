# # variable "test" {
# #     type = any
# #     default =  {"blasd" = "asdf"}
# #     description = "test"
# # }



# # resource "asdf" "test" {
# #     test = local.g
# # }

# variable "bla" {
#     type = map(string)

#     default = {"test":"test"}
# }
# # locals {
# #     g = [for key,value in var.bla :value]
# #     a = "test"
# # }

default_annotations {
    t = "test"
}

variable "bla" {

    type = list(map(number))
    default = [{
             containerPort = 80   
            }]
    description = "test"
}

resource "nane" {
  for_each = {"key" = "tasd","key2" = "asdf"}
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
          ports = var.bla
        }]
      }
    }
  }
}

# resource "t" {
#   apiVersion = "apps/v1"
#   kind       = "Deployment"
#   # metadata = "test"
# }
# resource "g" {
#     for_each = toset(["test","asdf"])
#     test = "asdf"
#     dynamic "b" {
#         for_each = toset(["ta","test"])
#         content {
#             test = "tesafdsaf"
#         }
#     }

# }

# resource "t" {
#     depends_on = [module.test]

# }

# resource "a" {
#     # count = 2
#     # # spec {
#     # #     t = "testing"
#     # #     i = "bla"
#     # # }
#     # test = count.index
# }

# resource "b" {
#     spec {
#         t = "testing"
#         i = "kjh"
#         agsda {
#             fgfhghgf = "tzdsssgest"
#         }
#     }
#     test = "bla"
# }


# module "test" {
#     source = "./"
#     bla = {"testing" = "testing"}
#     # depends_on = [resource.t,resource.a]
# }

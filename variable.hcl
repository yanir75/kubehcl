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

# default_annotations {
#     # t = var.bla
# }

variable "bla" {

    type = any
    default = {
                ra = {"gasdfasd"="gasdfas"},

                werwer = "grewrw"
            }
    description = "test"
}

resource "nane" {
  apiVersion = "apps/v1"
  kind       = "Deployment"
  metadata = {
    name = "nginx-deployment"
    labels = {
      app = "nginx"
    }
  }
  spec = {
    replicas = 3
    selector = {
      matchLabels = {
        app = "nginx"
      }
    }
    template = {
      metadata = {
        labels = {
          app = "nginx"
        }
      }
      spec = {
        containers = [{
          name  = "nginx"
          image = "nginx:1.14.2"
          ports = [
            {
              "containerPort" = 80
            }
          ]
        }]
      }
    }
  }
}


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
#     source = "./test"
#     bla = {"testing" = "testing"}
#     # depends_on = [resource.t,resource.a]
# }

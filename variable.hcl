# variable "test" {
#     type = any
#     default =  {"blasd" = "asdf"}
#     description = "test"
# }

variable "bla" {

    type = any
    default = {
                ra = {"asdf"="asdgasdfa"},

                werwer = "grewrw"
            }
    description = "test"
}

# resource "asdf" "test" {
#     test = local.g
# }

locals {
    g = [for key,value in var.bla :value]
    a = "test"
}

default_annotations {
t = "test"
}


# resource "nane" {
#     count = 1
#     depends_on = [resource.t]
# }


resource "t" {
    for_each = var.bla
    po = each.key
    bla = each.value
    # depends_on = [resource.a]
}

resource "a" {
    count = 2
    # spec {
    #     t = "testing"
    #     i = "bla"
    # }
    test = count.index
}

resource "b" {
    spec {
        t = "testing"
        i = "kjh"
        agsda {
            fgfhghgf = "tzdsssgest"
        }
    }
    test = "bla"
}


module "test" {
    source = "./test"
}

variable "test" {
    type = any
    default =  {"blasd" = "asdf"}
    description = "test"
}

variable "bla" {

    type = any
    default = {"ra" = {"test"="test"}}
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

}


resource "nane" {
    count = 1
    depends_on = [resource.t]
}


resource "t" {
    count = 1
    depends_on = [resource.a]
}

resource "a" {
    count = 1
}



variable "test" {
    type = any
    default =  {"blasd" = "asdf"}
    description = "test"
}

variable "bla" {
    type = any
    default = {"ra" = "gra"}
    description = "test"
}

resource "asdf" "test" {
    test = local.g
}

locals {
    g = [for key,value in var.bla :value]
}


variable "foo" {

  type        = list(string)
  description = "Names of the services"
}

variable "ports" {

  # type = list(map(number))
  # default = var.foo
  # default = [{
  #          containerPort = 80   
  #         }]
  # type = string
  description = "Ports of the container"
}
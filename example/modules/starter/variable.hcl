variable "foo" {

    type = list(string)
    description = "Names of the services"
}

variable "ports" {

    type = list(map(number))
    default = [{
             containerPort = 80   
            }]
    description = "Ports of the container"
}


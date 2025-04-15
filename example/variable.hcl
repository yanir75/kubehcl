variable "foo" {

    type = list(map(number))
    # default = [{
    #          containerPort = 80   
    #         }]
    description = "Ports of the container"
}


variable "foo" {
  # type = string
  type = list(map(string))
  default = [{
    containerPort = "80"
  }]
  description = "Ports of the container"
}
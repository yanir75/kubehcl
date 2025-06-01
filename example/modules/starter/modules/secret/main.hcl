resource "secret" {
  apiVersion = "v1"
  kind       = "Secret"
  metadata = {
    name = "dotfile-secret"
  }
  data = {
    ".secret-file" = base64encode("Hello World")
  }
}
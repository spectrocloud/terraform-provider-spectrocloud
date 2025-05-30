data "spectrocloud_macros" "macros" {
  context = "tenant"
}

output "macros" {
  value = data.spectrocloud_macros.macros.macros
  description = "Available macros"
}


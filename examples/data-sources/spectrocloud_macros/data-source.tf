data "spectrocloud_macros" "project" {
  project = "Default"
}

output "available_macros_project" {
  value = data.spectrocloud_macros.project.macros
}

data "spectrocloud_macros" "tenant" {

}
output "available_macros_tenant" {
  value = data.spectrocloud_macros.tenant.macros
}
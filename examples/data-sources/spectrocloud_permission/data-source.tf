data "spectrocloud_permission" "app_permission" {
  name = "App Profile"

}

output "permissions" {
  value = data.spectrocloud_permission.app_permission.permissions
}
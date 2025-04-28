resource "spectrocloud_appliance" "appliance" {
  uid = "test-dec9"
  tags = {
    "name" = "appliance_name"
  }
  wait                        = true
  remote_shell                = "disabled"
  temporary_shell_credentials = "disabled"
}
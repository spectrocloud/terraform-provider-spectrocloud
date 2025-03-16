data "spectrocloud_appliance" "test_appliance" {
  id = "test-dec9"
}

output "same" {
  value = data.spectrocloud_appliance.test_appliance
}

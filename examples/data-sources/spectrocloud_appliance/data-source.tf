data "spectrocloud_appliance" "test_appliance" {
  name = "nik-test-1"
}

output "same" {
  value = data.spectrocloud_appliance.test_appliance
}

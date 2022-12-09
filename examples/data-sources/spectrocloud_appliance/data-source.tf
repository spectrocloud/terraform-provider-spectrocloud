data "spectrocloud_appliance" "test_appliance" {
  name = "dev_502"
}

output "same" {
  value = data.spectrocloud_appliance.test_appliance
}

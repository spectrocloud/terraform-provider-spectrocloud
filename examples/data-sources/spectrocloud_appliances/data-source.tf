data "spectrocloud_appliances" "appliances" {
  #tags = ["env:prod", "store:502"]
  tags = ["env:dev"]
}

output "same" {
  value = data.spectrocloud_appliances.appliances
  #value = [for a in data.spectrocloud_appliance.appliances : a.name]
}

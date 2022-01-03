data "spectrocloud_project" "project_default" {
  name = "Default"
}

data "spectrocloud_appliance" "test_appliance" {
  name = "test-appliance-id"
  project_id = data.spectrocloud_project.project_default.id
}

output "same" {
  value = data.spectrocloud_appliance.test_appliance
}

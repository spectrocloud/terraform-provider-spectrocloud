data "spectrocloud_appliance" "virt_appliance" {
  name = "0ff1e31ab8898263eccb"
  project_id = data.spectrocloud_project.project_default.id
}
data "spectrocloud_workspace" "workspace" {
  name = "wsp-tf"
}

output "same" {
  value = data.spectrocloud_workspace.workspace
}

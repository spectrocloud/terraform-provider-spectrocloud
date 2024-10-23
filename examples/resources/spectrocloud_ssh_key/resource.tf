resource "spectrocloud_ssh_key" "ssh_project" {
  name    = "ssh-dev-1-project"
  context = "project"
  ssh_key = var.ssh_key_value
}
resource "spectrocloud_ssh_key" "ssh_tenant" {
  name    = "ssh-dev-1"
  context = "tenant"
  ssh_key = var.ssh_key_value
}
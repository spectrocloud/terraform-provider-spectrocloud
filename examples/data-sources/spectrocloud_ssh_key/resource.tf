data "spectrocloud_ssh_key" "ssh_project" {
  name    = "test-tf-ssh"
  context = "project"
}

resource "spectrocloud_ssh_key" "ssh_tenant" {
  name    = "ssh-dev-1"
  context = "tenant"
  ssh_key = data.spectrocloud_ssh_key.ssh_project.ssh_key
}
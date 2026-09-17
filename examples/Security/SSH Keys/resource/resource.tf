# Day-2 mutability: nothing on this resource is ForceNew - name, ssh_key, and context all
# update in place.
#
# Attributes:
#   name    - Required. The SSH key asset's name.
#   context - Optional, default "project". Allowed: "project", "tenant".
#   ssh_key - Required (credential material, not secret-sensitive by nature but marked
#             Sensitive). Public key in "authorized_keys" format, e.g. "ssh-rsa AAAAB3Nza...".
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

# terraform import spectrocloud_ssh_key.ssh_project "<ssh-key-uid>:project"
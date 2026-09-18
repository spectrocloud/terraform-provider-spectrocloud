# Looks up an SSH key asset by name or by ID.
#
# Lookup keys:
#   name    - Optional (conflicts with `id`), also Computed.
#   id      - Optional (conflicts with `name`), also Computed.
#   context - Optional, default "project". Allowed: "project", "tenant".
data "spectrocloud_ssh_key" "example" {
  name = "my-ssh-key"
  # id = "657ec9a27afca71b0dc98027"
  context = "project"
}

# Computed outputs:
#   ssh_key_value - Sensitive. Public key that was uploaded to Palette.
#   ssh_key_id
output "ssh_key_value" {
  value     = data.spectrocloud_ssh_key.example.ssh_key
  sensitive = true
}

output "ssh_key_id" {
  value = data.spectrocloud_ssh_key.example.id
}

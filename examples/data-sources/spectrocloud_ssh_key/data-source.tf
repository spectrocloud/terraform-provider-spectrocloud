# Looks up an SSH key asset by name or by ID.
data "spectrocloud_ssh_key" "example" {
  # Lookup key, optional (conflicts with `id`), also Computed.
  name = "my-ssh-key"
  # Lookup key, optional (conflicts with `name`), also Computed.
  # id = "657ec9a27afca71b0dc98027"
  # Optional lookup key, default "project". Allowed: "project", "tenant".
  context = "project"
}

# Computed, sensitive. Public key that was uploaded to Palette.
output "ssh_key_value" {
  value     = data.spectrocloud_ssh_key.example.ssh_key
  sensitive = true
}

# Computed.
output "ssh_key_id" {
  value = data.spectrocloud_ssh_key.example.id
}

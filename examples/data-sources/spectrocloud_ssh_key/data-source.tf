# Retrieve details of an existing SSH key
data "spectrocloud_ssh_key" "example" {
  name = "my-ssh-key"  # Specify the name of the SSH key resource
}

# Output the SSH key (sensitive)
output "ssh_key_value" {
  value     = data.spectrocloud_ssh_key.example.ssh_key
  sensitive = true
}

# Output the SSH key ID
output "ssh_key_id" {
  value = data.spectrocloud_ssh_key.example.id
}
data "spectrocloud_ssh_key" "key1"{
  name = "test-ssh-key-tf-123-edited"
  context = "project"
}

output "ssh_key" {
  value = base64decode(data.spectrocloud_ssh_key.key1.ssh_key)
  sensitive = true
}
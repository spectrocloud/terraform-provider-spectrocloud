# Looks up a workspace by name and resolves its ID.
data "spectrocloud_workspace" "example_workspace" {
  # Required lookup key.
  name = "my-workspace"
}

# Computed.
output "workspace_id" {
  value = data.spectrocloud_workspace.example_workspace.id
}

output "workspace_name" {
  value = data.spectrocloud_workspace.example_workspace.name
}

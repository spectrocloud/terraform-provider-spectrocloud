# Looks up a workspace by name and resolves its ID.
#
# Lookup keys:
#   name - Required.
data "spectrocloud_workspace" "example_workspace" {
  name = "my-workspace"
}

# Computed outputs:
#   workspace_id
#   workspace_name
output "workspace_id" {
  value = data.spectrocloud_workspace.example_workspace.id
}

output "workspace_name" {
  value = data.spectrocloud_workspace.example_workspace.name
}

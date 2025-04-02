# Retrieve details of a specific workspace
data "spectrocloud_workspace" "example_workspace" {
  name = "my-workspace" # Specify the name of the workspace
}

# Output the retrieved workspace id
output "workspace_name" {
  value = data.spectrocloud_workspace.example_workspace.id
}

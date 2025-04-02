# Fetch details of a specific project in SpectroCloud
data "spectrocloud_project" "example" {
  # Provide either `id` or `name`, but not both.
  id = "project-12345"
  # name = "MyProject"  # Alternative way to reference a project by name
}

# Output project details for reference
output "project_info" {
  value = {
    id   = data.spectrocloud_project.example.id
    name = data.spectrocloud_project.example.name
  }
}

# Looks up a Spectro Cloud project by name and resolves its ID.
data "spectrocloud_project" "example" {
  # Lookup key, optional, also Computed. Only `name`-based lookup is actually implemented by
  # the provider today - the schema also exposes `id` (ConflictsWith `name`), but setting `id`
  # alone does not resolve anything; use `name` as shown here.
  name = "MyProject"
}

# Output project details for reference
output "project_info" {
  value = {
    id   = data.spectrocloud_project.example.id
    name = data.spectrocloud_project.example.name
  }
}

# Looks up a Spectro Cloud project by name and resolves its ID.
data "spectrocloud_project" "example" {
  # Lookup key, optional, also Computed, ConflictsWith `id`. Set either `name` (shown here) or
  # `id` - both are independently implemented lookups (dataSourceProjectRead tries `name` first,
  # then falls back to `id`), so `id` alone works too if you already know the project's UID.
  name = "MyProject"
}

# Equivalent lookup by ID instead of name:
# data "spectrocloud_project" "by_id" {
#   id = "64f1a2b3c4d5e6f7a8b9c0d1"
# }

# Output project details for reference
output "project_info" {
  value = {
    id   = data.spectrocloud_project.example.id
    name = data.spectrocloud_project.example.name
  }
}

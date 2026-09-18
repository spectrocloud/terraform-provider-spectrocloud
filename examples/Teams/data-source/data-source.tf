# Looks up a team by ID or by name.
#
# Lookup keys:
#   id   - Optional (conflicts with `name`), also Computed.
#   name - Optional (conflicts with `id`).
data "spectrocloud_team" "example" {
  id = "team-12345"
  # name = "DevOps Team"
}

# Computed (read-only) outputs:
#   role_ids - Role IDs assigned to this team.
output "team_info" {
  value = {
    id       = data.spectrocloud_team.example.id
    name     = data.spectrocloud_team.example.name
    role_ids = data.spectrocloud_team.example.role_ids
  }
}

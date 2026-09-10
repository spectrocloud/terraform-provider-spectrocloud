# Looks up a team by ID or by name.
data "spectrocloud_team" "example" {
  # Lookup key, optional (conflicts with `name`), also Computed.
  id = "team-12345"
  # Lookup key, optional (conflicts with `id`).
  # name = "DevOps Team"
}

output "team_info" {
  value = {
    id   = data.spectrocloud_team.example.id
    name = data.spectrocloud_team.example.name
    # Computed. Role IDs assigned to this team.
    role_ids = data.spectrocloud_team.example.role_ids
  }
}

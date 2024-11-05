data "spectrocloud_team" "team1" {
  name = "team2"

  # (alternatively)
  # id =  "5fd0ca727c411c71b55a359c"
}

output "team-id" {
  value = data.spectrocloud_team.team1.id
}

output "team-role-ids" {
  value = data.spectrocloud_team.team1.role_ids
}
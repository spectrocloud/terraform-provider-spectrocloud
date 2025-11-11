data "spectrocloud_cluster_config_policy" "policy" {
  name = var.policy_name
}

output "policy_id" {
  value = data.spectrocloud_cluster_config_policy.policy.id
}

output "policy_schedules" {
  value = data.spectrocloud_cluster_config_policy.policy.schedules
}


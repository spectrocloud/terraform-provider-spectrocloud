# Looks up a cluster config (maintenance) policy by name.
#
# Lookup keys:
#   name    - Required.
#   context - Optional, default "project". Allowed: "project", "tenant".
data "spectrocloud_cluster_config_policy" "policy" {
  name    = var.policy_name
  context = var.policy_context
}

# Computed.
output "policy_id" {
  value = data.spectrocloud_cluster_config_policy.policy.id
}

# Computed. List of maintenance schedules, each with name/start_cron/duration_hrs.
output "policy_schedules" {
  value = data.spectrocloud_cluster_config_policy.policy.schedules
}

# Computed.
output "policy_tags" {
  value = data.spectrocloud_cluster_config_policy.policy.tags
}

# Looks up a cluster config template by name.
# Tech Preview: this data source may undergo changes.
data "spectrocloud_cluster_config_template" "template" {
  # Required lookup key.
  name = var.template_name
  # Optional lookup key, default "project". Allowed: "project", "tenant".
  context = var.template_context
}

# Computed.
output "template_id" {
  value = data.spectrocloud_cluster_config_template.template.id
}

output "template_cloud_type" {
  value = data.spectrocloud_cluster_config_template.template.cloud_type
}

output "template_description" {
  value = data.spectrocloud_cluster_config_template.template.description
}

output "template_tags" {
  value = data.spectrocloud_cluster_config_template.template.tags
}

# Computed. Set of {id, variables (name/value/assign_strategy)} per attached cluster profile.
output "template_cluster_profile" {
  value = data.spectrocloud_cluster_config_template.template.cluster_profile
}

# Computed. List of {id, kind} for policies attached to this template.
output "template_policy" {
  value = data.spectrocloud_cluster_config_template.template.policy
}

# Computed. List of {cluster_uid, name} for clusters this template has been applied to.
output "template_attached_clusters" {
  value = data.spectrocloud_cluster_config_template.template.attached_cluster
}

# Computed. One of "Pending", "Applied", "Failed", "PartiallyApplied".
output "template_execution_state" {
  value = data.spectrocloud_cluster_config_template.template.execution_state
}

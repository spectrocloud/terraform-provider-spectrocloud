# Project-level template
data "spectrocloud_cluster_config_template" "template" {
  name    = var.template_name
  context = var.template_context
}

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

output "template_profiles" {
  value = data.spectrocloud_cluster_config_template.template.profiles
}

output "template_policies" {
  value = data.spectrocloud_cluster_config_template.template.policies
}

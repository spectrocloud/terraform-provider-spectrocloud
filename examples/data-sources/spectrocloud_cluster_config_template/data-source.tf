data "spectrocloud_cluster_config_template" "template" {
  name = var.template_name
}

output "template_id" {
  value = data.spectrocloud_cluster_config_template.template.id
}

output "template_cloud_type" {
  value = data.spectrocloud_cluster_config_template.template.cloud_type
}

output "template_profiles" {
  value = data.spectrocloud_cluster_config_template.template.profiles
}

output "template_policies" {
  value = data.spectrocloud_cluster_config_template.template.policies
}


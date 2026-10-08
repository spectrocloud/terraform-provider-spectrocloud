# Retrieve the hardened image rollout status for the whole tenant
data "spectrocloud_hardened_images" "current" {}

# Narrow the report to a single project
data "spectrocloud_hardened_images" "production" {
  project_uid = var.production_project_uid
}

output "hardened_image_rollout_state" {
  value = data.spectrocloud_hardened_images.current.state
}

output "hardened_image_failures" {
  value = data.spectrocloud_hardened_images.current.failed_clusters_count
}

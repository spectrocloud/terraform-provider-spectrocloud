output "cluster_id" {
  value = spectrocloud_cluster_brownfield.cluster.id
}

# Computed. The manifest that must be applied to the real, existing cluster to complete the
# import into Palette.
output "manifest_url" {
  value = spectrocloud_cluster_brownfield.cluster.manifest_url
}

# Computed. Equivalent kubectl command form of the manifest above - also written to disk by
# import_command.tf (see import_command_path below).
output "kubectl_command" {
  value = spectrocloud_cluster_brownfield.cluster.kubectl_command
}

# Computed. Possible values: Pending, Provisioning, Running, Deleting, Deleted, Error, Importing.
output "status" {
  value = spectrocloud_cluster_brownfield.cluster.status
}

# Computed. Possible values: Healthy, UnHealthy, Unknown.
output "health_status" {
  value = spectrocloud_cluster_brownfield.cluster.health_status
}

# Computed. Deprecated in favor of cloud_config, but still populated.
output "cloud_config_id" {
  value = spectrocloud_cluster_brownfield.cluster.cloud_config_id
}

# Local file path written by import_command.tf - run this against the real cluster to complete
# the import:
#   bash $(terraform output -raw import_command_path)
output "import_command_path" {
  value = local_file.import_command.filename
}

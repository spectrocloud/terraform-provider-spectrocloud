output "cluster_id" {
  value = spectrocloud_cluster_import.cluster.id
}

output "cluster_import_manifest_url" {
  value = local.cluster_import_manifest_url
}

output "cluster_import_manifest" {
  value = local.cluster_import_manifest
}
locals {
  cluster_import_manifest_url = spectrocloud_cluster_import.cluster.cluster_import_manifest_apply_command
  cluster_import_manifest     = spectrocloud_cluster_import.cluster.cluster_import_manifest
}

resource "null_resource" "kubectl_apply" {

  provisioner "local-exec" {
    command = spectrocloud_cluster_import.cluster.cluster_import_manifest_url
  }
}
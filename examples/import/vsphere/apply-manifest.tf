resource "null_resource" "kubectl_apply" {
//  triggers = {
//    k8s_yaml_contents = filemd5("/Users/rishi/work/git_clones/terraform-provider-spectrocloud/.local/nginx.yaml")
//  }

  provisioner "local-exec" {
    command = spectrocloud_cluster_import.cluster.cluster_import_manifest_url
  }
}
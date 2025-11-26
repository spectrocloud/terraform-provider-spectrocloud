# Example 1: Basic addon deployment with a single pack from public registry
resource "spectrocloud_addon_deployment" "example_basic" {
  cluster_uid = var.cluster_uid

  cluster_profile {
    id = var.cluster_profile_uid
    variables = {
      test = "IfNotPresent"
    }
    pack {
      name = "nginx-ingress"
      tag  = "4.7.1"
      uid  = data.spectrocloud_pack.nginx.id
      values = <<-EOT
        controller:
          service:
            type: LoadBalancer
      EOT
    }
  }
}

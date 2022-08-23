resource "spectrocloud_cluster_profile" "profile_resource" {
  cloud       = "eks"
  description = "addon-profile-1"
  name        = "addon-profile-1"
  type        = "add-on"

  pack {
    name         = "kubevious-test"
    registry_uid = data.spectrocloud_registry_oci.registry1.id
    tag          = "0.8.15"
    type         = "helm"
    uid          = ""
    values       = <<-EOT
                   pack:
                     namespace: "helm-test-chart"
                     spectrocloud.com/install-priority: "230"
                     releaseNameOverride:
                       test-chart-service: test-chart-service-name
               EOT
  }
}

data "spectrocloud_pack" "dashboard" {
  registry_uid = data.spectrocloud_registry.registry.id
  name         = "spectro-k8s-dashboard"
  version      = "2.7.1"
}

resource "spectrocloud_cluster_profile" "profile-addon" {
  name        = "edge-profile-addon-tf"
  description = "addon cp"
  tags        = ["dev", "department:devops", "owner:alice"]
  cloud       = "edge-native"
  type        = "add-on"

  pack {
    name   = data.spectrocloud_pack.dashboard.name
    tag    = data.spectrocloud_pack.dashboard.version
    uid    = data.spectrocloud_pack.dashboard.id
    values = data.spectrocloud_pack.dashboard.values
  }
}

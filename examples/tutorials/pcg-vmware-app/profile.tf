resource "spectrocloud_cluster_profile" "profile" {
  name        = var.cluster_profile_name
  description = var.cluster_profile_description
  tags        = var.tags
  cloud       = "vsphere"
  type        = "cluster"

  ############################
  # Core layers
  ############################
  pack {
    name   = data.spectrocloud_pack.ubuntu.name
    tag    = data.spectrocloud_pack.ubuntu.version
    uid    = data.spectrocloud_pack.ubuntu.id
    values = data.spectrocloud_pack.ubuntu.values
  }

  pack {
    name   = data.spectrocloud_pack.k8s.name
    tag    = data.spectrocloud_pack.k8s.version
    uid    = data.spectrocloud_pack.k8s.id
    values = data.spectrocloud_pack.k8s.values
  }

  pack {
    name   = data.spectrocloud_pack.cni.name
    tag    = data.spectrocloud_pack.cni.version
    uid    = data.spectrocloud_pack.cni.id
    values = data.spectrocloud_pack.cni.values
  }

  pack {
    name   = data.spectrocloud_pack.csi.name
    tag    = data.spectrocloud_pack.csi.version
    uid    = data.spectrocloud_pack.csi.id
    values = data.spectrocloud_pack.csi.values
  }

  pack {
    name   = data.spectrocloud_pack.metallb.name
    tag    = data.spectrocloud_pack.metallb.version
    uid    = data.spectrocloud_pack.metallb.id
    values = replace(data.spectrocloud_pack.metallb.values, "192.168.10.0/24", var.metallb_ip)
  }

  ############################
  # Add-on layer
  ############################

  pack {
    name   = data.spectrocloud_pack.hellouniverse.name
    tag    = data.spectrocloud_pack.hellouniverse.version
    uid    = data.spectrocloud_pack.hellouniverse.id
    values = data.spectrocloud_pack.hellouniverse.values
    type   = "oci"
  }
}

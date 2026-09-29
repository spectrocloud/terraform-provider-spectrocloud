resource "spectrocloud_cluster_profile" "profile" {
  name        = var.cluster_profile_name
  description = var.cluster_profile_description
  tags        = var.tags
  cloud       = "aws"     # Possible values: "aws", "azure", "gcp"
  type        = "cluster" # Possible values: "cluster", "add-on"

  ############################
  # Core layers
  ############################
  pack {
    name   = "ubuntu-aws"
    tag    = "22.04"
    uid    = data.spectrocloud_pack.ubuntu.id
    values = data.spectrocloud_pack.ubuntu.values
  }

  pack {
    name   = "kubernetes"
    tag    = "1.32.3"
    uid    = data.spectrocloud_pack.k8s.id
    values = data.spectrocloud_pack.k8s.values
  }

  pack {
    name   = "cni-calico"
    tag    = "3.29.3"
    uid    = data.spectrocloud_pack.cni.id
    values = data.spectrocloud_pack.cni.values
  }

  pack {
    name   = "csi-aws-ebs"
    tag    = "1.41.0"
    uid    = data.spectrocloud_pack.csi.id
    values = data.spectrocloud_pack.csi.values
  }

  ############################
  # Add-on layer: the custom pack created in the tutorial
  ############################
  pack {
    name   = "hellouniverse"
    tag    = "1.2.0"
    uid    = data.spectrocloud_pack.hellouniverse.id
    values = data.spectrocloud_pack.hellouniverse.values
  }
}

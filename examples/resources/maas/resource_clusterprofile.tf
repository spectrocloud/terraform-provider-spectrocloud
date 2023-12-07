data "spectrocloud_registry_pack" "registry" {
  name = "Public Repo"

}

data "spectrocloud_pack" "csi" {
  name = "csi-maas-volume"
  #registry_uid = "5e2031962f090e2d3d8a3290"
  version = "1.0.0"
  registry_uid = data.spectrocloud_registry_pack.registry.id
}

data "spectrocloud_pack" "cni" {
  name = "cni-calico"
  #registry_uid = "5e2031962f090e2d3d8a3290"
  version = "3.26.0"
  registry_uid = data.spectrocloud_registry_pack.registry.id
}

data "spectrocloud_pack" "k8s" {
  name = "kubernetes"
  #registry_uid = "5e2031962f090e2d3d8a3290"
  version = "1.28.2"
  registry_uid = data.spectrocloud_registry_pack.registry.id
}

data "spectrocloud_pack" "ubuntu" {
  name = "ubuntu-maas"
  #registry_uid = "5e2031962f090e2d3d8a3290"
  version  = "22.04"
  registry_uid = data.spectrocloud_registry_pack.registry.id
}

resource "spectrocloud_cluster_profile" "profile" {
  name        = "maas-picard-cp-1"
  description = "basic cp"
  cloud       = "maas"
  type        = "cluster"
  context    = "project"


  pack {
    name   = data.spectrocloud_pack.csi.name
    tag    = data.spectrocloud_pack.csi.version
    uid    = data.spectrocloud_pack.csi.id
    values = data.spectrocloud_pack.csi.values
  }

  pack {
    name   = data.spectrocloud_pack.cni.name
    tag    = data.spectrocloud_pack.cni.version
    uid    = data.spectrocloud_pack.cni.id
    values = data.spectrocloud_pack.cni.values
  }

  pack {
    name   = data.spectrocloud_pack.k8s.name
    tag    = data.spectrocloud_pack.k8s.version
    uid    = data.spectrocloud_pack.k8s.id
    values = data.spectrocloud_pack.k8s.values
  }

  pack {
    name   = data.spectrocloud_pack.ubuntu.name
    tag    = data.spectrocloud_pack.ubuntu.version
    uid    = data.spectrocloud_pack.ubuntu.id
    values = data.spectrocloud_pack.ubuntu.values
  }
}

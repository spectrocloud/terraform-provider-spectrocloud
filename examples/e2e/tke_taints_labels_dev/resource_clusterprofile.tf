data "spectrocloud_registry" "registry" {
  name = "Public Repo"
}

data "spectrocloud_pack" "csi" {
  registry_uid = data.spectrocloud_registry.registry.id
  name    = "csi-tke"
  version = "1.0"
}

data "spectrocloud_pack" "cni" {
  registry_uid = data.spectrocloud_registry.registry.id
  name    = "cni-tke-global-router"
  version = "1.0"
}

data "spectrocloud_pack" "k8s" {
  registry_uid = data.spectrocloud_registry.registry.id
  name    = "kubernetes-tke"
  version = "1.20.6"
}

data "spectrocloud_pack" "ubuntu" {
  registry_uid = data.spectrocloud_registry.registry.id
  name    = "ubuntu-tke"
  version = "18.04"
}

resource "spectrocloud_cluster_profile" "profile" {
  name        = "ProdTKE-tf1"
  description = "basic tfe cp"
  cloud       = "tke"
  type        = "cluster"

  pack {
    name   = data.spectrocloud_pack.ubuntu.name
    tag    = "18.04"
    uid    = data.spectrocloud_pack.ubuntu.id
    values = data.spectrocloud_pack.ubuntu.values
  }
  pack {
    name   = data.spectrocloud_pack.k8s.name
    tag    = "1.20.x"
    uid    = data.spectrocloud_pack.k8s.id
    values = data.spectrocloud_pack.k8s.values
  }

  pack {
    name   = data.spectrocloud_pack.cni.name
    tag    = "1.0"
    uid    = data.spectrocloud_pack.cni.id
    values = data.spectrocloud_pack.cni.values
  }

  pack {
    name   = data.spectrocloud_pack.csi.name
    tag    = "1.0"
    uid    = data.spectrocloud_pack.csi.id
    values = data.spectrocloud_pack.csi.values
  }
}


data "spectrocloud_registry" "registry" {
  name = "Public Repo"
}

data "spectrocloud_pack" "csi" {
  name         = "csi-longhorn"
  registry_uid = data.spectrocloud_registry.registry.id
  version      = "1.5.1"
}

data "spectrocloud_pack" "cni" {
  name         = "cni-calico"
  registry_uid = data.spectrocloud_registry.registry.id
  version      = "3.26.1"
}

data "spectrocloud_pack" "k8s" {
  name         = "kubernetes-coxedge"
  registry_uid = data.spectrocloud_registry.registry.id
  version      = "1.27.2"
}

data "spectrocloud_pack" "ubuntu" {
  name         = "ubuntu-coxedge"
  registry_uid = data.spectrocloud_registry.registry.id
  version      = "20.04"
}

resource "spectrocloud_cluster_profile" "profile" {
  name        = "coxedge-profile-tf"
  description = "basic cp"
  cloud       = "coxedge"
  type        = "cluster"

  pack {
    name   = "ubuntu-coxedge"
    tag    = data.spectrocloud_pack.ubuntu.version
    uid    = data.spectrocloud_pack.ubuntu.id
    values = data.spectrocloud_pack.ubuntu.values
  }

  pack {
    name   = "kubernetes-coxedge"
    tag    = data.spectrocloud_pack.k8s.version
    uid    = data.spectrocloud_pack.k8s.id
    values = data.spectrocloud_pack.k8s.values
  }

  pack {
    name   = "cni-calico"
    tag    = data.spectrocloud_pack.cni.version
    uid    = data.spectrocloud_pack.cni.id
    values = data.spectrocloud_pack.cni.values
  }

  pack {
    name   = "csi-longhorn"
    tag    = data.spectrocloud_pack.csi.version
    uid    = data.spectrocloud_pack.csi.id
    values = data.spectrocloud_pack.csi.values
  }

}

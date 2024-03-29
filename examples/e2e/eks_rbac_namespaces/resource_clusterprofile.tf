data "spectrocloud_pack" "csi" {
  name    = "csi-aws"
  version = "1.0.0"
}

data "spectrocloud_pack" "cni" {
  name    = "cni-aws-vpc-eks"
  version = "1.0"
}

data "spectrocloud_pack" "k8s" {
  name    = "kubernetes-eks"
  version = "1.21"
}

data "spectrocloud_pack" "ubuntu" {
  name    = "amazon-linux-eks"
  version = "1.0.0"
}

resource "spectrocloud_cluster_profile" "profile" {
  name        = "ProdEKS-tf1"
  description = "basic eks cp"
  cloud       = "eks"
  type        = "cluster"

  pack {
    name   = data.spectrocloud_pack.ubuntu.name
    tag    = "1.0.0"
    uid    = data.spectrocloud_pack.ubuntu.id
    values = data.spectrocloud_pack.ubuntu.values
  }
  pack {
    name   = data.spectrocloud_pack.k8s.name
    tag    = "1.21.x"
    uid    = data.spectrocloud_pack.k8s.id
    values = data.spectrocloud_pack.k8s.values
  }

  pack {
    name   = data.spectrocloud_pack.cni.name
    tag    = "1.0.x"
    uid    = data.spectrocloud_pack.cni.id
    values = data.spectrocloud_pack.cni.values
  }

  pack {
    name   = data.spectrocloud_pack.csi.name
    tag    = "1.x"
    uid    = data.spectrocloud_pack.csi.id
    values = data.spectrocloud_pack.csi.values
  }
}


resource "spectrocloud_cluster_profile" "aws_profile" {
  count = var.deploy-aws ? 1 : 0

  name        = "tf-cluster-template-profile-aws"
  description = "Cluster profile for the cluster templates tutorial"
  cloud       = "aws"
  type        = "cluster"
  version     = "1.0.0"

  pack {
    name   = data.spectrocloud_pack.aws_ubuntu.name
    tag    = data.spectrocloud_pack.aws_ubuntu.version
    uid    = data.spectrocloud_pack.aws_ubuntu.id
    values = data.spectrocloud_pack.aws_ubuntu.values
    type   = "spectro"
  }

  pack {
    name   = data.spectrocloud_pack.aws_k8s.name
    tag    = data.spectrocloud_pack.aws_k8s.version
    uid    = data.spectrocloud_pack.aws_k8s.id
    values = data.spectrocloud_pack.aws_k8s.values
    type   = "spectro"
  }

  pack {
    name   = data.spectrocloud_pack.aws_cni.name
    tag    = data.spectrocloud_pack.aws_cni.version
    uid    = data.spectrocloud_pack.aws_cni.id
    values = data.spectrocloud_pack.aws_cni.values
    type   = "spectro"
  }

  pack {
    name   = data.spectrocloud_pack.aws_csi.name
    tag    = data.spectrocloud_pack.aws_csi.version
    uid    = data.spectrocloud_pack.aws_csi.id
    values = data.spectrocloud_pack.aws_csi.values
    type   = "spectro"
  }

  pack {
    name = data.spectrocloud_pack.hellouniverse.name
    tag  = data.spectrocloud_pack.hellouniverse.version
    uid  = data.spectrocloud_pack.hellouniverse.id
    values = templatefile("manifests/values-hello-universe.yaml", {
      port = var.app_port
    })
    type = "oci"
  }

  profile_variables {
    variable {
      name          = "app_replicas"
      display_name  = "Hello Universe Replicas"
      format        = "number"
      description   = "Number of replicas for the hello-universe workload"
      default_value = "1"
      required      = true
    }
  }
}

resource "spectrocloud_cluster_profile" "azure_profile" {
  count = var.deploy-azure ? 1 : 0

  name        = "tf-cluster-template-profile-azure"
  description = "Cluster profile for the cluster templates tutorial"
  cloud       = "azure"
  type        = "cluster"
  version     = "1.0.0"

  pack {
    name   = data.spectrocloud_pack.azure_ubuntu.name
    tag    = data.spectrocloud_pack.azure_ubuntu.version
    uid    = data.spectrocloud_pack.azure_ubuntu.id
    values = data.spectrocloud_pack.azure_ubuntu.values
    type   = "spectro"
  }

  pack {
    name   = data.spectrocloud_pack.azure_k8s.name
    tag    = data.spectrocloud_pack.azure_k8s.version
    uid    = data.spectrocloud_pack.azure_k8s.id
    values = data.spectrocloud_pack.azure_k8s.values
    type   = "spectro"
  }

  pack {
    name   = data.spectrocloud_pack.azure_cni.name
    tag    = data.spectrocloud_pack.azure_cni.version
    uid    = data.spectrocloud_pack.azure_cni.id
    values = data.spectrocloud_pack.azure_cni.values
    type   = "spectro"
  }

  pack {
    name   = data.spectrocloud_pack.azure_csi.name
    tag    = data.spectrocloud_pack.azure_csi.version
    uid    = data.spectrocloud_pack.azure_csi.id
    values = data.spectrocloud_pack.azure_csi.values
    type   = "spectro"
  }

  pack {
    name = data.spectrocloud_pack.hellouniverse.name
    tag  = data.spectrocloud_pack.hellouniverse.version
    uid  = data.spectrocloud_pack.hellouniverse.id
    values = templatefile("manifests/values-hello-universe.yaml", {
      port = var.app_port
    })
    type = "oci"
  }

  profile_variables {
    variable {
      name          = "app_replicas"
      display_name  = "Hello Universe Replicas"
      format        = "number"
      description   = "Number of replicas for the hello-universe workload"
      default_value = "1"
      required      = true
    }
  }
}

resource "spectrocloud_cluster_profile" "aws_profile_v110" {
  count = (var.deploy-aws && var.create_new_profile_version) ? 1 : 0

  name        = "tf-cluster-template-profile-aws"
  description = "Cluster profile for the cluster templates tutorial"
  cloud       = "aws"
  type        = "cluster"
  version     = "1.1.0"

  pack {
    name   = data.spectrocloud_pack.aws_ubuntu.name
    tag    = data.spectrocloud_pack.aws_ubuntu.version
    uid    = data.spectrocloud_pack.aws_ubuntu.id
    values = data.spectrocloud_pack.aws_ubuntu.values
    type   = "spectro"
  }

  pack {
    name   = data.spectrocloud_pack.aws_k8s.name
    tag    = data.spectrocloud_pack.aws_k8s.version
    uid    = data.spectrocloud_pack.aws_k8s.id
    values = data.spectrocloud_pack.aws_k8s.values
    type   = "spectro"
  }

  pack {
    name   = data.spectrocloud_pack.aws_cni.name
    tag    = data.spectrocloud_pack.aws_cni.version
    uid    = data.spectrocloud_pack.aws_cni.id
    values = data.spectrocloud_pack.aws_cni.values
    type   = "spectro"
  }

  pack {
    name   = data.spectrocloud_pack.aws_csi.name
    tag    = data.spectrocloud_pack.aws_csi.version
    uid    = data.spectrocloud_pack.aws_csi.id
    values = data.spectrocloud_pack.aws_csi.values
    type   = "spectro"
  }

  pack {
    name = data.spectrocloud_pack.hellouniverse.name
    tag  = data.spectrocloud_pack.hellouniverse.version
    uid  = data.spectrocloud_pack.hellouniverse.id
    values = templatefile("manifests/values-hello-universe.yaml", {
      port = var.app_port
    })
    type = "oci"
  }

  pack {
    name = data.spectrocloud_pack.kubecost[0].name
    tag  = data.spectrocloud_pack.kubecost[0].version
    uid  = data.spectrocloud_pack.kubecost[0].id
    type = "oci"
  }

  profile_variables {
    variable {
      name          = "app_replicas"
      display_name  = "Hello Universe Replicas"
      format        = "number"
      description   = "Number of replicas for the hello-universe workload"
      default_value = "1"
      required      = true
    }
  }
}

resource "spectrocloud_cluster_profile" "azure_profile_v110" {
  count = (var.deploy-azure && var.create_new_profile_version) ? 1 : 0

  name        = "tf-cluster-template-profile-azure"
  description = "Cluster profile for the cluster templates tutorial"
  cloud       = "azure"
  type        = "cluster"
  version     = "1.1.0"

  pack {
    name   = data.spectrocloud_pack.azure_ubuntu.name
    tag    = data.spectrocloud_pack.azure_ubuntu.version
    uid    = data.spectrocloud_pack.azure_ubuntu.id
    values = data.spectrocloud_pack.azure_ubuntu.values
    type   = "spectro"
  }

  pack {
    name   = data.spectrocloud_pack.azure_k8s.name
    tag    = data.spectrocloud_pack.azure_k8s.version
    uid    = data.spectrocloud_pack.azure_k8s.id
    values = data.spectrocloud_pack.azure_k8s.values
    type   = "spectro"
  }

  pack {
    name   = data.spectrocloud_pack.azure_cni.name
    tag    = data.spectrocloud_pack.azure_cni.version
    uid    = data.spectrocloud_pack.azure_cni.id
    values = data.spectrocloud_pack.azure_cni.values
    type   = "spectro"
  }

  pack {
    name   = data.spectrocloud_pack.azure_csi.name
    tag    = data.spectrocloud_pack.azure_csi.version
    uid    = data.spectrocloud_pack.azure_csi.id
    values = data.spectrocloud_pack.azure_csi.values
    type   = "spectro"
  }

  pack {
    name = data.spectrocloud_pack.hellouniverse.name
    tag  = data.spectrocloud_pack.hellouniverse.version
    uid  = data.spectrocloud_pack.hellouniverse.id
    values = templatefile("manifests/values-hello-universe.yaml", {
      port = var.app_port
    })
    type = "oci"
  }

  pack {
    name = data.spectrocloud_pack.kubecost[0].name
    tag  = data.spectrocloud_pack.kubecost[0].version
    uid  = data.spectrocloud_pack.kubecost[0].id
    type = "oci"
  }

  profile_variables {
    variable {
      name          = "app_replicas"
      display_name  = "Hello Universe Replicas"
      format        = "number"
      description   = "Number of replicas for the hello-universe workload"
      default_value = "1"
      required      = true
    }
  }
}

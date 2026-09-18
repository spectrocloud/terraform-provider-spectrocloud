#############################
# AWS Cluster Profile v1.0.0
#############################
resource "spectrocloud_cluster_profile" "aws-profile" {
  count = var.deploy-aws ? 1 : 0

  name        = "aws-profile-var-tf"
  description = "A basic cluster profile for AWS"
  tags        = concat(var.tags, ["env:aws"])
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
    name   = data.spectrocloud_pack.wordpress_chart.name
    tag    = data.spectrocloud_pack.wordpress_chart.version
    uid    = data.spectrocloud_pack.wordpress_chart.id
    values = file("manifests/wordpress-default.yaml")
    type   = "oci"
  }
}

############################
# AWS Cluster Profile v1.1.0
############################

resource "spectrocloud_cluster_profile" "aws-profile-var" {
  count = var.deploy-aws ? 1 : 0

  name        = "aws-profile-var-tf"
  description = "Updated version with new profile variable values"
  tags        = concat(var.tags, ["env:aws"])
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
    name   = data.spectrocloud_pack.wordpress_chart.name
    tag    = data.spectrocloud_pack.wordpress_chart.version
    uid    = data.spectrocloud_pack.wordpress_chart.id
    values = file("manifests/wordpress-variables.yaml")
    type   = "oci"
  }
  profile_variables {
    variable {
      name          = "wordpress_replica"
      display_name  = "Number of replicas"
      format        = "number"
      description   = "Number of replicas to deploy for Wordpress"
      default_value = var.wordpress_replica
      required      = true
    }
    variable {
      name          = "wordpress_namespace"
      display_name  = "WordPress: Namespace"
      format        = "string"
      description   = "Namespace for the Wordpress pack"
      default_value = var.wordpress_namespace
      required      = true
    }
    variable {
      name          = "wordpress_port"
      display_name  = "WordPress: Port"
      format        = "number"
      description   = "Port for Wordpress HTTP"
      default_value = var.wordpress_port
      required      = true
    }
  }
}

##############################
# Azure Cluster Profile v1.0.0
##############################
resource "spectrocloud_cluster_profile" "azure-profile" {
  count = var.deploy-azure ? 1 : 0

  name        = "azure-profile-variables-tf"
  description = "A basic cluster profile for Azure"
  tags        = concat(var.tags, ["env:azure"])
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
    name   = data.spectrocloud_pack.wordpress_chart.name
    tag    = data.spectrocloud_pack.wordpress_chart.version
    uid    = data.spectrocloud_pack.wordpress_chart.id
    values = file("manifests/wordpress-default.yaml")
    type   = "oci"
  }
}

##############################
# Azure Cluster Profile v1.1.0
##############################
resource "spectrocloud_cluster_profile" "azure-profile-var" {
  count = var.deploy-azure ? 1 : 0

  name        = "azure-profile-variables-tf"
  description = "A basic cluster profile for Azure with cluster profile variables"
  tags        = concat(var.tags, ["env:azure"])
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
    name   = data.spectrocloud_pack.wordpress_chart.name
    tag    = data.spectrocloud_pack.wordpress_chart.version
    uid    = data.spectrocloud_pack.wordpress_chart.id
    values = file("manifests/wordpress-variables.yaml")
    type   = "oci"
  }
  profile_variables {
    variable {
      name          = "wordpress_replica"
      display_name  = "Number of replicas"
      format        = "number"
      description   = "Number of replicas to deploy for Wordpress"
      default_value = var.wordpress_replica
      required      = true
    }
    variable {
      name          = "wordpress_namespace"
      display_name  = "WordPress: Namespace"
      format        = "string"
      description   = "Namespace for the Wordpress pack"
      default_value = var.wordpress_namespace
      required      = true
    }
    variable {
      name          = "wordpress_port"
      display_name  = "WordPress: Port"
      format        = "number"
      description   = "Port for Wordpress HTTP"
      default_value = var.wordpress_port
      required      = true
    }
  }
}


############################
# GCP Cluster Profile v1.0.0
############################
resource "spectrocloud_cluster_profile" "gcp-profile" {
  count = var.deploy-gcp ? 1 : 0

  name        = "gcp-profile-variables-tf"
  description = "A basic cluster profile for GCP"
  tags        = concat(var.tags, ["env:GCP"])
  cloud       = "gcp"
  type        = "cluster"
  version     = "1.0.0"

  pack {
    name   = data.spectrocloud_pack.gcp_ubuntu.name
    tag    = data.spectrocloud_pack.gcp_ubuntu.version
    uid    = data.spectrocloud_pack.gcp_ubuntu.id
    values = data.spectrocloud_pack.gcp_ubuntu.values
    type   = "spectro"
  }

  pack {
    name   = data.spectrocloud_pack.gcp_k8s.name
    tag    = data.spectrocloud_pack.gcp_k8s.version
    uid    = data.spectrocloud_pack.gcp_k8s.id
    values = data.spectrocloud_pack.gcp_k8s.values
    type   = "spectro"
  }

  pack {
    name   = data.spectrocloud_pack.gcp_cni.name
    tag    = data.spectrocloud_pack.gcp_cni.version
    uid    = data.spectrocloud_pack.gcp_cni.id
    values = data.spectrocloud_pack.gcp_cni.values
    type   = "spectro"
  }

  pack {
    name   = data.spectrocloud_pack.gcp_csi.name
    tag    = data.spectrocloud_pack.gcp_csi.version
    uid    = data.spectrocloud_pack.gcp_csi.id
    values = data.spectrocloud_pack.gcp_csi.values
    type   = "spectro"
  }

  pack {
    name   = data.spectrocloud_pack.wordpress_chart.name
    tag    = data.spectrocloud_pack.wordpress_chart.version
    uid    = data.spectrocloud_pack.wordpress_chart.id
    values = file("manifests/wordpress-default.yaml")
    type   = "oci"
  }
}

############################
# GCP Cluster Profile v1.1.0
############################
resource "spectrocloud_cluster_profile" "gcp-profile-var" {
  count = var.deploy-gcp ? 1 : 0

  name        = "gcp-profile-variables-tf"
  description = "A basic cluster profile for GCP with cluster profile variables"
  tags        = concat(var.tags, ["env:GCP"])
  cloud       = "gcp"
  type        = "cluster"
  version     = "1.1.0"

  pack {
    name   = data.spectrocloud_pack.gcp_ubuntu.name
    tag    = data.spectrocloud_pack.gcp_ubuntu.version
    uid    = data.spectrocloud_pack.gcp_ubuntu.id
    values = data.spectrocloud_pack.gcp_ubuntu.values
    type   = "spectro"
  }

  pack {
    name   = data.spectrocloud_pack.gcp_k8s.name
    tag    = data.spectrocloud_pack.gcp_k8s.version
    uid    = data.spectrocloud_pack.gcp_k8s.id
    values = data.spectrocloud_pack.gcp_k8s.values
    type   = "spectro"
  }

  pack {
    name   = data.spectrocloud_pack.gcp_cni.name
    tag    = data.spectrocloud_pack.gcp_cni.version
    uid    = data.spectrocloud_pack.gcp_cni.id
    values = data.spectrocloud_pack.gcp_cni.values
    type   = "spectro"
  }

  pack {
    name   = data.spectrocloud_pack.gcp_csi.name
    tag    = data.spectrocloud_pack.gcp_csi.version
    uid    = data.spectrocloud_pack.gcp_csi.id
    values = data.spectrocloud_pack.gcp_csi.values
    type   = "spectro"
  }

  pack {
    name   = data.spectrocloud_pack.wordpress_chart.name
    tag    = data.spectrocloud_pack.wordpress_chart.version
    uid    = data.spectrocloud_pack.wordpress_chart.id
    values = file("manifests/wordpress-variables.yaml")
    type   = "oci"
  }
  profile_variables {
    variable {
      name          = "wordpress_replica"
      display_name  = "Number of replicas"
      format        = "number"
      description   = "Number of replicas to deploy for Wordpress"
      default_value = var.wordpress_replica
      required      = true
    }
    variable {
      name          = "wordpress_namespace"
      display_name  = "WordPress: Namespace"
      format        = "string"
      description   = "Namespace for the Wordpress pack"
      default_value = var.wordpress_namespace
      required      = true
    }
    variable {
      name          = "wordpress_port"
      display_name  = "WordPress: Port"
      format        = "number"
      description   = "Port for Wordpress HTTP"
      default_value = var.wordpress_port
      required      = true
    }
  }
}

# EKS uses its own OS pack (AMI-based, since node groups are managed by AWS/Palette
# differently than self-hosted clouds) plus its own Kubernetes and CNI packs.
data "spectrocloud_pack" "ami" {
  name = "amazon-linux-eks"
}

data "spectrocloud_pack" "k8s" {
  name = "kubernetes-eks"
}

data "spectrocloud_pack" "cni" {
  name = "cni-aws-vpc-eks"
}

data "spectrocloud_pack" "csi" {
  name = "csi-aws"
}

locals {
  byo_manifest_values = <<-EOT
    manifests:
      byo-manifest:
        contents: |
          apiVersion: v1
          kind: Namespace
          metadata:
            name: e2e-demo-ns
  EOT
}

# A full-stack cluster profile: AMI + Kubernetes + CNI + CSI layers, plus one add-on pack. Day-2
# mutability: on spectrocloud_cluster_profile, `cloud` and `type` are ForceNew; `pack` and
# `profile_variables` update in place (`version` clones a new profile version on change - see
# examples/resources/spectrocloud_cluster_profile for the full mutability breakdown).
resource "spectrocloud_cluster_profile" "eks_profile" {
  name        = "e2e-eks-full-stack"
  description = "Comprehensive EKS cluster profile for the end-to-end usecase example"
  cloud       = "eks"
  type        = "cluster"
  context     = "project"
  tags        = ["e2e-usecase", "team:platform"]

  pack {
    name   = "amazon-linux-eks"
    tag    = data.spectrocloud_pack.ami.version
    uid    = data.spectrocloud_pack.ami.id
    values = data.spectrocloud_pack.ami.values
  }

  pack {
    name   = "kubernetes-eks"
    tag    = data.spectrocloud_pack.k8s.version
    uid    = data.spectrocloud_pack.k8s.id
    values = data.spectrocloud_pack.k8s.values
  }

  pack {
    name   = "cni-aws-vpc-eks"
    tag    = data.spectrocloud_pack.cni.version
    uid    = data.spectrocloud_pack.cni.id
    values = data.spectrocloud_pack.cni.values
  }

  pack {
    name   = "csi-aws"
    tag    = data.spectrocloud_pack.csi.version
    uid    = data.spectrocloud_pack.csi.id
    values = data.spectrocloud_pack.csi.values
  }

  # Add-on manifest pack, templated with a profile_variables value at apply time.
  pack {
    name = "byo-manifest-addon"
    type = "manifest"
    manifest {
      name    = "byo-manifest"
      content = local.byo_manifest_values
    }
  }

  # profile_variables lets values be supplied per-cluster (via cluster_profile.variables on the
  # cluster resource) without editing the profile itself. Exactly one profile_variables block is
  # allowed; all variables live in its single `variable` list. This demonstrates all 4 formats.
  profile_variables {
    variable {
      name         = "default_password"
      display_name = "Default Password"
      format       = "string"
      hidden       = true
    }
    variable {
      name          = "app_version"
      display_name  = "Application Version"
      format        = "version"
      description   = "Semantic version to roll out to the cluster's add-ons."
      default_value = "1.0.0"
      regex         = "^\\d+\\.\\d+\\.\\d+$"
      required      = true
      immutable     = false
    }
    variable {
      name          = "environment_tier"
      display_name  = "Environment Tier"
      format        = "string"
      input_type    = "dropdown"
      default_value = "staging"
      required      = false
      options {
        label       = "staging"
        value       = "staging"
        description = "Pre-production environment"
      }
      options {
        label       = "production"
        value       = "production"
        description = "Production environment"
      }
    }
    variable {
      name          = "notes"
      display_name  = "Deployment Notes"
      format        = "string"
      input_type    = "multiline"
      required      = false
      default_value = <<-EOT
      Provisioned via examples/end-to-end-usecases.
      Demonstrates the full spectrocloud_cluster_eks attribute surface.
      EOT
    }
  }
}

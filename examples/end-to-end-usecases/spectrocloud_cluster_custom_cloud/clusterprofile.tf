data "spectrocloud_pack" "k8s" {
  name    = "kubernetes"
  version = "1.28.5"
}

data "spectrocloud_pack" "cni" {
  name    = "cni-calico"
  version = "3.27.0"
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

# Custom cloud profiles skip the OS and CSI layers other cloud types need - node provisioning
# (image, disk, network) for a custom cloud is described directly in the raw Cluster API YAML
# (cloud_config.values / machine_pool.node_pool_config on the cluster resource, see cluster.tf),
# not through Palette-managed OS/CSI packs.
#
# Day-2 mutability: on spectrocloud_cluster_profile, `cloud` and `type` are ForceNew; `pack` and
# `profile_variables` update in place (`version` clones a new profile version on change - see
# examples/resources/spectrocloud_cluster_profile for the full mutability breakdown).
resource "spectrocloud_cluster_profile" "custom_cloud_profile" {
  name        = "e2e-custom-cloud-full-stack"
  description = "Comprehensive custom cloud (Nutanix) cluster profile for the end-to-end usecase example"
  cloud       = "nutanix"
  type        = "cluster"
  context     = "project"
  tags        = ["e2e-usecase", "team:platform"]

  pack {
    name   = "kubernetes"
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
      Demonstrates the full spectrocloud_cluster_custom_cloud attribute surface.
      EOT
    }
  }
}

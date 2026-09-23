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

# Virtual clusters get their Kubernetes control plane from the vcluster Helm chart configured in
# cloud_config on the cluster resource (see cluster.tf), not from an OS/kubernetes/cni/csi pack
# stack - so, matching the pattern used by examples/e2e/virtual_cluster, this is an `add-on`-type
# profile (cloud = "all") carrying only workload/add-on manifests.
#
# Day-2 mutability: on spectrocloud_cluster_profile, `cloud` and `type` are ForceNew; `pack` and
# `profile_variables` update in place (`version` clones a new profile version on change - see
# examples/resources/spectrocloud_cluster_profile for the full mutability breakdown).
resource "spectrocloud_cluster_profile" "virtual_cluster_profile" {
  name        = "e2e-virtual-cluster-addons"
  description = "Comprehensive add-on profile for the virtual cluster end-to-end usecase example"
  cloud       = "all"
  type        = "add-on"
  context     = "project"
  tags        = ["e2e-usecase", "team:platform"]

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
      Demonstrates the full spectrocloud_virtual_cluster attribute surface.
      EOT
    }
  }
}

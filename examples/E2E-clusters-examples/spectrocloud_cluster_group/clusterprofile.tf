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

# A cluster group's own `cluster_profile` attribute (see cluster_group.tf) applies an add-on
# profile across every host cluster in the group at once, rather than per virtual cluster - so,
# matching the pattern used by examples/E2E-clusters-examples/spectrocloud_virtual_cluster, this
# is an `add-on`-type profile (cloud = "all") carrying only workload/add-on manifests.
#
# Day-2 mutability: on spectrocloud_cluster_profile, `cloud` and `type` are ForceNew; `pack` and
# `profile_variables` update in place (`version` clones a new profile version on change - see
# examples/resources/spectrocloud_cluster_profile for the full mutability breakdown).
resource "spectrocloud_cluster_profile" "host_group_addon_profile" {
  name        = "e2e-cluster-group-addons"
  description = "Comprehensive add-on profile for the cluster group end-to-end usecase example"
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

  # profile_variables lets values be supplied per-group (via cluster_profile.variables on the
  # cluster_group resource) without editing the profile itself. Exactly one profile_variables
  # block is allowed; all variables live in its single `variable` list. This demonstrates all 4
  # formats.
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
      description   = "Semantic version to roll out to every host cluster's add-ons."
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
      Provisioned via examples/E2E-clusters-examples.
      Demonstrates the full spectrocloud_cluster_group attribute surface.
      EOT
    }
  }
}

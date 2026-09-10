data "spectrocloud_pack" "ubuntu" {
  name    = "ubuntu-vsphere"
  version = "18.04"
}

data "spectrocloud_pack" "k8s" {
  name    = "kubernetes"
  version = "1.28.5"
}

data "spectrocloud_pack" "cni" {
  name    = "cni-calico"
  version = "3.27.0"
}

data "spectrocloud_pack" "csi" {
  name    = "csi-vsphere-csi"
  version = "3.1.0"
}

locals {
  proxy_addon_values = <<-EOT
    manifests:
      spectro-proxy:
        namespace: "cluster-{{ .spectro.system.cluster.uid }}"
        server: "{{ .spectro.system.reverseproxy.server }}"
        clusterUid: "{{ .spectro.system.cluster.uid }}"
        subdomain: "cluster-{{ .spectro.system.cluster.uid }}"
  EOT
}

# A full-stack cluster profile: OS + Kubernetes + CNI + CSI layers, plus one add-on pack. Day-2
# mutability: on spectrocloud_cluster_profile, `cloud` and `type` are ForceNew; `pack` and
# `profile_variables` update in place (`version` clones a new profile version on change - see
# examples/resources/spectrocloud_cluster_profile for the full mutability breakdown).
resource "spectrocloud_cluster_profile" "vsphere_profile" {
  # Required, ForceNew.
  name = "e2e-vsphere-full-stack"
  # Optional. Free-text description shown in the Palette UI.
  description = "Comprehensive vSphere cluster profile for the end-to-end usecase example"
  # Required, ForceNew. Must match the target cluster's cloud - "vsphere" here.
  cloud = "vsphere"
  # Required, ForceNew. "cluster" (full-stack, used here) or "add-on"/"system" for
  # layer-specific profiles attached alongside a full-stack one.
  type = "cluster"
  # Optional, default "project". Allowed: "project", "tenant", "system".
  context = "project"
  # Optional. Tags in `key:value` form.
  tags = ["e2e-usecase", "team:platform"]

  # Optional, default false. Advanced Day-2 pattern (requires the `immutable-clusterprofiles`
  # feature_preview flag): when true, `terraform destroy` only removes this profile from
  # Terraform state and leaves the underlying version intact in Palette. Combined with bumping
  # `version` and `lifecycle { create_before_destroy = true }`, this lets you version profiles
  # in HCL while every previous version stays preserved and usable in Palette. Not used in this
  # example (left at its default), but documented here since it's part of the full schema.
  # skip_destroy = true

  pack {
    name   = "ubuntu-vsphere"
    tag    = data.spectrocloud_pack.ubuntu.version
    uid    = data.spectrocloud_pack.ubuntu.id
    values = data.spectrocloud_pack.ubuntu.values
  }

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

  pack {
    name   = "csi-vsphere-csi"
    tag    = data.spectrocloud_pack.csi.version
    uid    = data.spectrocloud_pack.csi.id
    values = data.spectrocloud_pack.csi.values
  }

  # Add-on manifest pack, templated with a profile_variables value at apply time.
  pack {
    name = "spectro-proxy-addon"
    type = "manifest"
    manifest {
      name    = "spectro-proxy-manifest"
      content = local.proxy_addon_values
    }
  }

  # profile_variables lets values be supplied per-cluster (via cluster_profile.variables on the
  # cluster resource) without editing the profile itself. Exactly one profile_variables block is
  # allowed; all variables live in its single `variable` list. This demonstrates all 4 formats.
  profile_variables {
    # string, hidden (e.g. for secrets)
    variable {
      name         = "default_password"
      display_name = "Default Password"
      format       = "string"
      hidden       = true
    }
    # version, with a regex constraint
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
    # string with a dropdown of allowed values
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
    # multiline string
    variable {
      name          = "notes"
      display_name  = "Deployment Notes"
      format        = "string"
      input_type    = "multiline"
      required      = false
      default_value = <<-EOT
      Provisioned via examples/end-to-end-usecases.
      Demonstrates the full spectrocloud_cluster_vsphere attribute surface.
      EOT
    }
  }
}

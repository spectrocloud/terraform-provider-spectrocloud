# This example attaches an existing cluster profile pack (nginx-ingress from the public registry)
# to an already-running cluster as an addon deployment - a way to layer additional packs onto a
# cluster after it has been created, without including them in the cluster's original profile.

data "spectrocloud_registry" "public_registry" {
  name = "Public Repo"
}

data "spectrocloud_pack" "nginx_ingress" {
  name         = "nginx-ingress"
  version      = "4.7.1"
  registry_uid = data.spectrocloud_registry.public_registry.id
}

resource "spectrocloud_addon_deployment" "example" {
  # Required. The UID of the already-existing cluster to attach this addon profile to.
  # No attribute on this resource is ForceNew - the whole resource updates in place.
  cluster_uid = var.cluster_uid

  # Optional. Scope of the cluster lookup: "project" (default) or "tenant".
  context = "project"

  # Optional, default "DownloadAndInstall". How Palette applies the profile once attached.
  # Allowed values: "DownloadAndInstall" (download and install in one action) or
  # "DownloadAndInstallLater" (download the artifact now, install later).
  apply_setting = "DownloadAndInstall"

  # Exactly one cluster_profile block is allowed per addon deployment - use a separate
  # spectrocloud_addon_deployment resource for each additional profile you want to attach.
  cluster_profile {
    # Required. The UID of the cluster profile being attached as an addon.
    id = var.cluster_profile_uid

    # Optional. Profile variable overrides, only if the profile defines a profile_variables
    # block (see the spectrocloud_cluster_profile resource). Keys and values are both strings.
    variables = {
      # replica_count = "2"
    }

    pack {
      # Required, must be unique within this cluster profile.
      name = data.spectrocloud_pack.nginx_ingress.name
      # Required in practice for "spectro"/"helm" packs - the pack version.
      tag = data.spectrocloud_pack.nginx_ingress.version
      # Optional/Computed. Resolved here via the data source above. If omitted, name + tag +
      # registry_uid (or registry_name) are used to resolve the pack UID internally instead.
      uid = data.spectrocloud_pack.nginx_ingress.id
      # Optional, default "spectro". Set to "oci" for OCI-hosted packs, "helm" for Helm charts,
      # or "manifest" for raw Kubernetes manifests (see the alternative example below).
      type = "spectro"
      # Optional. Pack configuration values in YAML, same shape as the pack's presets in the
      # Palette UI.
      values = <<-EOT
        controller:
          service:
            type: LoadBalancer
      EOT
    }
  }
}

# ---------------------------------------------------------------------------------------------
# Alternative: attach a raw Kubernetes manifest instead of a registry pack. Use pack.type =
# "manifest" and one or more manifest blocks (each requires "name" and "content") instead of
# "uid"/"tag"/"registry_uid".
# ---------------------------------------------------------------------------------------------
# resource "spectrocloud_addon_deployment" "example_manifest" {
#   cluster_uid = var.cluster_uid
#
#   cluster_profile {
#     id = var.cluster_profile_uid
#
#     pack {
#       name = "custom-namespace"
#       type = "manifest"
#
#       manifest {
#         name = "custom-namespace"
#         content = <<-EOT
#           apiVersion: v1
#           kind: Namespace
#           metadata:
#             name: custom-namespace
#         EOT
#       }
#     }
#   }
# }

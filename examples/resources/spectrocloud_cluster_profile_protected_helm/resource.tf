# This example shows a Helm chart pack sourced from a "protected" (authenticated) OCI registry -
# one that has already been registered in Palette with credentials, as opposed to a public,
# unauthenticated OCI registry. The registry itself must already exist in Palette; this
# configuration only looks it up by name and uses it in the profile's pack.
#
# Day-2 mutability: on spectrocloud_cluster_profile, `cloud` and `type` are ForceNew - changing
# either recreates the profile. `pack` (including this Helm pack) updates in place.

data "spectrocloud_registry_oci" "registry1" {
  name = "my-protected-oci-registry"
}

resource "spectrocloud_cluster_profile" "profile_resource" {
  cloud       = "eks"
  description = "addon-profile-1"
  name        = "addon-profile-1"
  type        = "add-on"

  pack {
    name         = "kubevious-test"
    type         = "helm"
    registry_uid = data.spectrocloud_registry_oci.registry1.id
    tag          = "0.8.15"
    # uid is left unset here - since name, tag, and registry_uid are all provided, the provider
    # resolves the pack's UID internally rather than requiring it to be looked up separately.
    values = <<-EOT
      pack:
        namespace: "helm-test-chart"
        spectrocloud.com/install-priority: "230"
        releaseNameOverride:
          test-chart-service: test-chart-service-name
    EOT
  }
}

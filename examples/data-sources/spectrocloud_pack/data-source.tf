# Looks up a single pack in the Palette registries. There are three mutually exclusive lookup
# modes - set attributes from exactly one of them:
#   1. Simple lookup (shown in "example" below): name (+ optional version/cloud/registry_uid/type).
#      If version is omitted, the latest available version is used.
#   2. Advanced filter (shown in "filtered" below): advance_filters (+ name/registry_uid),
#      structured filtering by pack_type/addon_type/pack_layer/environment/is_fips/pack_source.
#   3. Raw filter string: `filters`, a "key=value AND/OR ..." expression - deprecated in favor
#      of advance_filters; conflicts with id/cloud/name/version/registry_uid.
# A fourth mode looks up by `id` directly; it conflicts with filters/cloud/name/version/registry_uid.
#
# "example" (simple lookup) flat attributes - all lookup keys, optional, also Computed:
#   name         - Pack name (e.g., "nginx-pack", "k8s-core").
#   version      - If omitted, the latest available version is used.
#   cloud        - Set of cloud types to filter by; "all" is implied.
#   registry_uid - Registry to search within.
#   type         - Allowed: "helm", "manifest", "container", "operator-instance".

# Retrieve details of a specific pack using name and version
data "spectrocloud_pack" "example" {
  name    = "nginx-pack"
  version = "1.2.3"
  # cloud        = ["aws"]
  # registry_uid = "5ee9c5adc172449eeb9c30cf"
  # type         = "helm"
}

# Retrieve a pack using advanced filters
#
# Flat attributes:
#   name         - Pack name to search for.
#   registry_uid - Unique registry identifier.
data "spectrocloud_pack" "filtered" {
  name = "k8sgpt-operator"

  # advance_filters block (at most one), structured filtering by pack_type/addon_type/pack_layer/
  # environment/is_fips/pack_source:
  #   pack_type   - Allowed: "helm", "spectro", "oci", "manifest".
  #   addon_type  - Allowed: "load balancer", "ingress", "logging", "monitoring", "security",
  #                 "authentication", "servicemesh", "system app", "app services", "registry",
  #                 "csi", "cni", "integration".
  #   pack_layer  - Allowed: "kernel", "os", "k8s", "cni", "csi", "addon".
  #   environment - Allowed: "all", "aws", "eks", "azure", "aks", "gcp", "gke", "vsphere", "maas",
  #                 "edge-native".
  #   is_fips     - Boolean: true (FIPS-compliant) / false (default).
  #   pack_source - Allowed: "spectrocloud", "community".
  advance_filters {
    pack_type   = ["spectro"]
    addon_type  = ["system app"]
    pack_layer  = ["addon"]
    environment = ["all"]
    is_fips     = false
    pack_source = ["community"]
  }

  registry_uid = "5ee9c5adc172449eeb9c30cf"
}

# Output pack details (all Computed)
output "pack_id" {
  value = data.spectrocloud_pack.example.id
}

output "pack_version" {
  value = data.spectrocloud_pack.example.version
}

output "pack_cloud_types" {
  value = data.spectrocloud_pack.example.cloud
}

output "pack_registry_uid" {
  value = data.spectrocloud_pack.example.registry_uid
}

output "pack_values" {
  value = data.spectrocloud_pack.example.values # YAML values of the pack, as a string
}

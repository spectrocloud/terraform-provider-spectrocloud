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
# Retrieve details of a specific pack using name and version
data "spectrocloud_pack" "example" {
  # Lookup key, optional, also Computed. Pack name (e.g., "nginx-pack", "k8s-core").
  name = "nginx-pack"
  # Lookup key, optional, also Computed. If omitted, the latest available version is used.
  version = "1.2.3"
  # Lookup key, optional, also Computed. Set of cloud types to filter by; "all" is implied.
  # cloud = ["aws"]
  # Lookup key, optional, also Computed. Registry to search within.
  # registry_uid = "5ee9c5adc172449eeb9c30cf"
  # Lookup key, optional, also Computed. Allowed: "helm", "manifest", "container",
  # "operator-instance".
  # type = "helm"
}

# Retrieve a pack using advanced filters
data "spectrocloud_pack" "filtered" {
  name = "k8sgpt-operator" # Pack name to search for

  advance_filters {
    pack_type   = ["spectro"]    # Allowed: "helm", "spectro", "oci", "manifest"
    addon_type  = ["system app"] # Allowed: "load balancer", "ingress", "logging", "monitoring", "security", "authentication", "servicemesh", "system app", "app services", "registry", "csi", "cni", "integration"
    pack_layer  = ["addon"]      # Allowed: "kernel", "os", "k8s", "cni", "csi", "addon"
    environment = ["all"]        # Allowed: "all", "aws", "eks", "azure", "aks", "gcp", "gke", "vsphere", "maas", "edge-native"
    is_fips     = false          # Boolean: true (FIPS-compliant) / false (default)
    pack_source = ["community"]  # Allowed: "spectrocloud", "community"
  }

  registry_uid = "5ee9c5adc172449eeb9c30cf" # Unique registry identifier
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

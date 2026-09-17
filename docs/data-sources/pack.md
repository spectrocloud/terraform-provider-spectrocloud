---
page_title: "spectrocloud_pack Data Source - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  This data resource provides the ability to search for a pack in the Palette registries. It supports more advanced search criteria than the pack_simple data source.
---

# spectrocloud_pack (Data Source)

  This data resource provides the ability to search for a pack in the Palette registries. It supports more advanced search criteria than the pack_simple data source.



~> The existing `filters` attribute will be deprecated, and a new `pack_filters` attribute will be introduced for advanced search functionality.

## Example Usage

~> For certain packs such as vm-migration-assistant, virtual-machine-orchestrator, vm-migration-assistant-pack, spectro-k8s-dashboard, and spectro-vm-dashboard, the `addon_type` is considered as `integration`.

```terraform
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
```

### Deprecated: raw `filters` string

The `filters` attribute is deprecated in favor of `advance_filters` shown above, but remains supported. It is a string that can contain multiple filters separated by the `AND`/`OR` operator, using the attributes returned in the `spec` object of the payload provided by the `v1/packs/search` endpoint. Refer to the Palette Pack Search API endpoint [documentation](https://docs.spectrocloud.com/api/v1/v-1-packs-search/) for more information on the available filters.

In this example, a filter is applied to retrieve a Calico CNI pack from the Palette OCI registry that is compatible with Edge clusters and has a version greater than 3.26.9.

```hcl
data "spectrocloud_registry" "palette_registry_oci" {
  name = "Palette Registry"
}

data "spectrocloud_pack" "cni" {
  filters = "spec.cloudTypes=edge-nativeANDspec.layer=cniANDspec.displayName=CalicoANDspec.version>3.26.9ANDspec.registryUid=${data.spectrocloud_registry.palette_registry_oci.id}"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `advance_filters` (Block List, Max: 1) A set of advanced filters to refine the selection of packs. These filters allow users to specify criteria such as pack type, add-on type, pack layer, and environment. (see [below for nested schema](#nestedblock--advance_filters))
- `cloud` (Set of String) Set of cloud type strings used to filter results. If not provided, all cloud types are returned.
- `filters` (String) Filters to apply when searching for a pack. This is a string of the form 'key1=value1' with 'AND', 'OR` operators. Refer to the Palette API [pack search API endpoint documentation](https://docs.spectrocloud.com/api/v1/v-1-packs-search/) for filter examples. The filter attribute will be deprecated soon; use `advance_filter` instead.
- `id` (String) The UID of the pack returned.
- `name` (String) The name of the pack to search for.
- `registry_uid` (String) The unique identifier (UID) of the registry where the pack is located. Specify `registry_uid` to search within a specific registry.
- `type` (String) The type of pack to search for. Supported values are `helm`, `manifest`, `container`, `operator-instance`.
- `version` (String) Specify the version of the pack to search for. If not set, the latest available version from the specified registry will be used.

### Read-Only

- `values` (String) The YAML values of the pack returned as string.

<a id="nestedblock--advance_filters"></a>
### Nested Schema for `advance_filters`

Optional:

- `addon_type` (Set of String) Set of add-on type strings to filter by. Allowed values are `load balancer`, `ingress`, `logging`, `monitoring`, `security`, `authentication`, `servicemesh`, `system app`, `app services`, `registry`, and `integration`. If not specified, all options are included. For `storage` and `network`, set `pack_layer` to `csi` or `cni` respectively.
- `environment` (Set of String) Defines the environment where the pack will be deployed. Options include `all`, `aws`, `eks`, `azure`, `aks`, `gcp`, `gke`, `vsphere`, `maas` and `edge-native`. If not specified, all options will be set by default.
- `is_fips` (Boolean) Indicates whether the pack is FIPS-compliant. If `true`, only FIPS-compliant components will be used.
- `pack_layer` (Set of String) Indicates the pack layer, such as `kernel`, `os`, `k8s`, `cni`, `csi`, or `addon`. If not specified, all options will be set by default.
- `pack_source` (Set of String) Specify the source of the pack. Allowed values are `spectrocloud` and `community`. If not specified, all options will be set by default.
- `pack_type` (Set of String) Specify the type of pack. Allowed values are `helm`, `spectro`, `oci`, and `manifest`. If not specified, all options will be set by default.

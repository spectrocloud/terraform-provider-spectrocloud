---
page_title: "spectrocloud_cluster_tke Resource - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  
---

# Resource `spectrocloud_cluster_tke`





## Schema

### Required

- **cloud_account_id** (String)
- **cloud_config** (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--cloud_config))
- **machine_pool** (Block List, Min: 1) (see [below for nested schema](#nestedblock--machine_pool))
- **name** (String)

### Optional

- **backup_policy** (Block List, Max: 1) (see [below for nested schema](#nestedblock--backup_policy))
- **cluster_profile** (Block List) (see [below for nested schema](#nestedblock--cluster_profile))
- **cluster_rbac_binding** (Block List) (see [below for nested schema](#nestedblock--cluster_rbac_binding))
- **id** (String) The ID of this resource.
- **namespaces** (Block List) (see [below for nested schema](#nestedblock--namespaces))
- **pack** (Block List) (see [below for nested schema](#nestedblock--pack))
- **scan_policy** (Block List, Max: 1) (see [below for nested schema](#nestedblock--scan_policy))
- **tags** (Set of String)
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-only

- **cloud_config_id** (String)
- **kubeconfig** (String)

<a id="nestedblock--cloud_config"></a>
### Nested Schema for `cloud_config`

Required:

- **region** (String)

Optional:

- **az_subnets** (Map of String)
- **azs** (List of String)
- **endpoint_access** (String)
- **public_access_cidrs** (Set of String)
- **ssh_key_name** (String)
- **vpc_id** (String)


<a id="nestedblock--machine_pool"></a>
### Nested Schema for `machine_pool`

Required:

- **count** (Number)
- **disk_size_gb** (Number)
- **instance_type** (String)
- **name** (String)

Optional:

- **additional_labels** (Map of String)
- **az_subnets** (Map of String)
- **azs** (List of String)
- **capacity_type** (String)
- **max** (Number)
- **max_price** (String)
- **min** (Number)
- **taints** (Block List) (see [below for nested schema](#nestedblock--machine_pool--taints))

<a id="nestedblock--machine_pool--taints"></a>
### Nested Schema for `machine_pool.taints`

Required:

- **effect** (String)
- **key** (String)
- **value** (String)



<a id="nestedblock--backup_policy"></a>
### Nested Schema for `backup_policy`

Required:

- **backup_location_id** (String)
- **expiry_in_hour** (Number)
- **prefix** (String)
- **schedule** (String)

Optional:

- **include_cluster_resources** (Boolean)
- **include_disks** (Boolean)
- **namespaces** (Set of String)


<a id="nestedblock--cluster_profile"></a>
### Nested Schema for `cluster_profile`

Required:

- **id** (String) The ID of this resource.

Optional:

- **pack** (Block List) (see [below for nested schema](#nestedblock--cluster_profile--pack))

<a id="nestedblock--cluster_profile--pack"></a>
### Nested Schema for `cluster_profile.pack`

Required:

- **name** (String)

Optional:

- **manifest** (Block List) (see [below for nested schema](#nestedblock--cluster_profile--pack--manifest))
- **tag** (String)
- **type** (String)
- **values** (String)

<a id="nestedblock--cluster_profile--pack--manifest"></a>
### Nested Schema for `cluster_profile.pack.manifest`

Required:

- **content** (String)
- **name** (String)




<a id="nestedblock--cluster_rbac_binding"></a>
### Nested Schema for `cluster_rbac_binding`

Required:

- **type** (String)

Optional:

- **namespace** (String)
- **role** (Map of String)
- **subjects** (Block List) (see [below for nested schema](#nestedblock--cluster_rbac_binding--subjects))

<a id="nestedblock--cluster_rbac_binding--subjects"></a>
### Nested Schema for `cluster_rbac_binding.subjects`

Required:

- **name** (String)
- **type** (String)

Optional:

- **namespace** (String)



<a id="nestedblock--namespaces"></a>
### Nested Schema for `namespaces`

Required:

- **name** (String)
- **resource_allocation** (Map of String)


<a id="nestedblock--pack"></a>
### Nested Schema for `pack`

Required:

- **name** (String)
- **tag** (String)
- **values** (String)


<a id="nestedblock--scan_policy"></a>
### Nested Schema for `scan_policy`

Required:

- **configuration_scan_schedule** (String)
- **conformance_scan_schedule** (String)
- **penetration_scan_schedule** (String)


<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **create** (String)
- **delete** (String)
- **update** (String)


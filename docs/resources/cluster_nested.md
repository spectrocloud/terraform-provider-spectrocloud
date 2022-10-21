---
page_title: "spectrocloud_cluster_nested Resource - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  
---

# Resource `spectrocloud_cluster_nested`





## Schema

### Required

- **cluster_config** (Block List, Min: 1) (see [below for nested schema](#nestedblock--cluster_config))
- **name** (String)

### Optional

- **apply_setting** (String)
- **backup_policy** (Block List, Max: 1) (see [below for nested schema](#nestedblock--backup_policy))
- **cloud_config** (Block List, Max: 1) (see [below for nested schema](#nestedblock--cloud_config))
- **cluster_profile** (Block List) (see [below for nested schema](#nestedblock--cluster_profile))
- **cluster_rbac_binding** (Block List) (see [below for nested schema](#nestedblock--cluster_rbac_binding))
- **id** (String) The ID of this resource.
- **namespaces** (Block List) (see [below for nested schema](#nestedblock--namespaces))
- **os_patch_after** (String)
- **os_patch_on_boot** (Boolean)
- **os_patch_schedule** (String)
- **pack** (Block List) (see [below for nested schema](#nestedblock--pack))
- **scan_policy** (Block List, Max: 1) (see [below for nested schema](#nestedblock--scan_policy))
- **skip_completion** (Boolean)
- **tags** (Set of String)
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-only

- **cloud_config_id** (String)
- **kubeconfig** (String)

<a id="nestedblock--cluster_config"></a>
### Nested Schema for `cluster_config`

Required:

- **host_cluster_config** (Block List, Min: 1) (see [below for nested schema](#nestedblock--cluster_config--host_cluster_config))

Optional:

- **resources** (Block List) (see [below for nested schema](#nestedblock--cluster_config--resources))

<a id="nestedblock--cluster_config--host_cluster_config"></a>
### Nested Schema for `cluster_config.host_cluster_config`

Optional:

- **cluster_group** (Block List) (see [below for nested schema](#nestedblock--cluster_config--host_cluster_config--cluster_group))
- **host_cluster** (Block List) (see [below for nested schema](#nestedblock--cluster_config--host_cluster_config--host_cluster))

<a id="nestedblock--cluster_config--host_cluster_config--cluster_group"></a>
### Nested Schema for `cluster_config.host_cluster_config.cluster_group`

Optional:

- **uid** (String)


<a id="nestedblock--cluster_config--host_cluster_config--host_cluster"></a>
### Nested Schema for `cluster_config.host_cluster_config.host_cluster`

Optional:

- **uid** (String)



<a id="nestedblock--cluster_config--resources"></a>
### Nested Schema for `cluster_config.resources`

Optional:

- **max_cpu** (Number)
- **max_mem_in_mb** (Number)
- **min_cpu** (Number)
- **min_mem_in_mb** (Number)



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


<a id="nestedblock--cloud_config"></a>
### Nested Schema for `cloud_config`

Optional:

- **chart_name** (String)
- **chart_repo** (String)
- **chart_values** (String)
- **chart_version** (String)
- **k8s_version** (String)


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
- **values** (String)

Optional:

- **manifest** (Block List) (see [below for nested schema](#nestedblock--cluster_profile--pack--manifest))
- **registry_uid** (String)
- **tag** (String)
- **type** (String)

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

Optional:

- **registry_uid** (String)


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



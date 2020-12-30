---
page_title: "spectrocloud_cluster_azure Resource - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  
---

# Resource `spectrocloud_cluster_azure`





## Schema

### Required

- **cloud_account_id** (String)
- **cloud_config** (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--cloud_config))
- **cluster_profile_id** (String)
- **machine_pool** (Block Set, Min: 1) (see [below for nested schema](#nestedblock--machine_pool))
- **name** (String)

### Optional

- **id** (String) The ID of this resource.
- **pack** (Block Set) (see [below for nested schema](#nestedblock--pack))
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-only

- **cloud_config_id** (String)

<a id="nestedblock--cloud_config"></a>
### Nested Schema for `cloud_config`

Required:

- **location** (String)
- **resource_group** (String)
- **ssh_key** (String)
- **subscription_id** (String)


<a id="nestedblock--machine_pool"></a>
### Nested Schema for `machine_pool`

Required:

- **count** (Number)
- **instance_type** (String)
- **name** (String)

Optional:

- **control_plane** (Boolean)
- **control_plane_as_worker** (Boolean)
- **disk** (Block List, Max: 1) (see [below for nested schema](#nestedblock--machine_pool--disk))
- **update_strategy** (String)

<a id="nestedblock--machine_pool--disk"></a>
### Nested Schema for `machine_pool.disk`

Required:

- **size_gb** (Number)
- **type** (String)



<a id="nestedblock--pack"></a>
### Nested Schema for `pack`

Required:

- **name** (String)
- **tag** (String)
- **values** (String)


<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **create** (String)
- **delete** (String)
- **update** (String)



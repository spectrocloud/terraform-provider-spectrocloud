---
page_title: "spectrocloud_cluster_vsphere Resource - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  
---

# Resource `spectrocloud_cluster_vsphere`





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

- **datacenter** (String)
- **folder** (String)
- **network_search_domain** (String)
- **network_type** (String)
- **ssh_key** (String)


<a id="nestedblock--machine_pool"></a>
### Nested Schema for `machine_pool`

Required:

- **count** (Number)
- **instance_type** (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--machine_pool--instance_type))
- **name** (String)
- **placement** (Block List, Min: 1) (see [below for nested schema](#nestedblock--machine_pool--placement))

Optional:

- **control_plane** (Boolean)
- **control_plane_as_worker** (Boolean)
- **update_strategy** (String)

<a id="nestedblock--machine_pool--instance_type"></a>
### Nested Schema for `machine_pool.instance_type`

Required:

- **cpu** (Number)
- **disk_size_gb** (Number)
- **memory_mb** (Number)


<a id="nestedblock--machine_pool--placement"></a>
### Nested Schema for `machine_pool.placement`

Required:

- **cluster** (String)
- **datastore** (String)
- **network** (String)
- **resource_pool** (String)



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



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
- **os_patch_on_boot** (Boolean, Optional) OS Patch on boot when set, updates security patch of host OS 
of all nodes and monitors new nodes (which gets created when cluster is scaled up or cluster k8s version is upgraded) for security patch
- **os_patch_schedule** (String, Optional) Cron schedule to patch security updates on host OS for all nodes. Please see https://en.wikipedia.org/wiki/Cron for valid cron syntax
- **os_patch_after** (String, Optional) On demand security patch on host OS for all nodes. Please follow RFC3339 Date and Time Standards. Eg 2021-01-01T00:00:00.000Z

### Read-only

- **cloud_config_id** (String)
- **kubeconfig** (String)

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



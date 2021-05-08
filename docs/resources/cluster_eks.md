---
page_title: "spectrocloud_cluster_eks Resource - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  
---

# Resource `spectrocloud_cluster_eks`





## Schema

### Required

- **cloud_account_id** (String)
- **cloud_config** (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--cloud_config))
- **cluster_profile_id** (String)
- **machine_pool** (Block Set, Min: 1) (see [below for nested schema](#nestedblock--machine_pool))
- **name** (String)

### Optional

- **id** (String) The ID of this resource.
- **pack** (Block List) (see [below for nested schema](#nestedblock--pack))
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-only

- **cloud_config_id** (String)
- **kubeconfig** (String)

<a id="nestedblock--cloud_config"></a>
### Nested Schema for `cloud_config`

Required:

- **region** (String)
- **ssh_key_name** (String)

Optional:

- **endpoint_access** (String)
- **public_access_cidrs** (Set of String)
- **vpc_id** (String)


<a id="nestedblock--machine_pool"></a>
### Nested Schema for `machine_pool`

Required:

- **count** (Number)
- **instance_type** (String)
- **name** (String)

Optional:

- **az_subnets** (Map of String)
- **azs** (Set of String)
- **control_plane** (Boolean)
- **control_plane_as_worker** (Boolean)
- **disk_size_gb** (Number)
- **update_strategy** (String)


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



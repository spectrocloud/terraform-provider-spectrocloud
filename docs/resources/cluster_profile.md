---
page_title: "spectrocloud_cluster_profile Resource - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  
---

# Resource `spectrocloud_cluster_profile`





## Schema

### Required

- **cloud** (String)
- **name** (String)
- **pack** (Block List, Min: 1) (see [below for nested schema](#nestedblock--pack))

### Optional

- **description** (String)
- **id** (String) The ID of this resource.
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- **type** (String)

<a id="nestedblock--pack"></a>
### Nested Schema for `pack`

Required:

- **name** (String)
- **tag** (String)
- **uid** (String)
- **values** (String)


<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **create** (String)
- **delete** (String)
- **update** (String)



---
page_title: "spectrocloud_cluster_import Resource - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  
---

# Resource `spectrocloud_cluster_import`





## Schema

### Required

- **cloud** (String)
- **name** (String)

### Optional

- **cluster_profile_id** (String)
- **id** (String) The ID of this resource.
- **pack** (Block List) (see [below for nested schema](#nestedblock--pack))
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-only

- **cloud_config_id** (String)
- **cluster_import_manifest** (String)
- **cluster_import_manifest_apply_command** (String)

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



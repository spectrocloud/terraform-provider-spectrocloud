---
page_title: "spectrocloud_registry_oci Resource - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  
---

# Resource `spectrocloud_registry_oci`





## Schema

### Required

- **credentials** (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--credentials))
- **endpoint** (String)
- **is_private** (Boolean)
- **name** (String)
- **type** (String)

### Optional

- **id** (String) The ID of this resource.
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

<a id="nestedblock--credentials"></a>
### Nested Schema for `credentials`

Required:

- **credential_type** (String)

Optional:

- **arn** (String)
- **external_id** (String)


<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **create** (String)
- **delete** (String)
- **update** (String)



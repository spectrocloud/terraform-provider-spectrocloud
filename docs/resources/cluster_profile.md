---
page_title: "spectrocloud_cluster_profile Resource - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  
---

# Resource `spectrocloud_cluster_profile`



## Example Usage

```terraform
resource "spectrocloud_cluster_profile" "test-tf" {
  name        = "test-tf"
  description = "Terraform Test Profile"
  tags        = ["createdBy:spectro"]
  type        = "add-on"
  pack {
    name = "ambassador"
    type = "spectro"
    tag  = "6.6.0"
  }

}
```

## Schema

### Required

- **name** (String)
- **pack** (Block List, Min: 1) (see [below for nested schema](#nestedblock--pack))

### Optional

- **cloud** (String)
- **description** (String)
- **id** (String) The ID of this resource.
- **tags** (Set of String)
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- **type** (String)
- **version** (String)

<a id="nestedblock--pack"></a>
### Nested Schema for `pack`

Required:

- **name** (String)

Optional:

- **manifest** (Block List) (see [below for nested schema](#nestedblock--pack--manifest))
- **registry_uid** (String)
- **tag** (String)
- **type** (String)
- **uid** (String)
- **values** (String)

<a id="nestedblock--pack--manifest"></a>
### Nested Schema for `pack.manifest`

Required:

- **content** (String)
- **name** (String)

Read-only:

- **uid** (String)



<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **create** (String)
- **delete** (String)
- **update** (String)



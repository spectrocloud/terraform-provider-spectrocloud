---
page_title: "spectrocloud_project Resource - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  
---

# Resource `spectrocloud_project`



## Example Usage

```terraform
resource "spectrocloud_team" "project" {
  name = "dev1"
}
```

## Schema

### Required

- **name** (String)

### Optional

- **description** (String)
- **id** (String) The ID of this resource.
- **tags** (Set of String)
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **create** (String)
- **delete** (String)
- **update** (String)



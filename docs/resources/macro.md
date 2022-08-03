---
page_title: "spectrocloud_macro Resource - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  
---

# Resource `spectrocloud_macro`



## Example Usage

```terraform
resource "spectrocloud_macro" "project_macro" {
  name    = "project1"
  value   = "project_val2"
  project = "Default"
}

resource "spectrocloud_macro" "tenant_macro" {
  name  = "tenant1"
  value = "tenant_val1"
}
```

## Schema

### Required

- **name** (String)

### Optional

- **id** (String) The ID of this resource.
- **project** (String)
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- **value** (String)

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **create** (String)
- **delete** (String)
- **update** (String)



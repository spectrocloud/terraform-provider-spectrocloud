---
page_title: "spectrocloud_appliance Resource - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  
---

# Resource `spectrocloud_appliance`



## Example Usage

```terraform
resource "spectrocloud_appliance" "appliance" {
  uid = "nik-libvirt15-mar-20"
  labels = {
    "name" = "nik_appliance_name"
  }
  wait = true
}
```

## Schema

### Required

- **uid** (String)

### Optional

- **id** (String) The ID of this resource.
- **labels** (Map of String)
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- **wait** (Boolean)

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **create** (String)
- **delete** (String)
- **update** (String)



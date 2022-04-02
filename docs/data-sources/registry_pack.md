---
page_title: "spectrocloud_registry_pack Data Source - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  
---

# Data Source `spectrocloud_registry_pack`



## Example Usage

```terraform
data "spectrocloud_registry_pack" "registry1" {
  name = "Public Repo"

}

output "test" {
  value = data.spectrocloud_registry_pack.registry1
}
```

## Schema

### Required

- **name** (String)

### Optional

- **id** (String) The ID of this resource.



---
page_title: "spectrocloud_role Data Source - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  
---

# Data Source `spectrocloud_role`



## Example Usage

```terraform
data "spectrocloud_role" "role1" {
  name = "Project Editor"

  # (alternatively)
  # id =  "5fd0ca727c411c71b55a359c"
}
```

## Schema

### Optional

- **id** (String) The ID of this resource.
- **name** (String)



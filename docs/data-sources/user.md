---
page_title: "spectrocloud_user Data Source - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  
---

# Data Source `spectrocloud_user`



## Example Usage

```terraform
data "spectrocloud_user" "user1" {
  name = "Foo Bar"

  # (alternatively)
  # id =  "5fd0ca727c411c71b55a359c"
}
```

## Schema

### Optional

- **id** (String) The ID of this resource.
- **name** (String)



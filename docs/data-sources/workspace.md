---
page_title: "spectrocloud_workspace Data Source - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  
---

# Data Source `spectrocloud_workspace`



## Example Usage

```terraform
data "spectrocloud_workspace" "workspace" {
  name = "wsp-tf"
}

output "same" {
  value = data.spectrocloud_workspace.workspace
}
```

## Schema

### Required

- **name** (String)

### Optional

- **id** (String) The ID of this resource.



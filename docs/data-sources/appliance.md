---
page_title: "spectrocloud_appliance Data Source - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  
---

# Data Source `spectrocloud_appliance`



## Example Usage

```terraform
data "spectrocloud_appliance" "test_appliance" {
  name = "nik-test-1"
}

output "same" {
  value = data.spectrocloud_appliance.test_appliance
}
```

## Schema

### Required

- **name** (String)

### Optional

- **id** (String) The ID of this resource.
- **labels** (Map of String)



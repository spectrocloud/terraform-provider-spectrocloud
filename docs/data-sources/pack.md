---
page_title: "spectrocloud_pack Data Source - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  
---

# Data Source `spectrocloud_pack`



## Example Usage

```terraform
data "spectrocloud_pack" "cni-calico" {
  name    = "cni-calico"
  version = "3.16.0"

  # (alternatively)
  # id =  "5fd0ca727c411c71b55a359c"
  # name = "cni-calico-azure"
  # cloud = ["azure"]
}

output "same" {
  value = data.spectrocloud_pack.cni-calico
}
```

## Schema

### Optional

- **cloud** (Set of String)
- **filters** (String)
- **id** (String) The ID of this resource.
- **name** (String)
- **registry_uid** (String)
- **type** (String)
- **version** (String)

### Read-only

- **values** (String)



---
page_title: "spectrocloud_registry_helm Data Source - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  
---

# Data Source `spectrocloud_registry_helm`



## Example Usage

```terraform
data "spectrocloud_registry_helm" "registry1" {
  name = "spectro-helm-repo"

}

output "test" {
  value = data.spectrocloud_registry_helm.registry1
}
```

## Schema

### Required

- **name** (String)

### Optional

- **id** (String) The ID of this resource.



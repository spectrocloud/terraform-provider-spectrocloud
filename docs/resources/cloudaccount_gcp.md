---
page_title: "spectrocloud_cloudaccount_gcp Resource - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  
---

# Resource `spectrocloud_cloudaccount_gcp`



## Example Usage

```terraform
resource "spectrocloud_cloudaccount_gcp" "azure-1" {
  name                 = "gcp-1"
  gcp_json_credentials = var.gcp_serviceaccount_json
}
```

## Schema

### Required

- **gcp_json_credentials** (String, Sensitive)
- **name** (String)

### Optional

- **id** (String) The ID of this resource.



---
page_title: "spectrocloud_cloudaccount_maas Resource - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  
---

# Resource `spectrocloud_cloudaccount_maas`



## Example Usage

```terraform
resource "spectrocloud_cloudaccount_maas" "maas-1" {
  name              = "maas-1"
  maas_api_endpoint = var.maas_api_endpoint
  maas_api_key      = var.maas_api_key
}
```

## Schema

### Required

- **name** (String)

### Optional

- **id** (String) The ID of this resource.
- **maas_api_endpoint** (String)
- **maas_api_key** (String, Sensitive)
- **private_cloud_gateway_id** (String)



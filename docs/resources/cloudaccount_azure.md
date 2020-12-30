---
page_title: "spectrocloud_cloudaccount_azure Resource - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  
---

# Resource `spectrocloud_cloudaccount_azure`



## Example Usage

```terraform
resource "spectrocloud_cloudaccount_azure" "azure-1" {
  name                = "azure-1"
  azure_tenant_id     = "<....>"
  azure_client_id     = "<....>"
  azure_client_secret = "<....>"
}
```

## Schema

### Required

- **azure_client_id** (String)
- **azure_client_secret** (String, Sensitive)
- **azure_tenant_id** (String)
- **name** (String)

### Optional

- **id** (String) The ID of this resource.



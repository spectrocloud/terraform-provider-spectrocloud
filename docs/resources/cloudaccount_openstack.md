---
page_title: "spectrocloud_cloudaccount_openstack Resource - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  
---

# Resource `spectrocloud_cloudaccount_openstack`



## Example Usage

```terraform
resource "spectrocloud_cloudaccount_openstack" "account" {
  name                     = "openstack-dev"
  private_cloud_gateway_id = ""
  openstack_username       = var.openstack_username
  openstack_password       = var.openstack_password
  identity_endpoint        = var.identity_endpoint
  parent_region            = var.region
  default_domain           = var.domain
  default_project          = var.project
}
```

## Schema

### Required

- **default_domain** (String)
- **default_project** (String)
- **identity_endpoint** (String)
- **name** (String)
- **openstack_password** (String, Sensitive)
- **openstack_username** (String)
- **parent_region** (String)
- **private_cloud_gateway_id** (String)

### Optional

- **ca_certificate** (String)
- **id** (String) The ID of this resource.
- **openstack_allow_insecure** (Boolean)



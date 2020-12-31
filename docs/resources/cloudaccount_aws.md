---
page_title: "spectrocloud_cloudaccount_aws Resource - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  
---

# Resource `spectrocloud_cloudaccount_aws`



## Example Usage

```terraform
resource "spectrocloud_cloudaccount_aws" "aws-1" {
  name           = "aws-1"
  aws_access_key = var.aws_access_key
  aws_secret_key = var.aws_secret_key
}
```

## Schema

### Required

- **aws_access_key** (String)
- **aws_secret_key** (String, Sensitive)
- **name** (String)

### Optional

- **id** (String) The ID of this resource.



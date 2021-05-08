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

- **name** (String)

### Optional

- **arn** (String)
- **aws_access_key** (String)
- **aws_secret_key** (String, Sensitive)
- **external_id** (String, Sensitive)
- **id** (String) The ID of this resource.
- **type** (String)



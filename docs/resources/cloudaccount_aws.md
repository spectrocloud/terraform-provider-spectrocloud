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
```terraform
resource "spectrocloud_cloudaccount_aws" "aws-2" {
  name           = "aws-1"
  type           = "sts"
  arn            = var.arn
  external_id    = var.external_id
}
```


## Schema

### Required

- **aws_access_key** (String) & **aws_secret_key** (String, Sensitive) if `type` is `secret`
- **arn** (String) & **external_id** (String, Sensitive) if `type` is `sts`
- **name** (String)

### Optional

- **id** (String) The ID of this resource.



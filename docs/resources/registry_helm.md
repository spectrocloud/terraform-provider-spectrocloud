---
page_title: "spectrocloud_registry_helm Resource - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  
---

# Resource `spectrocloud_registry_helm`



## Example Usage

```terraform
resource "spectrocloud_registry_helm" "r1" {
  name       = "us-artifactory"
  endpoint   = "https://123456.dkr.ecr.us-west-1.amazonaws.com"
  is_private = true
  credentials {
    credential_type = "noAuth"
    username        = "abc"
    password        = "def"
  }
}
```

## Schema

### Required

- **credentials** (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--credentials))
- **endpoint** (String)
- **is_private** (Boolean)
- **name** (String)

### Optional

- **id** (String) The ID of this resource.
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

<a id="nestedblock--credentials"></a>
### Nested Schema for `credentials`

Required:

- **credential_type** (String)

Optional:

- **password** (String)
- **token** (String)
- **username** (String)


<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **create** (String)
- **delete** (String)
- **update** (String)



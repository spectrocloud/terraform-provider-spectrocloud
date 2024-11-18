---
page_title: "spectrocloud_role Resource - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  The role resource allows you to manage roles in Palette.
---

# spectrocloud_role (Resource)

  The role resource allows you to manage roles in Palette.

You can learn more about managing roles in Palette by reviewing the [Roles](https://docs.spectrocloud.com/glossary-all/#role) guide.

## Example Usage

```terraform
variable "roles" {
  type    = list(string)
  default = ["Cluster Admin", "Cluster Profile Editor"]
}

# Data source loop to retrieve multiple roles
data "spectrocloud_role" "roles" {
  for_each = toset(var.roles)
  name     = each.key
}

resource "spectrocloud_role" "custom_role" {
  name        = "Test Cluster Role"
  type        = "project"
  permissions = flatten([for role in data.spectrocloud_role.roles : role.permissions])
}
```

```
### Importing existing role state & config

```hcl
# import existing user example
  import {
    to = spectrocloud_role.test_role
    id = "{roleUID}"
  }

# To generate TF configuration.
  terraform plan -generate-config-out=test_role.tf

# To import State file
  terraform import spectrocloud_role.test_role {roleUID}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The name of the role.
- `permissions` (Set of String) The permission's assigned to the role.

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `type` (String) The role type. Allowed values are `project` or `tenant` or `project`

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `delete` (String)
- `update` (String)
---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "spectrocloud_role Data Source - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  
---

# spectrocloud_role (Data Source)



## Example Usage

```terraform
# Retrieve details of a specific role by name
data "spectrocloud_role" "example" {
  name = "admin-role"
}

# Output role ID
output "role_id" {
  value = data.spectrocloud_role.example.id
}

# Output permissions associated with the role
output "role_permissions" {
  value = data.spectrocloud_role.example.permissions
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `name` (String)

### Read-Only

- `id` (String) The ID of this resource.
- `permissions` (Set of String) List of permissions associated with the role.

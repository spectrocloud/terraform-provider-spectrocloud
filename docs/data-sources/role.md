---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "spectrocloud_role Data Source - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  
---

# spectrocloud_role (Data Source)



## Example Usage

```terraform
data "spectrocloud_role" "role" {
  name = "Resource Cluster Admin"

  # (alternatively)
  # id =  "66fbea622947f81fb62294ac"
}

output "role_id" {
  value = data.spectrocloud_role.role.id
}

output "role_permissions" {
  value = data.spectrocloud_role.role.permissions
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `name` (String)

### Read-Only

- `id` (String) The ID of this resource.
- `permissions` (Set of String) List of permissions associated with the role.

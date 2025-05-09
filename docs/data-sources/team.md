---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "spectrocloud_team Data Source - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  
---

# spectrocloud_team (Data Source)



## Example Usage

```terraform
# Fetch details of a specific team in SpectroCloud
data "spectrocloud_team" "example" {
  # Provide either `id` or `name`, but not both.
  # Allowed values:
  # - `id`: A unique identifier for the team (e.g., "team-12345").
  # - `name`: The readable name of the team (e.g., "DevOps Team").

  id = "team-12345"
  # name = "DevOps Team"  # Alternative way to reference a team by name
}

output "team_info" {
  value = {
    id       = data.spectrocloud_team.example.id
    name     = data.spectrocloud_team.example.name
    role_ids = data.spectrocloud_team.example.role_ids
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `id` (String) The unique ID of the team. If provided, `name` cannot be used.
- `name` (String) The name of the team. If provided, `id` cannot be used.

### Read-Only

- `role_ids` (List of String) The roles id's assigned to the team.

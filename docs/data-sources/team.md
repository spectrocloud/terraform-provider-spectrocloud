---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "spectrocloud_team Data Source - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  
---

# spectrocloud_team (Data Source)



## Example Usage

```terraform
data "spectrocloud_team" "team1" {
  name = "team2"

  # (alternatively)
  # id =  "5fd0ca727c411c71b55a359c"
}

output "team-id" {
  value = data.spectrocloud_team.team1.id
}

output "team-role-ids" {
  value = data.spectrocloud_team.team1.role_ids
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `id` (String) The unique ID of the team. If provided, `name` cannot be used.
- `name` (String) The name of the team. If provided, `id` cannot be used.

### Read-Only

- `role_ids` (List of String) The roles id's assigned to the team.

---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "spectrocloud_appliances Data Source - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  
---

# spectrocloud_appliances (Data Source)



## Example Usage

```terraform
data "spectrocloud_appliances" "appliances" {
  #tags = ["env:prod", "store:502"]
  tags = ["env:dev"]
}

output "same" {
  value = data.spectrocloud_appliances.appliances
  #value = [for a in data.spectrocloud_appliance.appliances : a.name]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `tags` (Set of String)

### Read-Only

- `id` (String) The ID of this resource.
- `ids` (List of String)


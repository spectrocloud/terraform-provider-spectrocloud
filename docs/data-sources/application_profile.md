---
page_title: "spectrocloud_application_profile Data Source - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  Use this data source to get the details of an existing application profile.
---

# spectrocloud_application_profile (Data Source)

  Use this data source to get the details of an existing application profile.

## Example Usage

```hcl
# Retrieve details of a specific application profile
data "spectrocloud_application_profile" "example_profile" {
  name = "my-app-profile"  # Specify the name of the application profile
}

# Output the retrieved application profile details
output "application_profile_version" {
  value = data.spectrocloud_application_profile.example_profile.version
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Name of the application profile

### Optional

- `version` (String) The version of the app profile. Default value is '1.0.0'.

### Read-Only

- `id` (String) The ID of this resource.
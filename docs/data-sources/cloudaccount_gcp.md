---
page_title: "spectrocloud_cloudaccount_gcp Data Source - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  A data source for retrieving information about a GCP cloud account registered in Palette.
---

# spectrocloud_cloudaccount_gcp (Data Source)

  A data source for retrieving information about a GCP cloud account registered in Palette.

## Example Usage


You can retrieve the details of a GCP cloud registered in Palette by specifying the ID of the cloud account.

```hcl

# Retrieve details of a GCP cloud account using name
data "spectrocloud_cloud_account_gcp" "example" {
  name    = "example-gcp-account"  # Required if 'id' is not provided
  context = "project"              # Optional: Allowed values are "project", "tenant", or "" (default)
}

# Consolidated output as a map
output "gcp_account_details" {
  value = {
    id      = data.spectrocloud_cloud_account_gcp.example.id
    name    = data.spectrocloud_cloud_account_gcp.example.name
    context = data.spectrocloud_cloud_account_gcp.example.context
  }
}

```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `context` (String) The context of the cluster. Allowed values are `project` or `tenant` or ``.
- `id` (String) ID of the GCP cloud account registered in Palette.
- `name` (String) Name of the GCP cloud account registered in Palette.
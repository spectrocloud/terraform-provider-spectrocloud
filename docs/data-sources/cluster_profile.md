---
page_title: "spectrocloud_cluster_profile Data Source - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  
---

# Data Source `spectrocloud_cluster_profile`



## Example Usage

```terraform
data "spectrocloud_cluster_profile" "profile1" {
  name = "niktest_profile"
}

output "same" {
  value = data.spectrocloud_cluster_profile.profile1
}
```

## Schema

### Optional

- **context** (String)
- **id** (String) The ID of this resource.
- **name** (String)
- **version** (String)

### Read-only

- **pack** (List of Object) (see [below for nested schema](#nestedatt--pack))

<a id="nestedatt--pack"></a>
### Nested Schema for `pack`

Read-only:

- **manifest** (List of Object) (see [below for nested schema](#nestedobjatt--pack--manifest))
- **name** (String)
- **registry_uid** (String)
- **tag** (String)
- **type** (String)
- **uid** (String)
- **values** (String)

<a id="nestedobjatt--pack--manifest"></a>
### Nested Schema for `pack.manifest`

Read-only:

- **content** (String)
- **name** (String)
- **uid** (String)



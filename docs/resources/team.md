---
page_title: "spectrocloud_team Resource - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  
---

# Resource `spectrocloud_team`



## Example Usage

```terraform
# user
data "spectrocloud_user" "user1" {
  name = "Foo Bar"
}

# role
data "spectrocloud_role" "role1" {
  name = "Project Editor"
}

data "spectrocloud_role" "role2" {
  name = "Cluster Admin"
}

data "spectrocloud_role" "role3" {
  name = "Project Admin"
}

# project
data "spectrocloud_project" "project1" {
  name = "Default"
}

data "spectrocloud_project" "project2" {
  name = "Prod"
}

resource "spectrocloud_team" "t1" {
  name  = "team1"
  users = [data.spectrocloud_user.user1.id]

  project_role_mapping {
    id    = data.spectrocloud_project.project1.id
    roles = [data.spectrocloud_role.role1.id, data.spectrocloud_role.role2.id]
  }

  project_role_mapping {
    id    = data.spectrocloud_project.project2.id
    roles = [data.spectrocloud_role.role3.id]
  }
}
```

## Schema

### Required

- **name** (String)
- **users** (Set of String)

### Optional

- **id** (String) The ID of this resource.
- **project_role_mapping** (Block List) (see [below for nested schema](#nestedblock--project_role_mapping))
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

<a id="nestedblock--project_role_mapping"></a>
### Nested Schema for `project_role_mapping`

Required:

- **id** (String) The ID of this resource.
- **roles** (Set of String)


<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **create** (String)
- **delete** (String)
- **update** (String)



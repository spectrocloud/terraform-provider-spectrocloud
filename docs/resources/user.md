---
page_title: "spectrocloud_user Resource - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  Create and manage projects in Palette.
---

# spectrocloud_user (Resource)

  Create and manage projects in Palette.

You can learn more about managing users in Palette by reviewing the [Users](https://docs.spectrocloud.com/user-management/) guide.

## Example Usage

An example of creating a user resource with assigned teams and custom roles in Palette.

```hcl
resource "spectrocloud_user" "user-test"{
  first_name = "tf"
  last_name = "test"
  email = "test-tf@spectrocloud.com"
  team_ids  = [data.spectrocloud_team.team2.id]
  project_role {
    project_id = data.spectrocloud_project.default.id
    role_ids =  [for r in data.spectrocloud_role.app_roles : r.id]
  }
  project_role {
    project_id = data.spectrocloud_project.ranjith.id
    role_ids = [for r in data.spectrocloud_role.app_roles : r.id]
  }

  tenant_role = [for t in data.spectrocloud_role.tenant_roles : t.id]

  workspace_role {
    project_id = data.spectrocloud_project.default.id
    workspace {
      id = data.spectrocloud_workspace.workspace.id
      role_ids = [for w in data.spectrocloud_role.workspace_roles : w.id]
    }
    workspace {
      id = data.spectrocloud_workspace.workspace2.id
      role_ids = ["66fbea622947f81fc26983e6"]
    }
  }

  resource_role {
    project_ids = [data.spectrocloud_project.default.id, data.spectrocloud_project.ranjith.id]
    filter_ids = [data.spectrocloud_filter.filter.id]
    role_ids = [for r in data.spectrocloud_role.resource_roles : r.id]
  }

  resource_role {
    project_ids = [data.spectrocloud_project.ranjith.id]
    filter_ids = [data.spectrocloud_filter.filter.id]
    role_ids = [for re in data.spectrocloud_role.resource_roles_editor : re.id]
  }

}
```

The example below demonstrates how to create an user with only assigned teams.

```hcl
resource "spectrocloud_user" "user-test"{
  first_name = "tf"
  last_name = "test"
  email = "test-tf@spectrocloud.com"
  team_ids  = [data.spectrocloud_team.team2.id]
}


```

### Importing existing user states

```hcl
# import existing user example
  import {
    to = spectrocloud_user.test_user
    id = "{userUID}"
  }

# To generate TF configuration.
  terraform plan -generate-config-out=test_user.tf

# To import State file
  terraform import spectrocloud_user.test_user {userUID}
```


<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `email` (String) The email of the user.
- `first_name` (String) The first name of the user.
- `last_name` (String) The last name of the user.

### Optional

- `project_role` (Block Set) List of project roles to be associated with the user. (see [below for nested schema](#nestedblock--project_role))
- `resource_role` (Block Set) (see [below for nested schema](#nestedblock--resource_role))
- `team_ids` (List of String) The team id's assigned to the user.
- `tenant_role` (Set of String) List of tenant role ids to be associated with the user.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `workspace_role` (Block Set) List of workspace roles to be associated with the user. (see [below for nested schema](#nestedblock--workspace_role))

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--project_role"></a>
### Nested Schema for `project_role`

Required:

- `project_id` (String) Project id to be associated with the user.
- `role_ids` (Set of String) List of project role ids to be associated with the user.


<a id="nestedblock--resource_role"></a>
### Nested Schema for `resource_role`

Required:

- `filter_ids` (Set of String) List of filter ids.
- `project_ids` (Set of String) Project id's to be associated with the user.
- `role_ids` (Set of String) List of resource role ids to be associated with the user.


<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `delete` (String)
- `update` (String)


<a id="nestedblock--workspace_role"></a>
### Nested Schema for `workspace_role`

Required:

- `project_id` (String) Project id to be associated with the user.
- `workspace` (Block Set, Min: 1) List of workspace roles to be associated with the user. (see [below for nested schema](#nestedblock--workspace_role--workspace))

<a id="nestedblock--workspace_role--workspace"></a>
### Nested Schema for `workspace_role.workspace`

Required:

- `id` (String) Workspace id to be associated with the user.
- `role_ids` (Set of String) List of workspace role ids to be associated with the user.
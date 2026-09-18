# Day-2 mutability: `email` and `team_ids` are ForceNew - changing either recreates the user.
# `first_name`, `last_name`, `project_role`, `tenant_role`, `workspace_role`, and `resource_role`
# all update in place.
#
# Flat attributes:
#   first_name, last_name - Required. Updates in place.
#   email                 - Required, ForceNew. Must be a valid email address; also serves as
#                            the user's login.
#   team_ids              - Optional, ForceNew. Teams this user belongs to.
resource "spectrocloud_user" "user-test" {
  first_name = "tf"
  last_name  = "test"
  email      = "test-tf@spectrocloud.com"
  team_ids   = [data.spectrocloud_team.team2.id]

  # project_role (default project) block, repeatable: grants this user a set of roles scoped to
  # one project.
  project_role {
    project_id = data.spectrocloud_project.default.id
    role_ids   = [for r in data.spectrocloud_role.app_roles : r.id]
  }

  # project_role (ranjith project) block, repeatable.
  project_role {
    project_id = data.spectrocloud_project.ranjith.id
    role_ids   = [for r in data.spectrocloud_role.app_roles : r.id]
  }

  # tenant_role - Optional. Tenant-scoped role IDs granted to this user.
  tenant_role = [for t in data.spectrocloud_role.tenant_roles : t.id]

  # workspace_role block, repeatable: grants this user a set of roles scoped to one workspace
  # within a project - each nested `workspace` block names a workspace in that project plus its
  # roles.
  workspace_role {
    project_id = data.spectrocloud_project.default.id

    # workspace (workspace) block, repeatable.
    workspace {
      id       = data.spectrocloud_workspace.workspace.id
      role_ids = [for w in data.spectrocloud_role.workspace_roles : w.id]
    }

    # workspace (workspace2) block, repeatable.
    workspace {
      id       = data.spectrocloud_workspace.workspace2.id
      role_ids = [for w in data.spectrocloud_role.workspace_roles : w.id]
    }
  }

  # resource_role (app_roles, default+ranjith projects) block, repeatable: grants this user a set
  # of roles scoped to specific projects, further restricted to the resources matching
  # filter_ids (e.g. by tag).
  resource_role {
    project_ids = [data.spectrocloud_project.default.id, data.spectrocloud_project.ranjith.id]
    filter_ids  = [data.spectrocloud_filter.filter.id]
    role_ids    = [for r in data.spectrocloud_role.resource_roles : r.id]
  }

  # resource_role (resource_roles_editor, ranjith project) block, repeatable.
  resource_role {
    project_ids = [data.spectrocloud_project.ranjith.id]
    filter_ids  = [data.spectrocloud_filter.filter.id]
    role_ids    = [for re in data.spectrocloud_role.resource_roles_editor : re.id]
  }
}

# import existing user example
#import {
#  to = spectrocloud_user.test_user
#  id = "66fcb5fe19eb6dc880776d59"
#}

# To generate TF configuration.
#terraform plan -generate-config-out=test_user.tf

# To import State file
#terraform import spectrocloud_user.test_user 672c5ae21adfa1c28c9e37c9

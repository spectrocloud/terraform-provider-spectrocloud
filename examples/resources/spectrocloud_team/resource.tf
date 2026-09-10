# user
data "spectrocloud_user" "user1" {
  email = "nikolay@spectrocloud.com"
}

# role
data "spectrocloud_role" "project_role1" {
  name = "Project Editor"
}

data "spectrocloud_role" "project_role2" {
  name = "Cluster Admin"
}

data "spectrocloud_role" "project_role3" {
  name = "Project Admin"
}

# project
data "spectrocloud_project" "project1" {
  name = "Default"
}

data "spectrocloud_project" "project2" {
  name = "providence-004"
}

data "spectrocloud_role" "tenant_role4" {
  name = "Tenant Admin"
}

data "spectrocloud_workspace" "workspace1" {
  name = "wsp-tf"
}

data "spectrocloud_role" "workspace_role5" {
  name = "Workspace Admin"
}

# Day-2 mutability: only `name` is ForceNew - changing it recreates the team. `users`,
# `project_role_mapping`, `tenant_role_mapping`, and `workspace_role_mapping` all update in
# place.
resource "spectrocloud_team" "t1" {
  # Required, ForceNew. The team's name.
  name = "team1"
  # Optional. User IDs that are members of this team.
  users = [data.spectrocloud_user.user1.id]

  # Optional, repeatable. Grants this team a set of roles scoped to one project.
  project_role_mapping {
    id    = data.spectrocloud_project.project1.id
    roles = [data.spectrocloud_role.project_role1.id, data.spectrocloud_role.project_role2.id]
  }

  project_role_mapping {
    id    = data.spectrocloud_project.project2.id
    roles = [data.spectrocloud_role.project_role3.id]
  }

  # Optional. Tenant-scoped role IDs granted to this team.
  tenant_role_mapping = [data.spectrocloud_role.tenant_role4.id]

  # Optional, repeatable. Grants this team a set of roles scoped to one workspace within a
  # project - `id` here is the project ID (matching project_role_mapping.id above), and each
  # nested `workspace` block names a workspace in that project plus its roles.
  workspace_role_mapping {
    id = data.spectrocloud_project.project1.id
    workspace {
      id    = data.spectrocloud_workspace.workspace1.id
      roles = [data.spectrocloud_role.workspace_role5.id]
    }
  }
}

# terraform import spectrocloud_team.t1 "<team-uid>"
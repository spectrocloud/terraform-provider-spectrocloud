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

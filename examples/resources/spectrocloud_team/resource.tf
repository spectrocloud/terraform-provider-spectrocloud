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

/*data "spectrocloud_role" "workspace_role5" {
  name = "Workspace Admin"
}*/

resource "spectrocloud_team" "t1" {
  name  = "team1"
  users = [data.spectrocloud_user.user1.id]

  project_role_mapping {
    id    = data.spectrocloud_project.project1.id
    roles = [data.spectrocloud_role.project_role1.id, data.spectrocloud_role.project_role2.id]
  }

  project_role_mapping {
    id    = data.spectrocloud_project.project2.id
    roles = [data.spectrocloud_role.project_role3.id]
  }

  tenant_role_mapping = [data.spectrocloud_role.tenant_role4.id]

  /*workspace_role_mapping {
    id = data.spectrocloud_project.project1.id
    workspace {
      id = data.spectrocloud_workspace.workspace1.id
      roles = ["621b2d1b77a605d2edce05d9"]
    }
  }*/
}
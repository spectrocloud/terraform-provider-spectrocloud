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
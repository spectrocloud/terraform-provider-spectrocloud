resource "spectrocloud_application" "application" {
  name                    = "app-beru-whitesun-lars"
  application_profile_uid = "6345f67d784e4d30683b9987" #data.spectrocloud_application_profile.id

  config {
    cluster_name      = "sandbox-scorpius"
    cluster_group_uid = "6358d799fad5aa39fa26a8c2" # data.spectrocloud_cluster_group.id
    limits {
      cpu     = 3
      memory  = 4096
      storage = 3
    }
  }
}

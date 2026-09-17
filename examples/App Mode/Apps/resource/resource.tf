# This example deploys an application onto an existing virtual cluster. To target a cluster
# group instead (letting Palette place the application on any cluster in the group), set
# config.cluster_group_uid instead of config.cluster_uid.
#
# Attributes:
#   name                    - Required. Updatable in place - nothing on this resource is
#                             ForceNew.
#   tags                    - Optional. Free-form tags for organizing applications.
#   application_profile_uid - Required. The application profile that defines what gets deployed.
#
# config block (Optional, at most one):
#   cluster_uid       - Either cluster_uid or cluster_group_uid must be set (not both) - this
#                       targets a specific, already-existing virtual cluster.
#   cluster_group_uid - Use instead of cluster_uid to target a cluster group.
#   cluster_context   - Required. The cluster's context: "project" or "tenant".
#   cluster_name      - Optional. A display name for the target cluster.
#
# config.limits block (Optional. Resource limits for the application; omit any of
# cpu/memory/storage to leave that dimension unconstrained):
#   cpu     - CPU allocation (integer core count).
#   memory  - Memory allocation in megabytes.
#   storage - Storage allocation in gigabytes.

data "spectrocloud_application_profile" "profile" {
  name = "example-application-profile"
}

data "spectrocloud_cluster" "target" {
  name    = "example-virtual-cluster"
  context = "project"
}

resource "spectrocloud_application" "application" {
  name                    = "app-beru-whitesun-lars"
  tags                    = ["team:platform", "env:sandbox"]
  application_profile_uid = data.spectrocloud_application_profile.profile.id

  config {
    cluster_uid = data.spectrocloud_cluster.target.id
    # cluster_group_uid = "REPLACE_ME"

    cluster_context = "project"
    cluster_name    = "sandbox-scorpius"

    limits {
      cpu     = 3
      memory  = 4096
      storage = 3
    }
  }
}

# This example deploys an application onto an existing virtual cluster. To target a cluster
# group instead (letting Palette place the application on any cluster in the group), set
# config.cluster_group_uid instead of config.cluster_uid.

data "spectrocloud_application_profile" "profile" {
  name = "example-application-profile"
}

data "spectrocloud_cluster" "target" {
  name    = "example-virtual-cluster"
  context = "project"
}

resource "spectrocloud_application" "application" {
  # Required. Updatable in place - nothing on this resource is ForceNew.
  name = "app-beru-whitesun-lars"

  # Optional. Free-form tags for organizing applications.
  tags = ["team:platform", "env:sandbox"]

  # Required. The application profile that defines what gets deployed.
  application_profile_uid = data.spectrocloud_application_profile.profile.id

  # Optional, at most one config block.
  config {
    # Either cluster_uid or cluster_group_uid must be set (not both) - this targets a specific,
    # already-existing virtual cluster.
    cluster_uid = data.spectrocloud_cluster.target.id
    # cluster_group_uid = "REPLACE_ME" # Use instead of cluster_uid to target a cluster group.

    # Required. The cluster's context: "project" or "tenant".
    cluster_context = "project"

    # Optional. A display name for the target cluster.
    cluster_name = "sandbox-scorpius"

    # Optional. Resource limits for the application; omit any of cpu/memory/storage to leave
    # that dimension unconstrained.
    limits {
      cpu     = 3    # CPU allocation (integer core count).
      memory  = 4096 # Memory allocation in megabytes.
      storage = 3    # Storage allocation in gigabytes.
    }
  }
}

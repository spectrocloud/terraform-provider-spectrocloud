##########################################
# Scenario 1: Single Application
##########################################
resource "spectrocloud_application" "scenario-1" {
  name                    = "scenario-1"
  application_profile_uid = spectrocloud_application_profile.hello-universe-ui.id

  config {
    cluster_name    = spectrocloud_virtual_cluster.cluster-1.name
    cluster_uid     = spectrocloud_virtual_cluster.cluster-1.id
    cluster_context = "project"
  }
  tags = concat(var.tags, ["scenario-1"])

  timeouts {
    create = "10m"
    update = "10m"
  }
}

##########################################
# Scenario 2: Multiple Applications
##########################################
resource "spectrocloud_application" "scenario-2" {
  count = var.enable-second-scenario == true ? 1 : 0

  name                    = "scenario-2"
  application_profile_uid = spectrocloud_application_profile.hello-universe-complete[0].id

  config {
    cluster_name    = spectrocloud_virtual_cluster.cluster-2[0].name
    cluster_uid     = spectrocloud_virtual_cluster.cluster-2[0].id
    cluster_context = "project"
  }
  tags = concat(var.tags, ["scenario-2"])

  timeouts {
    create = "10m"
    update = "10m"
  }
}

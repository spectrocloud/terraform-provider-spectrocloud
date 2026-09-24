##########################################
# Scenario 1: Single Application
##########################################
resource "spectrocloud_virtual_cluster" "cluster-1" {
  name              = var.scenario-one-cluster-name
  cluster_group_uid = data.spectrocloud_cluster_group.cluster-group.id

  resources {
    max_cpu           = 4
    max_mem_in_mb     = 4096
    min_cpu           = 0
    min_mem_in_mb     = 0
    max_storage_in_gb = "2"
    min_storage_in_gb = "0"
  }

  tags = concat(var.tags, ["scenario-1"])

  timeouts {
    create = "15m"
    delete = "15m"
  }
}

##########################################
# Scenario 2: Multiple Applications
##########################################
resource "spectrocloud_virtual_cluster" "cluster-2" {

  count = var.enable-second-scenario == true ? 1 : 0

  name              = var.scenario-two-cluster-name
  cluster_group_uid = data.spectrocloud_cluster_group.cluster-group.id

  resources {
    max_cpu           = 8
    max_mem_in_mb     = 12288
    min_cpu           = 0
    min_mem_in_mb     = 0
    max_storage_in_gb = "12"
    min_storage_in_gb = "0"
  }

  tags = concat(var.tags, ["scenario-2"])

  timeouts {
    create = "15m"
    delete = "15m"
  }
}

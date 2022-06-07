resource "spectrocloud_cluster_aks" "aks" {
  name             = var.cluster1_name
  tags             = ["owner:dmitry"]
  cloud_account_id = data.spectrocloud_cloudaccount_azure.aks_account.id

  cloud_config {
    subscription_id = var.subscription_id
    resource_group  = var.resource_group
    ssh_key         = var.ssh_key
    region          = var.region
  }

  cluster_profile {
    id = spectrocloud_cluster_profile.infra_profile.id
  }

  cluster_profile {
    id = spectrocloud_cluster_profile.addon_profile.id
  }

/*  backup_policy {
    schedule                  = "0 0 * * SUN"
    backup_location_id        = data.spectrocloud_backup_storage_location.bsl.id
    prefix                    = "demo001-dmitry-backup"
    expiry_in_hour            = 7200
    include_disks             = true
    include_cluster_resources = true
  } */

  scan_policy {
    configuration_scan_schedule = "0 0 * * SUN"
    penetration_scan_schedule   = "0 0 * * SUN"
    conformance_scan_schedule   = "0 0 1 * *"
  }

  machine_pool {
    name                 = "system"
    count                = 1
    instance_type        = "Standard_D2s_v3"
    disk_size_gb         = 50
    is_system_node_pool  = true
    storage_account_type = "Premium_LRS"
  }  

  machine_pool {
    name                 = "application"
    count                = 1
    instance_type        = "Standard_D2s_v3"
    disk_size_gb         = 50
    is_system_node_pool  = false
    storage_account_type = "Premium_LRS"
  }

}

resource "spectrocloud_cluster_aks" "aks2" {
  name             = var.cluster2_name
  tags             = ["owner:dmitry"]
  cloud_account_id = data.spectrocloud_cloudaccount_azure.aks_account.id

  cloud_config {
    subscription_id = var.subscription_id
    resource_group  = var.resource_group
    ssh_key         = var.ssh_key
    region          = var.region
  }

  cluster_profile {
    id = spectrocloud_cluster_profile.infra_profile.id
  }

  cluster_profile {
    id = spectrocloud_cluster_profile.addon_profile.id
  }

/*  backup_policy {
    schedule                  = "0 0 * * SUN"
    backup_location_id        = data.spectrocloud_backup_storage_location.bsl.id
    prefix                    = "demo001-dmitry-backup"
    expiry_in_hour            = 7200
    include_disks             = true
    include_cluster_resources = true
  } */

  scan_policy {
    configuration_scan_schedule = "0 0 * * SUN"
    penetration_scan_schedule   = "0 0 * * SUN"
    conformance_scan_schedule   = "0 0 1 * *"
  }

  machine_pool {
    name                 = "system"
    count                = 1
    instance_type        = "Standard_D2s_v3"
    disk_size_gb         = 50
    is_system_node_pool  = true
    storage_account_type = "Premium_LRS"
  }  

  machine_pool {
    name                 = "application"
    count                = 1
    instance_type        = "Standard_D2s_v3"
    disk_size_gb         = 50
    is_system_node_pool  = false
    storage_account_type = "Premium_LRS"
  }

}

data "spectrocloud_cloudaccount_openstack" "account" {
  # id = <uid>
  name = var.cluster_cloud_account_name
}

data "spectrocloud_cluster_profile" "profile" {
  # id = <uid>
  name = var.cluster_cluster_profile_name
}

data "spectrocloud_backup_storage_location" "bsl" {
  name = var.backup_storage_location_name
}

resource "spectrocloud_cluster_openstack" "cluster" {
  name = "openstack-piyush-tf-1"

  cluster_profile {
    id = data.spectrocloud_cluster_profile.profile.id
  }

  cloud_account_id = data.spectrocloud_cloudaccount_openstack.account.id
  tags             = ["dev"]


  cloud_config {
    domain      = "Default"
    project     = "dev"
    region      = "RegionOne"
    ssh_key     = "Spectro2021"
    dns_servers = ["10.10.128.8", "8.8.8.8"]
    subnet_cidr = "192.168.151.0/24"
  }

  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "cp-pool"
    count                   = 1
    instance_type           = "spectro-xlarge"
    azs                     = ["zone1"]
  }

  machine_pool {
    name          = "worker-basic"
    count         = 2
    instance_type = "spectro-large"
    azs           = ["zone1"]
  }

  backup_policy {
    schedule                  = "0 0 * * SUN"
    backup_location_id        = data.spectrocloud_backup_storage_location.bsl.id
    prefix                    = "prod-backup"
    expiry_in_hour            = 7200
    include_disks             = true
    include_cluster_resources = true
  }

  scan_policy {
    configuration_scan_schedule = "0 0 * * SUN"
    penetration_scan_schedule   = "0 0 * * SUN"
    conformance_scan_schedule   = "0 0 1 * *"
  }
}
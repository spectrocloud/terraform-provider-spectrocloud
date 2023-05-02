data "spectrocloud_cluster_profile" "infra_profile" {
  name    = "ehs-default-infra"
  version = "1.0.12"
  context = "project"
}
data "spectrocloud_cluster_profile" "system_profile" {
  name    = "ehs-system-profile-dc"
  version = "1.0.80"
  context = "project"
}

resource "spectrocloud_cluster_libvirt" "libvirt_cluster" {
  name = "test-libvirt"
  tags = ["test:TF"]
  # infra profile
  cluster_profile {
    id = data.spectrocloud_cluster_profile.infra_profile.id
  }
  # system profile
  cluster_profile {
    id = data.spectrocloud_cluster_profile.system_profile.id
  }
  apply_setting   = "test-setting"
  skip_completion = true
  cloud_config {
    ssh_key               = "sss2022"
    vip                   = "12.23.12.21"
    network_search_domain = "dev.spectrocloud.com"
    network_type          = "VIP" # By default is VIP
  }
  machine_pool {
    name = "master-pool"
    additional_labels = {
      "type" : "master"
    }
    control_plane           = true
    control_plane_as_worker = true
    count                   = 2
    update_strategy         = "RollingUpdateScaleOut"
    instance_type {
      disk_size_gb = 10
      memory_mb    = 2048
      cpu          = 2
    }
    placements {
      appliance_id        = "tf-test-edge-master"
      network_type        = "default"
      network_names       = "tf-test-network"
      image_storage_pool  = "tf-test-storage-pool"
      target_storage_pool = "tf-test-target-storage-pool"
      data_storage_pool   = "tf-test-data-storage-pool"
    }
  }
  machine_pool {
    name = "worker-pool"
    additional_labels = {
      "type" : "worker"
    }
    control_plane           = true
    control_plane_as_worker = true
    count                   = 2
    update_strategy         = "RollingUpdateScaleOut"
    instance_type {
      disk_size_gb = 10
      memory_mb    = 2048
      cpu          = 2
    }
    placements {
      appliance_id        = "tf-test-edge-host"
      network_type        = "default"
      network_names       = "tf-test-network"
      image_storage_pool  = "tf-test-storage-pool"
      target_storage_pool = "tf-test-target-storage-pool"
      data_storage_pool   = "tf-test-data-storage-pool"
    }
  }
}
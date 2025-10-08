# Example showing how to use the new storage field in dataVolumeTemplates.spec
# This demonstrates the alternative to using the pvc field

data "spectrocloud_cluster" "vm_enabled_base_cluster" {
  name    = "tenant-cluster-002"
  context = "project"
}

locals {
  storage_class_name = "spectro-storage-class"
}

# VM using the new storage field instead of pvc field
resource "spectrocloud_virtual_machine" "tf-test-vm-with-storage" {
  cluster_uid     = data.spectrocloud_cluster.vm_enabled_base_cluster.id
  cluster_context = data.spectrocloud_cluster.vm_enabled_base_cluster.context
  run_on_launch   = true
  name            = "tf-test-vm-with-storage"
  namespace       = "default"
  labels = {
    "tf" = "test"
  }

  # Using the new storage field in dataVolumeTemplates.spec
  data_volume_templates {
    metadata {
      name      = "test-vm-bootvolume-storage"
      namespace = "default"
    }
    spec {
      source {
        registry {
          image_url = "gcr.io/spectro-images-public/release/vm-dashboard/os/ubuntu-container-disk:20.04"
        }
      }
      # Using the new storage field instead of pvc
      storage {
        access_modes = ["ReadWriteOnce"]
        resources {
          requests = {
            storage = "10Gi"
          }
          limits = {
            storage = "20Gi"
          }
        }
        storage_class_name = local.storage_class_name
        volume_mode        = "Filesystem"
        # Optional selector for label-based volume selection
        selector {
          match_labels = {
            type = "ssd"
            tier = "premium"
          }
        }
      }
    }
  }

  volume {
    name = "test-vm-datavolumedisk1"
    volume_source {
      data_volume {
        name = "test-vm-bootvolume-storage"
      }
    }
  }

  disk {
    name = "test-vm-datavolumedisk1"
    disk_device {
      disk {
        bus = "virtio"
      }
    }
  }

  cpu {
    cores   = 2
    sockets = 1
    threads = 1
  }

  memory {
    size = "4Gi"
  }

  interface {
    name                     = "default"
    interface_binding_method = "InterfaceBridge"
  }

  network {
    name           = "default"
    network_source = "pod"
  }
}

# Example showing both pvc and storage can coexist (for different volumes)
resource "spectrocloud_virtual_machine" "tf-test-vm-mixed-storage" {
  cluster_uid     = data.spectrocloud_cluster.vm_enabled_base_cluster.id
  cluster_context = data.spectrocloud_cluster.vm_enabled_base_cluster.context
  run_on_launch   = true
  name            = "tf-test-vm-mixed-storage"
  namespace       = "default"
  labels = {
    "tf" = "test"
  }

  # Boot volume using storage field
  data_volume_templates {
    metadata {
      name      = "boot-volume-storage"
      namespace = "default"
    }
    spec {
      source {
        registry {
          image_url = "gcr.io/spectro-images-public/release/vm-dashboard/os/ubuntu-container-disk:20.04"
        }
      }
      storage {
        access_modes = ["ReadWriteOnce"]
        resources {
          requests = {
            storage = "10Gi"
          }
        }
        storage_class_name = local.storage_class_name
      }
    }
  }

  # Data volume using traditional pvc field
  data_volume_templates {
    metadata {
      name      = "data-volume-pvc"
      namespace = "default"
    }
    spec {
      source {
        blank {}
      }
      pvc {
        access_modes = ["ReadWriteOnce"]
        resources {
          requests = {
            storage = "5Gi"
          }
        }
        storage_class_name = local.storage_class_name
      }
    }
  }

  volume {
    name = "boot-disk"
    volume_source {
      data_volume {
        name = "boot-volume-storage"
      }
    }
  }

  volume {
    name = "data-disk"
    volume_source {
      data_volume {
        name = "data-volume-pvc"
      }
    }
  }

  disk {
    name = "boot-disk"
    disk_device {
      disk {
        bus = "virtio"
      }
    }
  }

  disk {
    name = "data-disk"
    disk_device {
      disk {
        bus = "virtio"
      }
    }
  }

  cpu {
    cores   = 2
    sockets = 1
    threads = 1
  }

  memory {
    size = "4Gi"
  }

  interface {
    name                     = "default"
    interface_binding_method = "InterfaceBridge"
  }

  network {
    name           = "default"
    network_source = "pod"
  }
}

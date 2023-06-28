resource "spectrocloud_virtual_machine" "tf-test-vm-data-volume-template" {
  cluster_uid   = data.spectrocloud_cluster.vm_enabled_base_cluster.id
  run_on_launch = false
  # vm_action = "start" //["start", "stop", "pause", "resume", "migrate", "restart"]
  name      = "tf-test-vm-data-volume-template"
  namespace = "default"
  labels = {
    "key1" = "value1"
  }
  run_strategy = "Manual"
  data_volume_template {
    metadata {
      name      = "test-vm-bootvolume"
      namespace = "default"
    }
    spec {
      source {
        registry {
          image_url = "gcr.io/spectro-images-public/release/vm-dashboard/os/ubuntu-container-disk:20.04"
        }
        /*http {
          url = "https://cloud.centos.org/centos/7/images/CentOS-7-x86_64-GenericCloud.qcow2"
        }*/
      }
      pvc {
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


  disk {
    name = "test-vm-datavolumedisk1"
    disk_device {
      disk {
        bus = "virtio"
      }
    }
  }
  volume {
    name = "test-vm-datavolumedisk1"
    volume_source {
      data_volume {
        name = "test-vm-bootvolume"
      }
    }
  }

  disk {
    name = "test-vm-datavolumedisk2"
    disk_device {
      disk {
        bus = "virtio"
      }
    }
  }
  volume {
    name = "test-vm-datavolumedisk2"
    volume_source {
      data_volume {
        name = "test-vm-bootvolume"
      }
    }
  }

  ## potentially we can flatten cpu and memory type
  cpu {
    cores   = 2
    sockets = 2
    threads = 50
  }
  memory {
    guest = "16G"
  }

  ## leave as is as it's standard for k8s API.
  resources {
    requests = {
      memory = "8G"
      cpu    = 2
    }
    limits = {
      cpu    = 4
      memory = "20G"
    }
  }


  interface {
    name                     = "main"
    interface_binding_method = "InterfaceMasquerade"
  }
  interface {
    name                     = "additional"
    interface_binding_method = "InterfaceBridge"
  }

  network {
    name = "main"
    network_source {
      pod {}
    }
  }
  network {
    name = "additional"
    network_source {
      multus {
        network_name = "macvlan-conf"
        default      = false
      }
    }
  }
}
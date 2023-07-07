data "spectrocloud_cluster" "vm_enabled_base_cluster" {
  name    = "sh-maas-jun30"
  context = "project"
}



// Creating VM with Data Volume Templates -- Getting Error
#resource "spectrocloud_virtual_machine" "tf-test-vm-data-volume-template" {
#  cluster_uid   = data.spectrocloud_cluster.vm_enabled_base_cluster.id
#  run_on_launch = true
#  name      = "tf-test-vm-data-volume-template"
#  namespace = "default"
#  labels = {
#    "tf" = "test"
#  }
#  volume {
#    name = "containerdisk"
#    volume_source {
#      container_disk {
#        image_url = "gcr.io/spectro-images-public/release/vm-dashboard/os/ubuntu-container-disk:20.04"
#
#      }
#    }
#  }
#  volume {
#    name = "cloudintdisk"
#    volume_source {
#      cloud_init_config_drive {
#        user_data = "\n#cloud-config\nssh_pwauth: True\nchpasswd: { expire: False }\npassword: spectro\ndisable_root: false\n"
#      }
#    }
#  }
#
#  disk {
#    name = "containerdisk"
#    disk_device {
#      disk {
#        bus = "virtio"
#      }
#    }
#  }
#  disk {
#    name = "cloudintdisk"
#    disk_device {
#      disk {
#        bus = "virtio"
#      }
#    }
#  }
#
#  ## potentially we can flatten cpu and memory type
#  cpu {
#    cores   = 2
#    sockets = 1
#    threads = 10
#  }
#  memory {
#    guest = "1Gi"
#  }
#
#  ## leave as is as it's standard for k8s API.
#  resources {
#    requests = {
#      memory = "1Gi"
#      cpu    = 2
#    }
#    limits = {
#      cpu    = 2
#      memory = "1Gi"
#    }
#  }
#
#  interface {
#    name                     = "default"
#    interface_binding_method = "InterfaceMasquerade"
#  }
#
#
#  network {
#    name = "default"
#    network_source {
#      pod {}
#    }
#  }
#}



// Create a VM with default cloud init disk, container disk , multus network interface with interface binding method as sr-iov and network model
resource "spectrocloud_virtual_machine" "tf-test-vm-clone-default" {
  cluster_uid  = data.spectrocloud_cluster.vm_enabled_base_cluster.id
  base_vm_name = "ub20-3"
  name      = "tf-test-vm-data-volume-template"
  namespace = "default"
  labels = {
    "tf" = "test"
  }
  volume {
    name = "containerdisk"
    volume_source {
      container_disk {
        image_url = "gcr.io/spectro-images-public/release/vm-dashboard/os/ubuntu-container-disk:20.04"

      }
    }
  }
  volume {
    name = "cloudintdisk"
    volume_source {
      cloud_init_config_drive {
        user_data = "\n#cloud-config\nssh_pwauth: True\nchpasswd: { expire: False }\npassword: spectro\ndisable_root: false\n"
      }
    }
  }

  disk {
    name = "containerdisk"
    disk_device {
      disk {
        bus = "virtio"
      }
    }
  }
  disk {
    name = "cloudintdisk"
    disk_device {
      disk {
        bus = "virtio"
      }
    }
  }

  ## potentially we can flatten cpu and memory type
  cpu {
    cores   = 2
    sockets = 1
    threads = 10
  }
  memory {
    guest = "1Gi"
  }

  ## leave as is as it's standard for k8s API.
  resources {
    requests = {
      memory = "1Gi"
      cpu    = 2
    }
    limits = {
      cpu    = 2
      memory = "1Gi"
    }
  }

  interface {
    name                     = "default"
    interface_binding_method = "InterfaceMasquerade"
  }


  network {
    name = "default"
    network_source {
      pod {}
    }
  }

}
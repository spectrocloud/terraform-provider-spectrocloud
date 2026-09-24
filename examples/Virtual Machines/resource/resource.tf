# Day-2 mutability: `name`, `namespace`, `cluster_uid`, and `base_vm_name` are ForceNew -
# changing any of them recreates the virtual machine. `run_on_launch` and `run_strategy` are
# mutually exclusive - set only one. Everything else - labels, volume, disk, cpu, memory,
# resources, interface, network, data_volume_templates, and the Day-2 scheduling/affinity
# attributes shown further below - updates in place (KubeVirt itself may still require a VM
# restart to pick up some of these changes, but Terraform does not recreate the resource for
# them).
#
# Flat attributes on "tf-test-vm-basic-type" below:
#   cluster_uid     - Required, ForceNew. The cluster UID this VM belongs to.
#   cluster_context - Optional, default "project". Allowed: "project", "tenant".
#   run_on_launch   - Optional. The schema's own description says "Default value is `true`", but
#                     there is no actual Default set - ExactlyOneOf with `run_strategy` below
#                     means one of the two must always be set explicitly anyway (see the
#                     data-volume-template example further below, which uses
#                     run_strategy = "Manual" instead).
#   name            - Required, ForceNew.
#   namespace       - Optional, default "default", ForceNew.
data "spectrocloud_cluster" "vm_enabled_base_cluster" {
  name    = "tenant-cluster-002"
  context = "project"
}

// Create a VM with default cloud init disk, container disk, interface and network. Other usage
// scenarios (clone, data-volume-template, multi-network, Day-2 scheduling options) each live in
// their own resource-vm-*.tf1 file in this folder - see resource-vm-with-storage-spec.tf1 for why
// they use .tf1 instead of .tf.
resource "spectrocloud_virtual_machine" "tf-test-vm-basic-type" {
  cluster_uid     = data.spectrocloud_cluster.vm_enabled_base_cluster.id
  cluster_context = data.spectrocloud_cluster.vm_enabled_base_cluster.context
  run_on_launch   = true
  name            = "tf-test-vm-basic-type"
  namespace       = "default"
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

  cpu {
    cores   = 2
    sockets = 1
    threads = 10
  }
  memory {
    guest = "1Gi"
  }

  resources {
    requests = {
      memory = "1Gi"
      cpu    = 1
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

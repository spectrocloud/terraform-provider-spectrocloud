data "spectrocloud_cluster" "vm_enabled_base_cluster" {
  name    = "tenant-cluster-002"
  context = "project"
}

# Attaches a new hot-pluggable data volume to an existing running virtual machine
# (spectrocloud_virtual_machine), sourcing its disk image from an OCI registry.
#
# Day-2 mutability: the provider implements every change to this resource as delete-then-create
# under the hood (there is no true in-place update) - Terraform still shows this as an update in
# the plan rather than a replace, but any changed attribute recreates the underlying data volume
# on the cluster. Treat every attribute here as effectively ForceNew in practice.
resource "spectrocloud_datavolume" "example" {
  # Required. UID of the cluster the target VM runs on.
  cluster_uid = data.spectrocloud_cluster.vm_enabled_base_cluster.id
  # Required. Allowed: "project", "tenant".
  cluster_context = data.spectrocloud_cluster.vm_enabled_base_cluster.context
  # Required. Name and namespace of the existing virtual machine to attach this volume to.
  vm_name      = var.vm_name
  vm_namespace = var.vm_namespace

  # Required, at most one block. Describes how the new data volume is attached to the VM's own
  # spec (disk + volume_source), separate from the data volume's own storage spec below.
  add_volume_options {
    # Name of this volume attachment in the VM spec.
    name = "extra-datavolumedisk1"
    disk {
      name = "extra-datavolumedisk1"
      bus  = "virtio"
    }
    volume_source {
      data_volume {
        # Must match metadata.name below.
        name = "extra-datavolume"
        # Optional, default true. Whether this volume can be hot-plugged into a running VM.
        hotpluggable = true
      }
    }
  }

  # Required, at most one block. Kubernetes object metadata for the DataVolume itself.
  metadata {
    # Required, ForceNew.
    name = "extra-datavolume"
    # Optional, default "default", ForceNew.
    namespace = var.vm_namespace
  }

  # Required, at most one block. The DataVolume's own spec - where its disk image comes from
  # and how much storage to provision.
  spec {
    source {
      registry {
        image_url = "gcr.io/spectro-images-public/release/vm-dashboard/os/ubuntu-container-disk:20.04"
      }
      # Alternatives to registry (mutually exclusive - use only one):
      # http { url = "https://cloud.centos.org/centos/7/images/CentOS-7-x86_64-GenericCloud.qcow2" }
      # pvc  { name = "source-pvc"; namespace = "default" }
      # blank {}
    }

    storage {
      access_modes = ["ReadWriteOnce"]
      resources {
        requests = {
          storage = "5Gi"
        }
      }
      storage_class_name = "spectro-storage-class"
    }

    # Optional. Allowed: "kubevirt", "archive".
    # content_type = "kubevirt"
  }
}

# Day-2 mutability: `name`, `namespace`, `cluster_uid`, and `base_vm_name` are ForceNew -
# changing any of them recreates the virtual machine. `run_on_launch` and `run_strategy` are
# mutually exclusive - set only one. Everything else - labels, volume, disk, cpu, memory,
# resources, interface, network, data_volume_templates, and the Day-2 scheduling/affinity
# attributes shown further below - updates in place (KubeVirt itself may still require a VM
# restart to pick up some of these changes, but Terraform does not recreate the resource for
# them).
data "spectrocloud_cluster" "vm_enabled_base_cluster" {
  name    = "tenant-cluster-002"
  context = "project"
}

locals {
  storage_class_name = "spectro-storage-class"
}


// Create a VM with default cloud init disk, container disk , interface and network
#/*
resource "spectrocloud_virtual_machine" "tf-test-vm-basic-type" {
  # Required, ForceNew. Unique within the namespace.
  cluster_uid = data.spectrocloud_cluster.vm_enabled_base_cluster.id
  # Optional, default "project". Allowed: "project", "tenant".
  cluster_context = data.spectrocloud_cluster.vm_enabled_base_cluster.context
  # Optional, default true. Mutually exclusive with `run_strategy` (see the data-volume-template
  # example further below, which uses run_strategy = "Manual" instead).
  run_on_launch = true
  # Required, ForceNew.
  name = "tf-test-vm-basic-type"
  # Optional, default "default", ForceNew.
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
#*/

// Cloning VM with base_vm_name "tf-test-vm-basic-type"
/*
resource "spectrocloud_virtual_machine" "tf-test-vm-clone-default" {
  cluster_uid     = data.spectrocloud_cluster.vm_enabled_base_cluster.id
  cluster_context = data.spectrocloud_cluster.vm_enabled_base_cluster.context
  base_vm_name    = spectrocloud_virtual_machine.tf-test-vm-basic-type.name
  name            = "tf-test-vm-clone"
  namespace       = "default"
  run_on_launch   = true
  labels = {
    "tf" = "test"
  }
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
  // For Cloning VM's disk, volume, interface, cpu, memory and network or optional, Those configuration will get cloned from base vm.
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
*/

// Creating VM with data volume template
/*
resource "spectrocloud_virtual_machine" "tf-test-vm-data-volume-template" {
  cluster_uid     = data.spectrocloud_cluster.vm_enabled_base_cluster.id
  cluster_context = data.spectrocloud_cluster.vm_enabled_base_cluster.context
#  run_on_launch   = true
  run_strategy = "Manual"
  name            = "tf-test-vm-data-volume-template"
  namespace       = "default"
  labels = {
    "tf" = "test"
  }
  data_volume_templates {
    metadata {
      name      = "test-vm-bootvolume"
      namespace = "default"
    }
    spec {
      source {
        registry {
          image_url = "gcr.io/spectro-images-public/release/vm-dashboard/os/ubuntu-container-disk:20.04"
        }
        #http {
        #  url = "https://cloud.centos.org/centos/7/images/CentOS-7-x86_64-GenericCloud.qcow2"
        #}
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
    name = "test-vm-datavolumedisk1"
    volume_source {
      data_volume {
        name = "test-vm-bootvolume"
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
    threads = 10
  }
  memory {
    guest = "1G"
  }

  resources {
    requests = {
      memory = "1G"
      cpu    = 2
    }
    limits = {
      cpu    = 2
      memory = "1G"
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
*/

# Create a VM with default cloud init disk, container disk , multus network interface with interface binding method as sr-iov and network model
/*
resource "spectrocloud_virtual_machine" "tf-test-vm-multi-networks" {
  cluster_uid     = data.spectrocloud_cluster.vm_enabled_base_cluster.id
  cluster_context = data.spectrocloud_cluster.vm_enabled_base_cluster.context
  name            = "tf-test-vm-multi-network-interface"
  namespace       = "default"
  run_on_launch = true
  labels = {
    "key1" = "value1"
  }
  volume {
    name = "test-vm-containerdisk1"
    volume_source {
      container_disk {
        image_url = "quay.io/kubevirt/fedora-cloud-container-disk-demo"
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

  resources {
    requests = {
      memory = "2G"
      cpu    = 1
    }
  }

  disk {
    name = "test-vm-containerdisk1"
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

  interface {
    name                     = "main"
    interface_binding_method = "InterfaceMasquerade"
    model                    = "virtio"
  }
  interface {
    name                     = "additional"
    interface_binding_method = "InterfaceBridge"
    model                    = "e1000e"
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
*/

# Create a VM with default with all available day2 attributes
/*
resource "spectrocloud_virtual_machine" "tf-test-vm-all-option-template-spec" {
  cluster_uid     = data.spectrocloud_cluster.vm_enabled_base_cluster.id
  cluster_context = data.spectrocloud_cluster.vm_enabled_base_cluster.context
  name            = "tf-test-vm-all-option-spec-day2"
  namespace       = "default"
  run_on_launch = true
  labels = {
    "key1" = "value1"
  }
  #  Sample Day 2 Operation attributes
  #  priority_class_name = "high"
  #  scheduler_name = "test"
  #  node_selector = {
  #    "test_vmi" = "node_labels"
  #  }
  #  eviction_strategy = "LiveMigrate"
  #  termination_grace_period_seconds = 60
  #  hostname = "spectro-com"
  #  subdomain = "test-spectro-com"
  #  dns_policy = "Default" //["ClusterFirstWithHostNet", "ClusterFirst", "Default", "None"]
  #  tolerations {
  #    effect = "NoExecute" // ["NoSchedule", "PreferNoSchedule", "NoExecute"]
  #    key = "tolerationKey"
  #    operator = "Equal" // ["Exists", "Equal"]
  #    toleration_seconds = "60"
  #    value = "taintValue"
  #  }
  #  pod_dns_config {
  #    nameservers = ["10.0.0.10", "10.0.0.11"]
  #    option {
  #      name = "test_dns_name"
  #      value = "dns_value"
  #    }
  #    searches = ["policy1", "policy2"]
  #  }
  #  affinity {
  #    pod_anti_affinity {
  #      preferred_during_scheduling_ignored_during_execution {
  #        weight = 10
  #        pod_affinity_term {
  #          label_selector {
  #            match_labels = {
  #              anti-affinity-key = "anti-affinity-val"
  #            }
  #          }
  #          topology_key = "kubernetes.io/hostname"
  #        }
  #      }
  #    }
  #  }
  volume {
    name = "test-vm-containerdisk1"
    volume_source {
      container_disk {
        image_url = "quay.io/kubevirt/fedora-cloud-container-disk-demo"
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

  resources {
    requests = {
      memory = "2G"
      cpu    = 1
    }
    # Sample Day 2 Operation disk
    # limits = {
    #   "test_limit" = "10"
    # }
    over_commit_guest_overhead = false
  }

  disk {
    name = "test-vm-containerdisk1"
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
        bus       = "virtio"
        read_only = false
        # pci_address = "0000:03:07.0"
      }
    }
    serial = "1"
  }
  interface {
    name                     = "main"
    interface_binding_method = "InterfaceMasquerade" //["InterfaceBridge", "InterfaceSlirp", "InterfaceMasquerade","InterfaceSRIOV",]
    model                    = "virtio"
  }

  #  interface {
  #    name                     = "additional"
  #    interface_binding_method = "InterfaceBridge"
  #    model                    = "e1000e" // ["", "e1000", "e1000e", "ne2k_pci", "pcnet", "rtl8139", "virtio"]
  #  }

  network {
    name = "main"
    network_source {
      pod {}
    }
  }

  #  network {
  #    name = "additional"
  #    network_source {
  #      multus {
  #        network_name = "macvlan-conf"
  #        default      = false
  #      }
  #    }
  #  }


}
*/
data spectrocloud_cluster "vm_enabled_base_cluster" {
  name = "shruthi-aws-ugadi"
}
locals {
  # storage_class_name = "sumit-storage-class"
  storage_class_name = "spectro-storage-class-immediate"
}

// Creating VM with Data Volume Templates
resource "spectrocloud_virtual_machine" "tf-test-vm-data-volume-template" {
  cluster_uid = data.spectrocloud_cluster.vm_enabled_base_cluster.id
  run_on_launch = false
  metadata {
    name      = "tf-test-vm-data-volume-template"
    namespace = "default"
    labels = {
      "key1" = "value1"
    }
  }
  spec {
    run_strategy = "Manual"
    data_volume_templates {
      metadata {
        name      = "test-vm-bootvolume"
        namespace = "default"
      }
      spec {
        source {
          http {
            url = "https://cloud.centos.org/centos/7/images/CentOS-7-x86_64-GenericCloud.qcow2"
          }
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
    template {
      metadata {
        labels = {
          "kubevirt.io/vm" = "test-vm"
        }
      }
      spec {
        volume {
          name = "test-vm-datavolumedisk1"
          volume_source {
            data_volume {
              name = "test-vm-bootvolume"
            }
          }
        }
        domain {
          resources {
            requests = {
              memory = "8G"
              cpu    = 2
            }
          }
          devices {
            disk {
              name = "test-vm-datavolumedisk1"
              disk_device {
                disk {
                  bus = "virtio"
                }
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
          }
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
        /*affinity {
          pod_anti_affinity {
            preferred_during_scheduling_ignored_during_execution {
              weight = 100
              pod_affinity_term {
                label_selector {
                  match_labels = {
                    anti-affinity-key = "anti-affinity-val"
                  }
                }
                topology_key = "kubernetes.io/hostname"
              }
            }
          }
        }*/
      }
    }
  }
}

# Creates a VM with cloud init and contianer disk (with all default values)
resource "spectrocloud_virtual_machine" "tf-test-vm-default" {
  cluster_uid = data.spectrocloud_cluster.vm_enabled_base_cluster.id
  # run_on_launch = true
  vm_action = "stop"
  metadata {
    name      = "test-vm-default-container-disk"
    namespace = "default"
    labels = {
      "key1" = "value1"
    }
  }
  spec {
    template {
      metadata {
        labels = {
          "kubevirt.io/vm" = "test-vm-cont"
        }
      }
      spec {
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
        domain {
          resources {
            requests = {
              memory = "1G"
              cpu    = 1
            }
          }
          devices {
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
            }
          }
        }
        network {
          name = "main"
          network_source {
            pod {}
          }
        }
      }
    }
  }
}

// Creating VM by cloning existing VM
resource "spectrocloud_virtual_machine" "tf-test-vm-clone-default" {
  cluster_uid = data.spectrocloud_cluster.vm_enabled_base_cluster.id
  base_vm_name = spectrocloud_virtual_machine.tf-test-vm-default.metadata.0.name
  metadata {
    name      = "tf-test-vm-clone-default"
    namespace = "default"
    labels = {
      "key1" = "value1"
    }
  }
}

# Create a VM with default cloud init disk, container disk , multus network interface with interface binding method as sr-iov and network model
resource "spectrocloud_virtual_machine" "tf-test-vm-multinetwork" {
  cluster_uid = data.spectrocloud_cluster.vm_enabled_base_cluster.id
  # run_on_launch = true
  # vm_action = "stop"
  metadata {
    name      = "tf-test-vm-multi-network-interface"
    namespace = "default"
    labels = {
      "key1" = "value1"
    }
  }
  spec {
    template {
      metadata {
        labels = {
          "kubevirt.io/vm" = "test-vm-cont"
        }
      }
      spec {
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
        domain {
          resources {
            requests = {
              memory = "2G"
              cpu    = 1
            }
          }
          devices {
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
              model                   = "virtio"
            }
            interface {
              name                     = "additional"
              interface_binding_method = "InterfaceBridge"
              model                   = "e1000e"
            }
          }
        }
        network {
          name = "main"
          network_source {
            pod {}
          }
        }
        network {
          name                     = "additional"
          network_source {
            multus {
              network_name = "macvlan-conf"
              default      = false
            }
          }
        }
      }
    }
  }
}


# Create a VM with default with all available day2 attributes
resource "spectrocloud_virtual_machine" "tf-test-vm-all-option-template-spec" {
  cluster_uid = data.spectrocloud_cluster.vm_enabled_base_cluster.id
  metadata {
    name      = "tf-test-vm-all-option-template-spec"
    namespace = "default"
    labels = {
      "key1" = "value1"
    }
  }
  spec {
    template {
      metadata {
        labels = {
          "kubevirt.io/vm" = "test-vm-cont"
        }
      }
      spec {
        # Sample Day 2 Operation disk
        /*
        priority_class_name = "high"
        scheduler_name = "test"
        node_selector = {
          "test_vmi" = "node_labels"
        }
        eviction_strategy = "LiveMigrate"
        termination_grace_period_seconds = 60
        hostname = "spectro-com"
        subdomain = "test-spectro-com"
        dns_policy = "Default" //["ClusterFirstWithHostNet", "ClusterFirst", "Default", "None"]
        tolerations {
          effect = "NoExecute" // ["NoSchedule", "PreferNoSchedule", "NoExecute"]
          key = "tolerationKey"
          operator = "Equal" // ["Exists", "Equal"]
          toleration_seconds = "60"
          value = "taintValue"
        }
        pod_dns_config {
          nameservers = ["10.0.0.10", "10.0.0.11"]
          option {
            name = "test_dns_name"
            value = "dns_value"
          }
          searches = ["policy1", "policy2"]
        }
        affinity {
          pod_anti_affinity {
            preferred_during_scheduling_ignored_during_execution {
              weight = 10
              pod_affinity_term {
                label_selector {
                  match_labels = {
                    anti-affinity-key = "anti-affinity-val"
                  }
                }
                topology_key = "kubernetes.io/hostname"
              }
            }
          }
        }
        */
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
        domain {
          resources {
            requests = {
              memory = "2G"
              cpu    = 1
            }
            # Sample Day 2 Operation disk
            /*
            limits = {
              "test_limit" = "10"
            }
            */
            over_commit_guest_overhead = false
          }
          devices {
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
                  read_only = false
                  # pci_address = "0000:03:07.0"
                }
              }
              serial = "1"
            }
            interface {
              name                     = "main"
              interface_binding_method = "InterfaceMasquerade" //["InterfaceBridge", "InterfaceSlirp", "InterfaceMasquerade","InterfaceSRIOV",]
              model                   = "virtio"
            }
            interface {
              name                     = "additional"
              interface_binding_method = "InterfaceBridge"
              model                   = "e1000e" // ["", "e1000", "e1000e", "ne2k_pci", "pcnet", "rtl8139", "virtio"]
            }
          }
        }
        network {
          name = "main"
          network_source {
            pod {}
          }
        }
        network {
          name                     = "additional"
          network_source {
            multus {
              network_name = "macvlan-conf"
              default      = false
            }
          }
        }
      }
    }
  }
}

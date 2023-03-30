resource "spectrocloud_virtual_machine" "virtual_machine" {
  cluster_uid = "6414899fa4e47d6788678ecf"
  run_on_launch = false
  metadata {
    name      = "test-vm"
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
          storage_class_name = "sumit-storage-class"
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
        /*       affinity {
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

// Creating VM by cloning existing VM
resource "spectrocloud_virtual_machine" "tf-test-vm-clone-default" {
  cluster_uid  = "6414899fa4e47d6788678ecf"
  base_vm_name = spectrocloud_virtual_machine.virtual_machine.metadata.0.name
  metadata {
    name      = "tf-test-vm-clone-default"
    namespace = "default"
    labels = {
      "key1" = "value1"
    }
  }
}


#Creates a VM with cloud init and contianer disk
resource "spectrocloud_virtual_machine" "tf-test-vm-default" {
  cluster_uid = "6419c4e33964c35b04e62656"
  #run_on_launch = true
  vm_action = "stop"
  metadata {
    name      = "test-vm"
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
              memory = "8G"
              cpu    = 2
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


# Create a VM with default cloud init disk, container disk and multus network interface with interface binding method as sr-iov
resource "spectrocloud_virtual_machine" "tf-test-vm-multinetwork" {
  cluster_uid = "6419c4e33964c35b04e62656"
  #run_on_launch = true
  vm_action = "stop"
  metadata {
    name      = "test-vm-ni"
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
              memory = "8G"
              cpu    = 2
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
              interface_binding_method = "InterfaceSRIOV"
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
---
page_title: "spectrocloud_virtual_machine Resource - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  
---

# spectrocloud_virtual_machine (Resource)

  

## Example Usage

```terraform
data "spectrocloud_cluster" "vm_enabled_base_cluster" {
  name    = "milwaukee"
  context = "tenant"
}
locals {
  storage_class_name = "spectro-storage-class-immediate"
}

// Creating VM with Data Volume Templates
resource "spectrocloud_virtual_machine" "tf-test-vm-data-volume-template" {
  cluster_uid   = data.spectrocloud_cluster.vm_enabled_base_cluster.id
  run_on_launch = false
  # vm_action = "start" //["start", "stop", "pause", "resume", "migrate", "restart"]
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
          cpu {
            cores   = 2
            sockets = 2
            threads = 50
          }
          memory {
            guest = "16G"
          }
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
  cluster_uid  = data.spectrocloud_cluster.vm_enabled_base_cluster.id
  base_vm_name = spectrocloud_virtual_machine.tf-test-vm-default.metadata.0.name
  metadata {
    name      = "tf-test-vm-clone-default"
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

# Create a VM with default cloud init disk, container disk , multus network interface with interface binding method as sr-iov and network model
resource "spectrocloud_virtual_machine" "tf-test-vm-multinetwork" {
  cluster_uid = data.spectrocloud_cluster.vm_enabled_base_cluster.id
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
              model                    = "virtio"
            }
            interface {
              name                     = "additional"
              interface_binding_method = "InterfaceBridge"
              model                    = "e1000e"
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
            interface {
              name                     = "additional"
              interface_binding_method = "InterfaceBridge"
              model                    = "e1000e" // ["", "e1000", "e1000e", "ne2k_pci", "pcnet", "rtl8139", "virtio"]
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
      }
    }
  }
}
```


<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `cluster_uid` (String) The cluster UID to which the virtual machine belongs to.
- `metadata` (Block List, Min: 1, Max: 1) Standard VirtualMachine's metadata. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata (see [below for nested schema](#nestedblock--metadata))

### Optional

- `base_vm_name` (String) The name of the source virtual machine that a clone will be created of.
- `run_on_launch` (Boolean) If set to `true`, the virtual machine will be started when the cluster is launched. Default value is `true`.
- `spec` (Block List, Max: 1) VirtualMachineSpec describes how the proper VirtualMachine should look like. (see [below for nested schema](#nestedblock--spec))
- `status` (Block List, Max: 1) VirtualMachineStatus represents the status returned by the controller to describe how the VirtualMachine is doing. (see [below for nested schema](#nestedblock--status))
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `vm_action` (String) The action to be performed on the virtual machine. Valid values are: `start`, `stop`, `restart`, `pause`, `resume`, `migrate`. Default value is `start`.

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--metadata"></a>
### Nested Schema for `metadata`

Optional:

- `annotations` (Map of String) An unstructured key value map stored with the VirtualMachine that may be used to store arbitrary metadata. More info: http://kubernetes.io/docs/user-guide/annotations
- `labels` (Map of String) Map of string keys and values that can be used to organize and categorize (scope and select) the VirtualMachine. May match selectors of replication controllers and services. More info: http://kubernetes.io/docs/user-guide/labels
- `name` (String) Name of the VirtualMachine, must be unique. Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names
- `namespace` (String) Namespace defines the space within which name of the VirtualMachine must be unique.

Read-Only:

- `generation` (Number) A sequence number representing a specific generation of the desired state.
- `resource_version` (String) An opaque value that represents the internal version of this VirtualMachine that can be used by clients to determine when VirtualMachine has changed. Read more: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency
- `self_link` (String) A URL representing this VirtualMachine.
- `uid` (String) The unique in time and space value for this VirtualMachine. More info: http://kubernetes.io/docs/user-guide/identifiers#uids


<a id="nestedblock--spec"></a>
### Nested Schema for `spec`

Optional:

- `data_volume_templates` (Block List) dataVolumeTemplates is a list of dataVolumes that the VirtualMachineInstance template can reference. (see [below for nested schema](#nestedblock--spec--data_volume_templates))
- `run_strategy` (String) Running state indicates the requested running state of the VirtualMachineInstance, mutually exclusive with Running.
- `template` (Block List, Max: 1) Template is the direct specification of VirtualMachineInstance. (see [below for nested schema](#nestedblock--spec--template))

<a id="nestedblock--spec--data_volume_templates"></a>
### Nested Schema for `spec.data_volume_templates`

Required:

- `metadata` (Block List, Min: 1, Max: 1) Standard DataVolume's metadata. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata (see [below for nested schema](#nestedblock--spec--data_volume_templates--metadata))
- `spec` (Block List, Min: 1, Max: 1) DataVolumeSpec defines our specification for a DataVolume type (see [below for nested schema](#nestedblock--spec--data_volume_templates--spec))

<a id="nestedblock--spec--data_volume_templates--metadata"></a>
### Nested Schema for `spec.data_volume_templates.metadata`

Optional:

- `annotations` (Map of String) An unstructured key value map stored with the DataVolume that may be used to store arbitrary metadata. More info: http://kubernetes.io/docs/user-guide/annotations
- `labels` (Map of String) Map of string keys and values that can be used to organize and categorize (scope and select) the DataVolume. May match selectors of replication controllers and services. More info: http://kubernetes.io/docs/user-guide/labels
- `name` (String) Name of the DataVolume, must be unique. Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names
- `namespace` (String) Namespace defines the space within which name of the DataVolume must be unique.

Read-Only:

- `generation` (Number) A sequence number representing a specific generation of the desired state.
- `resource_version` (String) An opaque value that represents the internal version of this DataVolume that can be used by clients to determine when DataVolume has changed. Read more: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency
- `self_link` (String) A URL representing this DataVolume.
- `uid` (String) The unique in time and space value for this DataVolume. More info: http://kubernetes.io/docs/user-guide/identifiers#uids


<a id="nestedblock--spec--data_volume_templates--spec"></a>
### Nested Schema for `spec.data_volume_templates.spec`

Required:

- `pvc` (Block List, Min: 1, Max: 1) PVC is a pointer to the PVC Spec we want to use. (see [below for nested schema](#nestedblock--spec--data_volume_templates--spec--pvc))

Optional:

- `content_type` (String) ContentType options: "kubevirt", "archive".
- `source` (Block List, Max: 1) Source is the src of the data for the requested DataVolume. (see [below for nested schema](#nestedblock--spec--data_volume_templates--spec--source))

<a id="nestedblock--spec--data_volume_templates--spec--pvc"></a>
### Nested Schema for `spec.data_volume_templates.spec.pvc`

Required:

- `access_modes` (Set of String) A set of the desired access modes the volume should have. More info: http://kubernetes.io/docs/user-guide/persistent-volumes#access-modes-1
- `resources` (Block List, Min: 1, Max: 1) A list of the minimum resources the volume should have. More info: http://kubernetes.io/docs/user-guide/persistent-volumes#resources (see [below for nested schema](#nestedblock--spec--data_volume_templates--spec--pvc--resources))

Optional:

- `selector` (Block List, Max: 1) A label query over volumes to consider for binding. (see [below for nested schema](#nestedblock--spec--data_volume_templates--spec--pvc--selector))
- `storage_class_name` (String) Name of the storage class requested by the claim
- `volume_mode` (String) volumeMode defines what type of volume is required by the claim. Value of Filesystem is implied when not included in claim spec.
- `volume_name` (String) The binding reference to the PersistentVolume backing this claim.

<a id="nestedblock--spec--data_volume_templates--spec--pvc--resources"></a>
### Nested Schema for `spec.data_volume_templates.spec.pvc.resources`

Optional:

- `limits` (Map of String) Map describing the maximum amount of compute resources allowed. More info: http://kubernetes.io/docs/user-guide/compute-resources/
- `requests` (Map of String) Map describing the minimum amount of compute resources required. If this is omitted for a container, it defaults to `limits` if that is explicitly specified, otherwise to an implementation-defined value. More info: http://kubernetes.io/docs/user-guide/compute-resources/


<a id="nestedblock--spec--data_volume_templates--spec--pvc--selector"></a>
### Nested Schema for `spec.data_volume_templates.spec.pvc.selector`

Optional:

- `match_expressions` (Block List) A list of label selector requirements. The requirements are ANDed. (see [below for nested schema](#nestedblock--spec--data_volume_templates--spec--pvc--selector--match_expressions))
- `match_labels` (Map of String) A map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of `match_expressions`, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.

<a id="nestedblock--spec--data_volume_templates--spec--pvc--selector--match_expressions"></a>
### Nested Schema for `spec.data_volume_templates.spec.pvc.selector.match_expressions`

Optional:

- `key` (String) The label key that the selector applies to.
- `operator` (String) A key's relationship to a set of values. Valid operators ard `In`, `NotIn`, `Exists` and `DoesNotExist`.
- `values` (Set of String) An array of string values. If the operator is `In` or `NotIn`, the values array must be non-empty. If the operator is `Exists` or `DoesNotExist`, the values array must be empty. This array is replaced during a strategic merge patch.




<a id="nestedblock--spec--data_volume_templates--spec--source"></a>
### Nested Schema for `spec.data_volume_templates.spec.source`

Optional:

- `blank` (Block List, Max: 1) DataVolumeSourceBlank provides the parameters to create a Data Volume from an empty source. (see [below for nested schema](#nestedblock--spec--data_volume_templates--spec--source--blank))
- `http` (Block List, Max: 1) DataVolumeSourceHTTP provides the parameters to create a Data Volume from an HTTP source. (see [below for nested schema](#nestedblock--spec--data_volume_templates--spec--source--http))
- `pvc` (Block List, Max: 1) DataVolumeSourcePVC provides the parameters to create a Data Volume from an existing PVC. (see [below for nested schema](#nestedblock--spec--data_volume_templates--spec--source--pvc))
- `registry` (Block List, Max: 1) DataVolumeSourceRegistry provides the parameters to create a Data Volume from an existing PVC. (see [below for nested schema](#nestedblock--spec--data_volume_templates--spec--source--registry))

<a id="nestedblock--spec--data_volume_templates--spec--source--blank"></a>
### Nested Schema for `spec.data_volume_templates.spec.source.blank`


<a id="nestedblock--spec--data_volume_templates--spec--source--http"></a>
### Nested Schema for `spec.data_volume_templates.spec.source.http`

Optional:

- `cert_config_map` (String) Cert_config_map provides a reference to the Registry certs.
- `secret_ref` (String) Secret_ref provides the secret reference needed to access the HTTP source.
- `url` (String) url is the URL of the http source.


<a id="nestedblock--spec--data_volume_templates--spec--source--pvc"></a>
### Nested Schema for `spec.data_volume_templates.spec.source.pvc`

Optional:

- `name` (String) The name of the PVC.
- `namespace` (String) The namespace which the PVC located in.


<a id="nestedblock--spec--data_volume_templates--spec--source--registry"></a>
### Nested Schema for `spec.data_volume_templates.spec.source.registry`

Optional:

- `image_url` (String) The registry URL of the image to download.





<a id="nestedblock--spec--template"></a>
### Nested Schema for `spec.template`

Required:

- `metadata` (Block List, Min: 1, Max: 1) Standard VirtualMachineInstanceTemplateSpec's metadata. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata (see [below for nested schema](#nestedblock--spec--template--metadata))

Optional:

- `spec` (Block List, Max: 1) Template is the direct specification of VirtualMachineInstance. (see [below for nested schema](#nestedblock--spec--template--spec))

<a id="nestedblock--spec--template--metadata"></a>
### Nested Schema for `spec.template.metadata`

Optional:

- `annotations` (Map of String) An unstructured key value map stored with the VirtualMachineInstanceTemplateSpec that may be used to store arbitrary metadata. More info: http://kubernetes.io/docs/user-guide/annotations
- `labels` (Map of String) Map of string keys and values that can be used to organize and categorize (scope and select) the VirtualMachineInstanceTemplateSpec. May match selectors of replication controllers and services. More info: http://kubernetes.io/docs/user-guide/labels
- `name` (String) Name of the VirtualMachineInstanceTemplateSpec, must be unique. Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names
- `namespace` (String) Namespace defines the space within which name of the VirtualMachineInstanceTemplateSpec must be unique.

Read-Only:

- `generation` (Number) A sequence number representing a specific generation of the desired state.
- `resource_version` (String) An opaque value that represents the internal version of this VirtualMachineInstanceTemplateSpec that can be used by clients to determine when VirtualMachineInstanceTemplateSpec has changed. Read more: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency
- `self_link` (String) A URL representing this VirtualMachineInstanceTemplateSpec.
- `uid` (String) The unique in time and space value for this VirtualMachineInstanceTemplateSpec. More info: http://kubernetes.io/docs/user-guide/identifiers#uids


<a id="nestedblock--spec--template--spec"></a>
### Nested Schema for `spec.template.spec`

Optional:

- `affinity` (Block List, Max: 1) Optional pod scheduling constraints. (see [below for nested schema](#nestedblock--spec--template--spec--affinity))
- `dns_policy` (String) DNSPolicy defines how a pod's DNS will be configured.
- `domain` (Block List, Max: 1) Specification of the desired behavior of the VirtualMachineInstance on the host. (see [below for nested schema](#nestedblock--spec--template--spec--domain))
- `eviction_strategy` (String) EvictionStrategy can be set to "LiveMigrate" if the VirtualMachineInstance should be migrated instead of shut-off in case of a node drain.
- `hostname` (String) Specifies the hostname of the vmi.
- `liveness_probe` (Block List, Max: 1) Specification of the desired behavior of the VirtualMachineInstance on the host. (see [below for nested schema](#nestedblock--spec--template--spec--liveness_probe))
- `network` (Block List) List of networks that can be attached to a vm's virtual interface. (see [below for nested schema](#nestedblock--spec--template--spec--network))
- `node_selector` (Map of String) NodeSelector is a selector which must be true for the vmi to fit on a node. Selector which must match a node's labels for the vmi to be scheduled on that node.
- `pod_dns_config` (Block List, Max: 1) Specifies the DNS parameters of a pod. Parameters specified here will be merged to the generated DNS configuration based on DNSPolicy. Optional: Defaults to empty (see [below for nested schema](#nestedblock--spec--template--spec--pod_dns_config))
- `priority_class_name` (String) If specified, indicates the pod's priority. If not specified, the pod priority will be default or zero if there is no default.
- `readiness_probe` (Block List, Max: 1) Specification of the desired behavior of the VirtualMachineInstance on the host. (see [below for nested schema](#nestedblock--spec--template--spec--readiness_probe))
- `scheduler_name` (String) If specified, the VMI will be dispatched by specified scheduler. If not specified, the VMI will be dispatched by default scheduler.
- `subdomain` (String) If specified, the fully qualified vmi hostname will be "<hostname>.<subdomain>.<pod namespace>.svc.<cluster domain>".
- `termination_grace_period_seconds` (Number) Grace period observed after signalling a VirtualMachineInstance to stop after which the VirtualMachineInstance is force terminated.
- `tolerations` (Block List) If specified, the pod's toleration. Optional: Defaults to empty (see [below for nested schema](#nestedblock--spec--template--spec--tolerations))
- `volume` (Block List) Specification of the desired behavior of the VirtualMachineInstance on the host. (see [below for nested schema](#nestedblock--spec--template--spec--volume))

<a id="nestedblock--spec--template--spec--affinity"></a>
### Nested Schema for `spec.template.spec.affinity`

Optional:

- `node_affinity` (Block List, Max: 1) Node affinity scheduling rules for the pod. (see [below for nested schema](#nestedblock--spec--template--spec--affinity--node_affinity))
- `pod_affinity` (Block List, Max: 1) Inter-pod topological affinity. rules that specify that certain pods should be placed in the same topological domain (e.g. same node, same rack, same zone, same power domain, etc.) (see [below for nested schema](#nestedblock--spec--template--spec--affinity--pod_affinity))
- `pod_anti_affinity` (Block List, Max: 1) Inter-pod topological affinity. rules that specify that certain pods should be placed in the same topological domain (e.g. same node, same rack, same zone, same power domain, etc.) (see [below for nested schema](#nestedblock--spec--template--spec--affinity--pod_anti_affinity))

<a id="nestedblock--spec--template--spec--affinity--node_affinity"></a>
### Nested Schema for `spec.template.spec.affinity.node_affinity`

Optional:

- `preferred_during_scheduling_ignored_during_execution` (Block List) The scheduler will prefer to schedule pods to nodes that satisfy the affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, RequiredDuringScheduling affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding 'weight' to the sum if the node matches the corresponding MatchExpressions; the node(s) with the highest sum are the most preferred. (see [below for nested schema](#nestedblock--spec--template--spec--affinity--node_affinity--preferred_during_scheduling_ignored_during_execution))
- `required_during_scheduling_ignored_during_execution` (Block List, Max: 1) If the affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to a node label update), the system may or may not try to eventually evict the pod from its node. (see [below for nested schema](#nestedblock--spec--template--spec--affinity--node_affinity--required_during_scheduling_ignored_during_execution))

<a id="nestedblock--spec--template--spec--affinity--node_affinity--preferred_during_scheduling_ignored_during_execution"></a>
### Nested Schema for `spec.template.spec.affinity.node_affinity.preferred_during_scheduling_ignored_during_execution`

Required:

- `preference` (Block List, Min: 1, Max: 1) A node selector term, associated with the corresponding weight. (see [below for nested schema](#nestedblock--spec--template--spec--affinity--node_affinity--preferred_during_scheduling_ignored_during_execution--preference))
- `weight` (Number) weight is in the range 1-100

<a id="nestedblock--spec--template--spec--affinity--node_affinity--preferred_during_scheduling_ignored_during_execution--preference"></a>
### Nested Schema for `spec.template.spec.affinity.node_affinity.preferred_during_scheduling_ignored_during_execution.preference`

Optional:

- `match_expressions` (Block List) List of node selector requirements. The requirements are ANDed. (see [below for nested schema](#nestedblock--spec--template--spec--affinity--node_affinity--preferred_during_scheduling_ignored_during_execution--preference--match_expressions))

<a id="nestedblock--spec--template--spec--affinity--node_affinity--preferred_during_scheduling_ignored_during_execution--preference--match_expressions"></a>
### Nested Schema for `spec.template.spec.affinity.node_affinity.preferred_during_scheduling_ignored_during_execution.preference.match_expressions`

Optional:

- `key` (String) The label key that the selector applies to.
- `operator` (String) Operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.
- `values` (Set of String) Values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be interpreted as an integer. This array is replaced during a strategic merge patch.




<a id="nestedblock--spec--template--spec--affinity--node_affinity--required_during_scheduling_ignored_during_execution"></a>
### Nested Schema for `spec.template.spec.affinity.node_affinity.required_during_scheduling_ignored_during_execution`

Optional:

- `node_selector_term` (Block List) List of node selector terms. The terms are ORed. (see [below for nested schema](#nestedblock--spec--template--spec--affinity--node_affinity--required_during_scheduling_ignored_during_execution--node_selector_term))

<a id="nestedblock--spec--template--spec--affinity--node_affinity--required_during_scheduling_ignored_during_execution--node_selector_term"></a>
### Nested Schema for `spec.template.spec.affinity.node_affinity.required_during_scheduling_ignored_during_execution.node_selector_term`

Optional:

- `match_expressions` (Block List) List of node selector requirements. The requirements are ANDed. (see [below for nested schema](#nestedblock--spec--template--spec--affinity--node_affinity--required_during_scheduling_ignored_during_execution--node_selector_term--match_expressions))

<a id="nestedblock--spec--template--spec--affinity--node_affinity--required_during_scheduling_ignored_during_execution--node_selector_term--match_expressions"></a>
### Nested Schema for `spec.template.spec.affinity.node_affinity.required_during_scheduling_ignored_during_execution.node_selector_term.match_expressions`

Optional:

- `key` (String) The label key that the selector applies to.
- `operator` (String) Operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.
- `values` (Set of String) Values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be interpreted as an integer. This array is replaced during a strategic merge patch.





<a id="nestedblock--spec--template--spec--affinity--pod_affinity"></a>
### Nested Schema for `spec.template.spec.affinity.pod_affinity`

Optional:

- `preferred_during_scheduling_ignored_during_execution` (Block List) The scheduler will prefer to schedule pods to nodes that satisfy the anti-affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, RequiredDuringScheduling anti-affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding 'weight' to the sum if the node matches the corresponding MatchExpressions; the node(s) with the highest sum are the most preferred. (see [below for nested schema](#nestedblock--spec--template--spec--affinity--pod_affinity--preferred_during_scheduling_ignored_during_execution))
- `required_during_scheduling_ignored_during_execution` (Block List) If the affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to a pod label update), the system may or may not try to eventually evict the pod from its node. When there are multiple elements, the lists of nodes corresponding to each PodAffinityTerm are intersected, i.e. all terms must be satisfied. (see [below for nested schema](#nestedblock--spec--template--spec--affinity--pod_affinity--required_during_scheduling_ignored_during_execution))

<a id="nestedblock--spec--template--spec--affinity--pod_affinity--preferred_during_scheduling_ignored_during_execution"></a>
### Nested Schema for `spec.template.spec.affinity.pod_affinity.preferred_during_scheduling_ignored_during_execution`

Required:

- `pod_affinity_term` (Block List, Min: 1, Max: 1) A pod affinity term, associated with the corresponding weight (see [below for nested schema](#nestedblock--spec--template--spec--affinity--pod_affinity--preferred_during_scheduling_ignored_during_execution--pod_affinity_term))
- `weight` (Number) weight associated with matching the corresponding podAffinityTerm, in the range 1-100

<a id="nestedblock--spec--template--spec--affinity--pod_affinity--preferred_during_scheduling_ignored_during_execution--pod_affinity_term"></a>
### Nested Schema for `spec.template.spec.affinity.pod_affinity.preferred_during_scheduling_ignored_during_execution.pod_affinity_term`

Optional:

- `label_selector` (Block List) A label query over a set of resources, in this case pods. (see [below for nested schema](#nestedblock--spec--template--spec--affinity--pod_affinity--preferred_during_scheduling_ignored_during_execution--pod_affinity_term--label_selector))
- `namespaces` (Set of String) namespaces specifies which namespaces the labelSelector applies to (matches against); null or empty list means 'this pod's namespace'
- `topology_key` (String) empty topology key is interpreted by the scheduler as 'all topologies'

<a id="nestedblock--spec--template--spec--affinity--pod_affinity--preferred_during_scheduling_ignored_during_execution--pod_affinity_term--label_selector"></a>
### Nested Schema for `spec.template.spec.affinity.pod_affinity.preferred_during_scheduling_ignored_during_execution.pod_affinity_term.label_selector`

Optional:

- `match_expressions` (Block List) A list of label selector requirements. The requirements are ANDed. (see [below for nested schema](#nestedblock--spec--template--spec--affinity--pod_affinity--preferred_during_scheduling_ignored_during_execution--pod_affinity_term--label_selector--match_expressions))
- `match_labels` (Map of String) A map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of `match_expressions`, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.

<a id="nestedblock--spec--template--spec--affinity--pod_affinity--preferred_during_scheduling_ignored_during_execution--pod_affinity_term--label_selector--match_expressions"></a>
### Nested Schema for `spec.template.spec.affinity.pod_affinity.preferred_during_scheduling_ignored_during_execution.pod_affinity_term.label_selector.match_expressions`

Optional:

- `key` (String) The label key that the selector applies to.
- `operator` (String) A key's relationship to a set of values. Valid operators ard `In`, `NotIn`, `Exists` and `DoesNotExist`.
- `values` (Set of String) An array of string values. If the operator is `In` or `NotIn`, the values array must be non-empty. If the operator is `Exists` or `DoesNotExist`, the values array must be empty. This array is replaced during a strategic merge patch.





<a id="nestedblock--spec--template--spec--affinity--pod_affinity--required_during_scheduling_ignored_during_execution"></a>
### Nested Schema for `spec.template.spec.affinity.pod_affinity.required_during_scheduling_ignored_during_execution`

Optional:

- `label_selector` (Block List) A label query over a set of resources, in this case pods. (see [below for nested schema](#nestedblock--spec--template--spec--affinity--pod_affinity--required_during_scheduling_ignored_during_execution--label_selector))
- `namespaces` (Set of String) namespaces specifies which namespaces the labelSelector applies to (matches against); null or empty list means 'this pod's namespace'
- `topology_key` (String) empty topology key is interpreted by the scheduler as 'all topologies'

<a id="nestedblock--spec--template--spec--affinity--pod_affinity--required_during_scheduling_ignored_during_execution--label_selector"></a>
### Nested Schema for `spec.template.spec.affinity.pod_affinity.required_during_scheduling_ignored_during_execution.label_selector`

Optional:

- `match_expressions` (Block List) A list of label selector requirements. The requirements are ANDed. (see [below for nested schema](#nestedblock--spec--template--spec--affinity--pod_affinity--required_during_scheduling_ignored_during_execution--label_selector--match_expressions))
- `match_labels` (Map of String) A map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of `match_expressions`, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.

<a id="nestedblock--spec--template--spec--affinity--pod_affinity--required_during_scheduling_ignored_during_execution--label_selector--match_expressions"></a>
### Nested Schema for `spec.template.spec.affinity.pod_affinity.required_during_scheduling_ignored_during_execution.label_selector.match_expressions`

Optional:

- `key` (String) The label key that the selector applies to.
- `operator` (String) A key's relationship to a set of values. Valid operators ard `In`, `NotIn`, `Exists` and `DoesNotExist`.
- `values` (Set of String) An array of string values. If the operator is `In` or `NotIn`, the values array must be non-empty. If the operator is `Exists` or `DoesNotExist`, the values array must be empty. This array is replaced during a strategic merge patch.





<a id="nestedblock--spec--template--spec--affinity--pod_anti_affinity"></a>
### Nested Schema for `spec.template.spec.affinity.pod_anti_affinity`

Optional:

- `preferred_during_scheduling_ignored_during_execution` (Block List) The scheduler will prefer to schedule pods to nodes that satisfy the anti-affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, RequiredDuringScheduling anti-affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding 'weight' to the sum if the node matches the corresponding MatchExpressions; the node(s) with the highest sum are the most preferred. (see [below for nested schema](#nestedblock--spec--template--spec--affinity--pod_anti_affinity--preferred_during_scheduling_ignored_during_execution))
- `required_during_scheduling_ignored_during_execution` (Block List) If the affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to a pod label update), the system may or may not try to eventually evict the pod from its node. When there are multiple elements, the lists of nodes corresponding to each PodAffinityTerm are intersected, i.e. all terms must be satisfied. (see [below for nested schema](#nestedblock--spec--template--spec--affinity--pod_anti_affinity--required_during_scheduling_ignored_during_execution))

<a id="nestedblock--spec--template--spec--affinity--pod_anti_affinity--preferred_during_scheduling_ignored_during_execution"></a>
### Nested Schema for `spec.template.spec.affinity.pod_anti_affinity.preferred_during_scheduling_ignored_during_execution`

Required:

- `pod_affinity_term` (Block List, Min: 1, Max: 1) A pod affinity term, associated with the corresponding weight (see [below for nested schema](#nestedblock--spec--template--spec--affinity--pod_anti_affinity--preferred_during_scheduling_ignored_during_execution--pod_affinity_term))
- `weight` (Number) weight associated with matching the corresponding podAffinityTerm, in the range 1-100

<a id="nestedblock--spec--template--spec--affinity--pod_anti_affinity--preferred_during_scheduling_ignored_during_execution--pod_affinity_term"></a>
### Nested Schema for `spec.template.spec.affinity.pod_anti_affinity.preferred_during_scheduling_ignored_during_execution.pod_affinity_term`

Optional:

- `label_selector` (Block List) A label query over a set of resources, in this case pods. (see [below for nested schema](#nestedblock--spec--template--spec--affinity--pod_anti_affinity--preferred_during_scheduling_ignored_during_execution--pod_affinity_term--label_selector))
- `namespaces` (Set of String) namespaces specifies which namespaces the labelSelector applies to (matches against); null or empty list means 'this pod's namespace'
- `topology_key` (String) empty topology key is interpreted by the scheduler as 'all topologies'

<a id="nestedblock--spec--template--spec--affinity--pod_anti_affinity--preferred_during_scheduling_ignored_during_execution--pod_affinity_term--label_selector"></a>
### Nested Schema for `spec.template.spec.affinity.pod_anti_affinity.preferred_during_scheduling_ignored_during_execution.pod_affinity_term.label_selector`

Optional:

- `match_expressions` (Block List) A list of label selector requirements. The requirements are ANDed. (see [below for nested schema](#nestedblock--spec--template--spec--affinity--pod_anti_affinity--preferred_during_scheduling_ignored_during_execution--pod_affinity_term--label_selector--match_expressions))
- `match_labels` (Map of String) A map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of `match_expressions`, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.

<a id="nestedblock--spec--template--spec--affinity--pod_anti_affinity--preferred_during_scheduling_ignored_during_execution--pod_affinity_term--label_selector--match_expressions"></a>
### Nested Schema for `spec.template.spec.affinity.pod_anti_affinity.preferred_during_scheduling_ignored_during_execution.pod_affinity_term.label_selector.match_expressions`

Optional:

- `key` (String) The label key that the selector applies to.
- `operator` (String) A key's relationship to a set of values. Valid operators ard `In`, `NotIn`, `Exists` and `DoesNotExist`.
- `values` (Set of String) An array of string values. If the operator is `In` or `NotIn`, the values array must be non-empty. If the operator is `Exists` or `DoesNotExist`, the values array must be empty. This array is replaced during a strategic merge patch.





<a id="nestedblock--spec--template--spec--affinity--pod_anti_affinity--required_during_scheduling_ignored_during_execution"></a>
### Nested Schema for `spec.template.spec.affinity.pod_anti_affinity.required_during_scheduling_ignored_during_execution`

Optional:

- `label_selector` (Block List) A label query over a set of resources, in this case pods. (see [below for nested schema](#nestedblock--spec--template--spec--affinity--pod_anti_affinity--required_during_scheduling_ignored_during_execution--label_selector))
- `namespaces` (Set of String) namespaces specifies which namespaces the labelSelector applies to (matches against); null or empty list means 'this pod's namespace'
- `topology_key` (String) empty topology key is interpreted by the scheduler as 'all topologies'

<a id="nestedblock--spec--template--spec--affinity--pod_anti_affinity--required_during_scheduling_ignored_during_execution--label_selector"></a>
### Nested Schema for `spec.template.spec.affinity.pod_anti_affinity.required_during_scheduling_ignored_during_execution.label_selector`

Optional:

- `match_expressions` (Block List) A list of label selector requirements. The requirements are ANDed. (see [below for nested schema](#nestedblock--spec--template--spec--affinity--pod_anti_affinity--required_during_scheduling_ignored_during_execution--label_selector--match_expressions))
- `match_labels` (Map of String) A map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of `match_expressions`, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.

<a id="nestedblock--spec--template--spec--affinity--pod_anti_affinity--required_during_scheduling_ignored_during_execution--label_selector--match_expressions"></a>
### Nested Schema for `spec.template.spec.affinity.pod_anti_affinity.required_during_scheduling_ignored_during_execution.label_selector.match_expressions`

Optional:

- `key` (String) The label key that the selector applies to.
- `operator` (String) A key's relationship to a set of values. Valid operators ard `In`, `NotIn`, `Exists` and `DoesNotExist`.
- `values` (Set of String) An array of string values. If the operator is `In` or `NotIn`, the values array must be non-empty. If the operator is `Exists` or `DoesNotExist`, the values array must be empty. This array is replaced during a strategic merge patch.






<a id="nestedblock--spec--template--spec--domain"></a>
### Nested Schema for `spec.template.spec.domain`

Required:

- `devices` (Block List, Min: 1, Max: 1) Devices allows adding disks, network interfaces, ... (see [below for nested schema](#nestedblock--spec--template--spec--domain--devices))
- `resources` (Block List, Min: 1, Max: 1) Resources describes the Compute Resources required by this vmi. (see [below for nested schema](#nestedblock--spec--template--spec--domain--resources))

Optional:

- `cpu` (Block List, Max: 1) CPU allows to specifying the CPU topology. Valid resource keys are "cores" , "sockets" and "threads" (see [below for nested schema](#nestedblock--spec--template--spec--domain--cpu))
- `memory` (Block List, Max: 1) Memory allows specifying the vmi memory features. (see [below for nested schema](#nestedblock--spec--template--spec--domain--memory))

<a id="nestedblock--spec--template--spec--domain--devices"></a>
### Nested Schema for `spec.template.spec.domain.devices`

Required:

- `disk` (Block List, Min: 1) Disks describes disks, cdroms, floppy and luns which are connected to the vmi. (see [below for nested schema](#nestedblock--spec--template--spec--domain--devices--disk))

Optional:

- `interface` (Block List) Interfaces describe network interfaces which are added to the vmi. (see [below for nested schema](#nestedblock--spec--template--spec--domain--devices--interface))

<a id="nestedblock--spec--template--spec--domain--devices--disk"></a>
### Nested Schema for `spec.template.spec.domain.devices.disk`

Required:

- `disk_device` (Block List, Min: 1) DiskDevice specifies as which device the disk should be added to the guest. (see [below for nested schema](#nestedblock--spec--template--spec--domain--devices--disk--disk_device))
- `name` (String) Name is the device name

Optional:

- `serial` (String) Serial provides the ability to specify a serial number for the disk device.

<a id="nestedblock--spec--template--spec--domain--devices--disk--disk_device"></a>
### Nested Schema for `spec.template.spec.domain.devices.disk.disk_device`

Optional:

- `disk` (Block List) Attach a volume as a disk to the vmi. (see [below for nested schema](#nestedblock--spec--template--spec--domain--devices--disk--disk_device--disk))

<a id="nestedblock--spec--template--spec--domain--devices--disk--disk_device--disk"></a>
### Nested Schema for `spec.template.spec.domain.devices.disk.disk_device.disk`

Required:

- `bus` (String) Bus indicates the type of disk device to emulate.

Optional:

- `pci_address` (String) If specified, the virtual disk will be placed on the guests pci address with the specifed PCI address. For example: 0000:81:01.10
- `read_only` (Boolean) ReadOnly. Defaults to false.




<a id="nestedblock--spec--template--spec--domain--devices--interface"></a>
### Nested Schema for `spec.template.spec.domain.devices.interface`

Required:

- `interface_binding_method` (String) Represents the Interface model, One of: e1000, e1000e, ne2k_pci, pcnet, rtl8139, virtio. Defaults to virtio.
- `name` (String) Logical name of the interface as well as a reference to the associated networks.

Optional:

- `model` (String) Represents the method which will be used to connect the interface to the guest.



<a id="nestedblock--spec--template--spec--domain--resources"></a>
### Nested Schema for `spec.template.spec.domain.resources`

Optional:

- `limits` (Map of String) Requests is the maximum amount of compute resources allowed. Valid resource keys are "memory" and "cpu"
- `over_commit_guest_overhead` (Boolean) Don't ask the scheduler to take the guest-management overhead into account. Instead put the overhead only into the container's memory limit. This can lead to crashes if all memory is in use on a node. Defaults to false.
- `requests` (Map of String) Requests is a description of the initial vmi resources.


<a id="nestedblock--spec--template--spec--domain--cpu"></a>
### Nested Schema for `spec.template.spec.domain.cpu`

Optional:

- `cores` (Number) Cores is the number of cores inside the vmi. Must be a value greater or equal 1
- `sockets` (Number) Sockets is the number of sockets inside the vmi. Must be a value greater or equal 1.
- `threads` (Number) Threads is the number of threads inside the vmi. Must be a value greater or equal 1.


<a id="nestedblock--spec--template--spec--domain--memory"></a>
### Nested Schema for `spec.template.spec.domain.memory`

Optional:

- `guest` (String) Guest is the amount of memory allocated to the vmi. This value must be less than or equal to the limit if specified.
- `hugepages` (String) Hugepages attribute specifies the hugepage size, for x86_64 architecture valid values are 1Gi and 2Mi.



<a id="nestedblock--spec--template--spec--liveness_probe"></a>
### Nested Schema for `spec.template.spec.liveness_probe`


<a id="nestedblock--spec--template--spec--network"></a>
### Nested Schema for `spec.template.spec.network`

Required:

- `name` (String) Network name.

Optional:

- `network_source` (Block List, Max: 1) NetworkSource represents the network type and the source interface that should be connected to the virtual machine. (see [below for nested schema](#nestedblock--spec--template--spec--network--network_source))

<a id="nestedblock--spec--template--spec--network--network_source"></a>
### Nested Schema for `spec.template.spec.network.network_source`

Optional:

- `multus` (Block List, Max: 1) Multus network. (see [below for nested schema](#nestedblock--spec--template--spec--network--network_source--multus))
- `pod` (Block List, Max: 1) Pod network. (see [below for nested schema](#nestedblock--spec--template--spec--network--network_source--pod))

<a id="nestedblock--spec--template--spec--network--network_source--multus"></a>
### Nested Schema for `spec.template.spec.network.network_source.multus`

Required:

- `network_name` (String) References to a NetworkAttachmentDefinition CRD object. Format: <networkName>, <namespace>/<networkName>. If namespace is not specified, VMI namespace is assumed.

Optional:

- `default` (Boolean) Select the default network and add it to the multus-cni.io/default-network annotation.


<a id="nestedblock--spec--template--spec--network--network_source--pod"></a>
### Nested Schema for `spec.template.spec.network.network_source.pod`

Optional:

- `vm_network_cidr` (String) CIDR for vm network.




<a id="nestedblock--spec--template--spec--pod_dns_config"></a>
### Nested Schema for `spec.template.spec.pod_dns_config`

Optional:

- `nameservers` (List of String) A list of DNS name server IP addresses. This will be appended to the base nameservers generated from DNSPolicy. Duplicated nameservers will be removed.
- `option` (Block List) A list of DNS resolver options. This will be merged with the base options generated from DNSPolicy. Duplicated entries will be removed. Resolution options given in Options will override those that appear in the base DNSPolicy. (see [below for nested schema](#nestedblock--spec--template--spec--pod_dns_config--option))
- `searches` (List of String) A list of DNS search domains for host-name lookup. This will be appended to the base search paths generated from DNSPolicy. Duplicated search paths will be removed.

<a id="nestedblock--spec--template--spec--pod_dns_config--option"></a>
### Nested Schema for `spec.template.spec.pod_dns_config.option`

Required:

- `name` (String) Name of the option.

Optional:

- `value` (String) Value of the option. Optional: Defaults to empty.



<a id="nestedblock--spec--template--spec--readiness_probe"></a>
### Nested Schema for `spec.template.spec.readiness_probe`


<a id="nestedblock--spec--template--spec--tolerations"></a>
### Nested Schema for `spec.template.spec.tolerations`

Optional:

- `effect` (String) Effect indicates the taint effect to match. Empty means match all taint effects. When specified, allowed values are NoSchedule, PreferNoSchedule and NoExecute.
- `key` (String) Key is the taint key that the toleration applies to. Empty means match all taint keys. If the key is empty, operator must be Exists; this combination means to match all values and all keys.
- `operator` (String) Operator represents a key's relationship to the value. Valid operators are Exists and Equal. Defaults to Equal. Exists is equivalent to wildcard for value, so that a pod can tolerate all taints of a particular category.
- `toleration_seconds` (String) TolerationSeconds represents the period of time the toleration (which must be of effect NoExecute, otherwise this field is ignored) tolerates the taint. By default, it is not set, which means tolerate the taint forever (do not evict). Zero and negative values will be treated as 0 (evict immediately) by the system.
- `value` (String) Value is the taint value the toleration matches to. If the operator is Exists, the value should be empty, otherwise just a regular string.


<a id="nestedblock--spec--template--spec--volume"></a>
### Nested Schema for `spec.template.spec.volume`

Required:

- `name` (String) Volume's name.
- `volume_source` (Block List, Min: 1, Max: 1) VolumeSource represents the location and type of the mounted volume. Defaults to Disk, if no type is specified. (see [below for nested schema](#nestedblock--spec--template--spec--volume--volume_source))

<a id="nestedblock--spec--template--spec--volume--volume_source"></a>
### Nested Schema for `spec.template.spec.volume.volume_source`

Optional:

- `cloud_init_config_drive` (Block List, Max: 1) CloudInitConfigDrive represents a cloud-init Config Drive user-data source. (see [below for nested schema](#nestedblock--spec--template--spec--volume--volume_source--cloud_init_config_drive))
- `cloud_init_no_cloud` (Block Set) Used to specify a cloud-init `noCloud` image. The image is expected to contain a disk image in a supported format. The disk image is extracted from the cloud-init `noCloud `image and used as the disk for the VM (see [below for nested schema](#nestedblock--spec--template--spec--volume--volume_source--cloud_init_no_cloud))
- `config_map` (Block List, Max: 1) ConfigMapVolumeSource adapts a ConfigMap into a volume. (see [below for nested schema](#nestedblock--spec--template--spec--volume--volume_source--config_map))
- `container_disk` (Block Set) A container disk is a disk that is backed by a container image. The container image is expected to contain a disk image in a supported format. The disk image is extracted from the container image and used as the disk for the VM. (see [below for nested schema](#nestedblock--spec--template--spec--volume--volume_source--container_disk))
- `data_volume` (Block List, Max: 1) DataVolume represents the dynamic creation a PVC for this volume as well as the process of populating that PVC with a disk image. (see [below for nested schema](#nestedblock--spec--template--spec--volume--volume_source--data_volume))
- `empty_disk` (Block List, Max: 1) EmptyDisk represents a temporary disk which shares the VM's lifecycle. (see [below for nested schema](#nestedblock--spec--template--spec--volume--volume_source--empty_disk))
- `ephemeral` (Block List, Max: 1) EphemeralVolumeSource represents a volume that is populated with the contents of a pod. Ephemeral volumes do not support ownership management or SELinux relabeling. (see [below for nested schema](#nestedblock--spec--template--spec--volume--volume_source--ephemeral))
- `host_disk` (Block List, Max: 1) HostDisk represents a disk created on the host. (see [below for nested schema](#nestedblock--spec--template--spec--volume--volume_source--host_disk))
- `persistent_volume_claim` (Block List, Max: 1) PersistentVolumeClaimVolumeSource represents a reference to a PersistentVolumeClaim in the same namespace. (see [below for nested schema](#nestedblock--spec--template--spec--volume--volume_source--persistent_volume_claim))
- `service_account` (Block List, Max: 1) ServiceAccountVolumeSource represents a reference to a service account. (see [below for nested schema](#nestedblock--spec--template--spec--volume--volume_source--service_account))

<a id="nestedblock--spec--template--spec--volume--volume_source--cloud_init_config_drive"></a>
### Nested Schema for `spec.template.spec.volume.volume_source.cloud_init_config_drive`

Optional:

- `network_data` (String) NetworkData contains config drive inline cloud-init networkdata.
- `network_data_base64` (String) NetworkDataBase64 contains config drive cloud-init networkdata as a base64 encoded string.
- `network_data_secret_ref` (Block List, Max: 1) NetworkDataSecretRef references a k8s secret that contains config drive networkdata. (see [below for nested schema](#nestedblock--spec--template--spec--volume--volume_source--cloud_init_config_drive--network_data_secret_ref))
- `user_data` (String) UserData contains config drive inline cloud-init userdata.
- `user_data_base64` (String) UserDataBase64 contains config drive cloud-init userdata as a base64 encoded string.
- `user_data_secret_ref` (Block List, Max: 1) UserDataSecretRef references a k8s secret that contains config drive userdata. (see [below for nested schema](#nestedblock--spec--template--spec--volume--volume_source--cloud_init_config_drive--user_data_secret_ref))

<a id="nestedblock--spec--template--spec--volume--volume_source--cloud_init_config_drive--network_data_secret_ref"></a>
### Nested Schema for `spec.template.spec.volume.volume_source.cloud_init_config_drive.network_data_secret_ref`

Required:

- `name` (String) Name of the referent.


<a id="nestedblock--spec--template--spec--volume--volume_source--cloud_init_config_drive--user_data_secret_ref"></a>
### Nested Schema for `spec.template.spec.volume.volume_source.cloud_init_config_drive.user_data_secret_ref`

Required:

- `name` (String) Name of the referent.



<a id="nestedblock--spec--template--spec--volume--volume_source--cloud_init_no_cloud"></a>
### Nested Schema for `spec.template.spec.volume.volume_source.cloud_init_no_cloud`

Required:

- `user_data` (String) The user data to use for the cloud-init no cloud disk. This can be a local file path, a remote URL, or a registry URL.


<a id="nestedblock--spec--template--spec--volume--volume_source--config_map"></a>
### Nested Schema for `spec.template.spec.volume.volume_source.config_map`

Optional:

- `default_mode` (Number) Optional: mode bits to use on created files by default. Must be a value between 0 and 0777. Defaults to 0644. Directories within the path are not affected by this setting. This might be in conflict with other options that affect the file mode, like fsGroup, and the result can be other mode bits set.
- `items` (Block List) If unspecified, each key-value pair in the Data field of the referenced ConfigMap will be projected into the volume as a file whose name is the key and content is the value. If specified, the listed keys will be projected into the specified paths, and unlisted keys will not be present. If a key is specified which is not present in the ConfigMap, the volume setup will error unless it is marked optional. Paths must be relative and may not contain the '..' path or start with '..'. (see [below for nested schema](#nestedblock--spec--template--spec--volume--volume_source--config_map--items))

<a id="nestedblock--spec--template--spec--volume--volume_source--config_map--items"></a>
### Nested Schema for `spec.template.spec.volume.volume_source.config_map.items`

Optional:

- `key` (String)



<a id="nestedblock--spec--template--spec--volume--volume_source--container_disk"></a>
### Nested Schema for `spec.template.spec.volume.volume_source.container_disk`

Required:

- `image_url` (String) The URL of the container image to use as the disk. This can be a local file path, a remote URL, or a registry URL.


<a id="nestedblock--spec--template--spec--volume--volume_source--data_volume"></a>
### Nested Schema for `spec.template.spec.volume.volume_source.data_volume`

Required:

- `name` (String) Name represents the name of the DataVolume in the same namespace.


<a id="nestedblock--spec--template--spec--volume--volume_source--empty_disk"></a>
### Nested Schema for `spec.template.spec.volume.volume_source.empty_disk`

Required:

- `capacity` (String) Capacity of the sparse disk.


<a id="nestedblock--spec--template--spec--volume--volume_source--ephemeral"></a>
### Nested Schema for `spec.template.spec.volume.volume_source.ephemeral`

Optional:

- `persistent_volume_claim` (Block List, Max: 1) PersistentVolumeClaimVolumeSource represents a reference to a PersistentVolumeClaim in the same namespace. (see [below for nested schema](#nestedblock--spec--template--spec--volume--volume_source--ephemeral--persistent_volume_claim))

<a id="nestedblock--spec--template--spec--volume--volume_source--ephemeral--persistent_volume_claim"></a>
### Nested Schema for `spec.template.spec.volume.volume_source.ephemeral.persistent_volume_claim`

Required:

- `claim_name` (String) ClaimName is the name of a PersistentVolumeClaim in the same namespace as the pod using this volume. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims

Optional:

- `read_only` (Boolean) Will force the ReadOnly setting in VolumeMounts. Default false.



<a id="nestedblock--spec--template--spec--volume--volume_source--host_disk"></a>
### Nested Schema for `spec.template.spec.volume.volume_source.host_disk`

Required:

- `path` (String) Path of the disk.
- `type` (String) Type of the disk, supported values are disk, directory, socket, char, block.


<a id="nestedblock--spec--template--spec--volume--volume_source--persistent_volume_claim"></a>
### Nested Schema for `spec.template.spec.volume.volume_source.persistent_volume_claim`

Required:

- `claim_name` (String) ClaimName is the name of a PersistentVolumeClaim in the same namespace as the pod using this volume. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims

Optional:

- `read_only` (Boolean) Will force the ReadOnly setting in VolumeMounts. Default false.


<a id="nestedblock--spec--template--spec--volume--volume_source--service_account"></a>
### Nested Schema for `spec.template.spec.volume.volume_source.service_account`

Required:

- `service_account_name` (String) Name of the service account in the pod's namespace to use.







<a id="nestedblock--status"></a>
### Nested Schema for `status`

Required:

- `conditions` (Block List, Min: 1) Hold the state information of the VirtualMachine and its VirtualMachineInstance. (see [below for nested schema](#nestedblock--status--conditions))
- `state_change_requests` (Block List, Min: 1) StateChangeRequests indicates a list of actions that should be taken on a VMI. (see [below for nested schema](#nestedblock--status--state_change_requests))

Optional:

- `created` (Boolean) Created indicates if the virtual machine is created in the cluster.
- `ready` (Boolean) Ready indicates if the virtual machine is running and ready.

<a id="nestedblock--status--conditions"></a>
### Nested Schema for `status.conditions`

Optional:

- `message` (String) Condition message.
- `reason` (String) Condition reason.
- `status` (String) ConditionStatus represents the status of this VM condition, if the VM currently in the condition.
- `type` (String) VirtualMachineConditionType represent the type of the VM as concluded from its VMi status.


<a id="nestedblock--status--state_change_requests"></a>
### Nested Schema for `status.state_change_requests`

Optional:

- `action` (String) Indicates the type of action that is requested. e.g. Start or Stop.
- `data` (Map of String) Provides additional data in order to perform the Action.
- `uid` (String) Indicates the UUID of an existing Virtual Machine Instance that this change request applies to -- if applicable.



<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `delete` (String)
---
page_title: "spectrocloud_cluster_cloudstack Resource - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  Resource for managing CloudStack clusters in Spectro Cloud through Palette.
---

# spectrocloud_cluster_cloudstack (Resource)

Resource for managing CloudStack clusters in Spectro Cloud through Palette.

## Example Usage

### Basic Single-Zone CloudStack Cluster

```terraform
data "spectrocloud_cloudaccount_cloudstack" "account" {
  name = var.cloudstack_account_name
}

data "spectrocloud_cluster_profile" "profile" {
  name = var.cluster_profile_name
}

resource "spectrocloud_cluster_cloudstack" "cluster" {
  name             = "cloudstack-cluster-basic"
  cloud_account_id = data.spectrocloud_cloudaccount_cloudstack.account.id

  cluster_profile {
    id = data.spectrocloud_cluster_profile.profile.id
  }

  cloud_config {
    domain       = "production"
    ssh_key_name = "my-ssh-key"
    
    zone {
      name = "zone1"
      network {
        name = "prod-network"
      }
    }
  }

  machine_pool {
    name          = "control-plane-pool"
    count         = 3
    control_plane = true
    
    template = "ubuntu-22.04-kube-v1.28.0"
    offering = "Medium Instance"
  }

  machine_pool {
    name  = "worker-pool"
    count = 3
    
    template = "ubuntu-22.04-kube-v1.28.0"
    offering = "Large Instance"
  }
}
```

### Multi-Zone CloudStack Cluster with Autoscaling

```terraform
data "spectrocloud_cloudaccount_cloudstack" "account" {
  name = var.cloudstack_account_name
}

data "spectrocloud_cluster_profile" "profile" {
  name = var.cluster_profile_name
}

data "spectrocloud_backup_storage_location" "bsl" {
  name = var.backup_storage_location_name
}

resource "spectrocloud_cluster_cloudstack" "cluster_ha" {
  name             = "cloudstack-cluster-ha"
  tags             = ["env:production", "team:devops"]
  cloud_account_id = data.spectrocloud_cloudaccount_cloudstack.account.id

  cluster_profile {
    id = data.spectrocloud_cluster_profile.profile.id
  }

  cloud_config {
    domain       = "production"
    project      = "prod-project"
    ssh_key_name = "prod-ssh-key"
    
    # Static control plane endpoint (optional)
    control_plane_endpoint = "192.168.1.100"
    
    # Multiple zones for high availability
    zone {
      name = "zone1"
      network {
        name    = "zone1-network"
        type    = "Isolated"
        gateway = "192.168.1.1"
        netmask = "255.255.255.0"
      }
    }
    
    zone {
      name = "zone2"
      network {
        name    = "zone2-network"
        type    = "Isolated"
        gateway = "192.168.2.1"
        netmask = "255.255.255.0"
      }
    }
  }

  machine_pool {
    name                    = "control-plane-pool"
    count                   = 3
    control_plane           = true
    control_plane_as_worker = false
    
    template           = "ubuntu-22.04-kube-v1.28.0"
    offering           = "Medium Instance"
    root_disk_size_gb  = 100
    
    # Affinity groups for anti-affinity
    affinity_group_ids = ["anti-affinity-group-1"]
    
    network {
      network_name = "control-plane-network"
    }
  }

  machine_pool {
    name  = "worker-pool"
    count = 5
    
    # Enable autoscaling
    min = 3
    max = 10
    
    template           = "ubuntu-22.04-kube-v1.28.0"
    offering           = "Large Instance"
    root_disk_size_gb  = 200
    disk_offering      = "Custom SSD"
    
    # Static IP for first worker (optional)
    network {
      network_name = "worker-network"
      ip_address   = "192.168.1.50"
    }
    
    additional_labels = {
      "workload" = "general"
      "tier"     = "backend"
    }
    
    taints {
      key    = "dedicated"
      value  = "worker"
      effect = "NoSchedule"
    }
    
    # Custom instance details
    details = {
      "custom_key" = "custom_value"
    }
  }

  # Backup policy
  backup_policy {
    schedule                  = "0 1 * * *"
    backup_location_id        = data.spectrocloud_backup_storage_location.bsl.id
    prefix                    = "cloudstack-backup"
    expiry_in_hour            = 168  # 7 days
    include_disks             = true
    include_cluster_resources = true
  }

  # Security scan policy
  scan_policy {
    configuration_scan_schedule = "0 2 * * *"
    penetration_scan_schedule   = "0 3 * * 6"  # Weekly on Saturday
    conformance_scan_schedule   = "0 4 * * *"
  }

  # OS patching
  os_patch_on_boot = false
  os_patch_schedule = "0 5 * * SUN"  # Weekly on Sunday
}
```

### CloudStack Cluster with VPC Networking

```terraform
data "spectrocloud_cloudaccount_cloudstack" "account" {
  name = var.cloudstack_account_name
}

data "spectrocloud_cluster_profile" "profile" {
  name = var.cluster_profile_name
}

resource "spectrocloud_cluster_cloudstack" "vpc_cluster" {
  name             = "cloudstack-vpc-cluster"
  tags             = ["env:production", "network:vpc"]
  cloud_account_id = data.spectrocloud_cloudaccount_cloudstack.account.id

  cluster_profile {
    id = data.spectrocloud_cluster_profile.profile.id
  }

  cloud_config {
    domain       = "production"
    project      = "vpc-project"
    ssh_key_name = "vpc-ssh-key"
    
    zone {
      name = "zone1"
      network {
        name         = "vpc-network"
        type         = "Isolated"
        gateway      = "10.0.1.1"
        netmask      = "255.255.255.0"
        offering     = "DefaultNetworkOffering"
        routing_mode = "Static"
        
        # VPC configuration for VPC-based deployments
        vpc {
          name     = "production-vpc"
          cidr     = "10.0.0.0/16"
          offering = "Default VPC Offering"
        }
      }
    }
    
    # Additional zone in the same VPC
    zone {
      name = "zone2"
      network {
        name         = "vpc-network-zone2"
        type         = "Isolated"
        gateway      = "10.0.2.1"
        netmask      = "255.255.255.0"
        offering     = "DefaultNetworkOffering"
        routing_mode = "Static"
        
        vpc {
          name     = "production-vpc"
          cidr     = "10.0.0.0/16"
          offering = "Default VPC Offering"
        }
      }
    }
  }

  machine_pool {
    name                    = "vpc-control-plane"
    count                   = 3
    control_plane           = true
    control_plane_as_worker = false
    
    template = "ubuntu-22.04-kube-v1.28.0"
    offering = "Medium Instance"
    
    network {
      network_name = "vpc-network"
      ip_address   = "10.0.1.10"  # Static IP in VPC subnet
    }
  }

  machine_pool {
    name  = "vpc-workers"
    count = 3
    min   = 3
    max   = 10
    
    template          = "ubuntu-22.04-kube-v1.28.0"
    offering          = "Large Instance"
    root_disk_size_gb = 200
    
    network {
      network_name = "vpc-network"
    }
    
    additional_labels = {
      "vpc"      = "production-vpc"
      "workload" = "general"
    }
  }
}
```

### CloudStack Cluster with Custom Pack Values

```terraform
data "spectrocloud_cloudaccount_cloudstack" "account" {
  name = var.cloudstack_account_name
}

data "spectrocloud_cluster_profile" "profile" {
  name = var.cluster_profile_name
}

resource "spectrocloud_cluster_cloudstack" "cluster_custom" {
  name             = "cloudstack-cluster-custom"
  cloud_account_id = data.spectrocloud_cloudaccount_cloudstack.account.id

  cluster_profile {
    id = data.spectrocloud_cluster_profile.profile.id

    # Override pack values
    pack {
      name   = "kubernetes"
      tag    = "1.28.0"
      values = <<-EOT
        pack:
          k8sHardening: true
          podCIDR: "192.168.0.0/16"
          serviceCIDR: "10.96.0.0/12"
      EOT
    }
    
    pack {
      name   = "cni-calico"
      tag    = "3.26.x"
      values = <<-EOT
        manifests:
          calico:
            env:
              calicoBackend: "bird"
      EOT
    }
  }

  cloud_config {
    domain       = "development"
    ssh_key_name = "dev-key"
    
    zone {
      name = "zone1"
      network {
        name = "dev-network"
      }
    }
  }

  machine_pool {
    name          = "control-plane-pool"
    count         = 1
    control_plane = true
    
    template = "ubuntu-22.04-kube-v1.28.0"
    offering = "Small Instance"
  }

  machine_pool {
    name  = "worker-pool"
    count = 2
    
    template = "ubuntu-22.04-kube-v1.28.0"
    offering = "Medium Instance"
  }
}
```

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import)
to import the resource spectrocloud_cluster_cloudstack by using its `id` with the Palette `context` separated by a colon. For example:

```terraform
import {
  to = spectrocloud_cluster_cloudstack.example
  id = "example_id:context"
}
```

Using `terraform import`, import the cluster using the `id` colon separated with `context`. For example:

```console
terraform import spectrocloud_cluster_cloudstack.example example_id:project
```

Refer to the [Import section](/docs#import) to learn more.

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `cloud_account_id` (String) ID of the CloudStack cloud account used for the cluster. This cloud account must be of type `cloudstack`.
- `cloud_config` (Block List, Min: 1, Max: 1) CloudStack cluster configuration. (see [below for nested schema](#nestedblock--cloud_config))
- `machine_pool` (Block Set, Min: 1) Machine pool configuration for the cluster. (see [below for nested schema](#nestedblock--machine_pool))
- `name` (String) The name of the cluster.

### Optional

- `apply_setting` (String) The setting to apply the cluster profile. `DownloadAndInstall` will download and install packs in one action. `DownloadAndInstallLater` will only download artifact and postpone install for later. Default value is `DownloadAndInstall`.
- `backup_policy` (Block List, Max: 1) The backup policy for the cluster. If not specified, no backups will be taken. (see [below for nested schema](#nestedblock--backup_policy))
- `cluster_meta_attribute` (String) `cluster_meta_attribute` can be used to set additional cluster metadata information, eg `{'nic_name': 'test', 'env': 'stage'}`
- `cluster_profile` (Block List) (see [below for nested schema](#nestedblock--cluster_profile))
- `cluster_rbac_binding` (Block List) The RBAC binding for the cluster. (see [below for nested schema](#nestedblock--cluster_rbac_binding))
- `context` (String) The context of the CloudStack configuration. Allowed values are `project` or `tenant`. Default is `project`. If  the `project` context is specified, the project name will sourced from the provider configuration parameter [`project_name`](https://registry.terraform.io/providers/spectrocloud/spectrocloud/latest/docs#schema).
- `description` (String) The description of the cluster. Default value is empty string.
- `host_config` (Block List) The host configuration for the cluster. (see [below for nested schema](#nestedblock--host_config))
- `namespaces` (Block List) The namespaces for the cluster. (see [below for nested schema](#nestedblock--namespaces))
- `os_patch_after` (String) The date and time after which to patch the cluster. Prefix the time value with the respective RFC. Ex: `RFC3339: 2006-01-02T15:04:05Z07:00`
- `os_patch_on_boot` (Boolean) Whether to apply OS patch on boot. Default is `false`.
- `os_patch_schedule` (String) Cron schedule for OS patching. This must be in the form of `0 0 * * *`.
- `pause_agent_upgrades` (String) The pause agent upgrades setting allows to control the automatic upgrade of the Palette component and agent for an individual cluster. The default value is `unlock`, meaning upgrades occur automatically. Setting it to `lock` pauses automatic agent upgrades for the cluster.
- `review_repave_state` (String) To authorize the cluster repave, set the value to `Approved` for approval and `""` to decline. Default value is `""`.
- `scan_policy` (Block List, Max: 1) The scan policy for the cluster. (see [below for nested schema](#nestedblock--scan_policy))
- `tags` (Set of String) A list of tags to be applied to the cluster. Tags must be in the form of `key:value`.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `admin_kube_config` (String) Admin Kube-config for the cluster. This can be used to connect to the cluster using `kubectl`, With admin privilege.
- `cloud_config_id` (String, Deprecated) ID of the cloud config used for the cluster. This cloud config must be of type `cloudstack`.
- `id` (String) The ID of this resource.
- `kubeconfig` (String) Kubeconfig for the cluster. This can be used to connect to the cluster using `kubectl`.
- `location_config` (List of Object) The location of the cluster. (see [below for nested schema](#nestedatt--location_config))

<a id="nestedblock--cloud_config"></a>
### Nested Schema for `cloud_config`

Required:

- `domain` (String) CloudStack domain name in which the cluster will be provisioned.
- `zone` (Block List, Min: 1) List of CloudStack zones for multi-AZ deployments. If only one zone is specified, it will be treated as single-zone deployment. (see [below for nested schema](#nestedblock--cloud_config--zone))

Optional:

- `control_plane_endpoint` (String) Endpoint IP to be used for the API server. Should only be set for static CloudStack networks.
- `project` (String) CloudStack project name (optional). If not specified, the cluster will be created in the domain's default project.
- `ssh_key_name` (String) SSH key name for accessing cluster nodes.

<a id="nestedblock--cloud_config--zone"></a>
### Nested Schema for `cloud_config.zone`

Required:

- `name` (String) CloudStack zone name where the cluster will be deployed.

Optional:

- `network` (Block List, Max: 1) Network configuration for this zone. (see [below for nested schema](#nestedblock--cloud_config--zone--network))

<a id="nestedblock--cloud_config--zone--network"></a>
### Nested Schema for `cloud_config.zone.network`

Required:

- `name` (String) Network name in this zone.

Optional:

- `gateway` (String) Gateway IP address for the network.
- `netmask` (String) Network mask for the network.
- `offering` (String) Network offering name to use when creating the network. Optional for advanced network configurations.
- `routing_mode` (String) Routing mode for the network (e.g., Static, Dynamic). Optional, defaults to CloudStack's default routing mode.
- `type` (String) Network type: Isolated, Shared, etc.
- `vpc` (Block List, Max: 1) VPC configuration for VPC-based network deployments. Optional, only needed when deploying in a VPC. (see [below for nested schema](#nestedblock--cloud_config--zone--network--vpc))

<a id="nestedblock--cloud_config--zone--network--vpc"></a>
### Nested Schema for `cloud_config.zone.network.vpc`

Required:

- `name` (String) VPC name.

Optional:

- `cidr` (String) CIDR block for the VPC (e.g., 10.0.0.0/16).
- `offering` (String) VPC offering name.





<a id="nestedblock--machine_pool"></a>
### Nested Schema for `machine_pool`

Required:

- `count` (Number) Number of nodes in the machine pool.
- `name` (String) Name of the machine pool.
- `offering` (String) CloudStack compute offering (instance type/size) name.
- `template` (String) CloudStack VM template (image) name to use for the instances.

Optional:

- `additional_labels` (Map of String) Additional labels to be applied to the machine pool. Labels must be in the form of `key:value`.
- `affinity_group_ids` (Set of String) List of affinity group IDs for VM placement (optional).
- `control_plane` (Boolean) Whether this machine pool is a control plane. Defaults to `false`.
- `control_plane_as_worker` (Boolean) Whether this machine pool is a control plane and a worker. Defaults to `false`.
- `details` (Map of String) Additional details for instance creation as key-value pairs.
- `disk_offering` (String) CloudStack disk offering name for root disk (optional).
- `max` (Number) Maximum number of nodes in the machine pool. This is used for autoscaling.
- `min` (Number) Minimum number of nodes in the machine pool. This is used for autoscaling.
- `network` (Block List) Network configuration for the machine pool instances. (see [below for nested schema](#nestedblock--machine_pool--network))
- `node` (Block List) (see [below for nested schema](#nestedblock--machine_pool--node))
- `root_disk_size_gb` (Number) Root disk size in GB (optional).
- `taints` (Block List) (see [below for nested schema](#nestedblock--machine_pool--taints))
- `update_strategy` (String) Update strategy for the machine pool. Valid values are `RollingUpdateScaleOut` and `RollingUpdateScaleIn`.

<a id="nestedblock--machine_pool--network"></a>
### Nested Schema for `machine_pool.network`

Required:

- `network_name` (String) Network name to attach to the machine pool.

Optional:

- `ip_address` (String) Static IP address to assign (optional, for static IP configuration).


<a id="nestedblock--machine_pool--node"></a>
### Nested Schema for `machine_pool.node`

Required:

- `action` (String) The action to perform on the node. Valid values are: `cordon`, `uncordon`.
- `node_id` (String) The node_id of the node, For example `i-07f899a33dee624f7`


<a id="nestedblock--machine_pool--taints"></a>
### Nested Schema for `machine_pool.taints`

Required:

- `effect` (String) The effect of the taint. Allowed values are: `NoSchedule`, `PreferNoSchedule` or `NoExecute`.
- `key` (String) The key of the taint.
- `value` (String) The value of the taint.



<a id="nestedblock--backup_policy"></a>
### Nested Schema for `backup_policy`

Required:

- `backup_location_id` (String) The ID of the backup location to use for the backup.
- `expiry_in_hour` (Number) The number of hours after which the backup will be deleted. For example, if the expiry is set to 24, the backup will be deleted after 24 hours.
- `prefix` (String) Prefix for the backup name. The backup name will be of the format <prefix>-<cluster-name>-<timestamp>.
- `schedule` (String) The schedule for the backup. The schedule is specified in cron format. For example, to run the backup every day at 1:00 AM, the schedule should be set to `0 1 * * *`.

Optional:

- `cluster_uids` (Set of String) The list of cluster UIDs to include in the backup. If `include_all_clusters` is set to `true`, then all clusters will be included.
- `include_all_clusters` (Boolean) Whether to include all clusters in the backup. If set to false, only the clusters specified in `cluster_uids` will be included.
- `include_cluster_resources` (Boolean) Indicates whether to include cluster resources in the backup. If set to false, only the cluster configuration and disks will be backed up. (Note: Starting with Palette version 4.6, the include_cluster_resources attribute will be deprecated, and a new attribute, include_cluster_resources_mode, will be introduced.)
- `include_cluster_resources_mode` (String) Specifies whether to include the cluster resources in the backup. Supported values are `always`, `never`, and `auto`.
- `include_disks` (Boolean) Whether to include the disks in the backup. If set to false, only the cluster configuration will be backed up.
- `namespaces` (Set of String) The list of Kubernetes namespaces to include in the backup. If not specified, all namespaces will be included.


<a id="nestedblock--cluster_profile"></a>
### Nested Schema for `cluster_profile`

Required:

- `id` (String) The ID of the cluster profile.

Optional:

- `pack` (Block List) For packs of type `spectro`, `helm`, and `manifest`, at least one pack must be specified. (see [below for nested schema](#nestedblock--cluster_profile--pack))
- `variables` (Map of String) A map of cluster profile variables, specified as key-value pairs. For example: `priority = "5"`.

<a id="nestedblock--cluster_profile--pack"></a>
### Nested Schema for `cluster_profile.pack`

Required:

- `name` (String) The name of the pack. The name must be unique within the cluster profile.

Optional:

- `manifest` (Block List) (see [below for nested schema](#nestedblock--cluster_profile--pack--manifest))
- `registry_name` (String) The registry name of the pack. The registry name is the human-readable name of the registry. This attribute can be used instead of `registry_uid` for better readability. If `uid` is not provided, this field can be used along with `name` and `tag` to resolve the pack UID internally. Either `registry_uid` or `registry_name` can be specified, but not both.
- `registry_uid` (String) The registry UID of the pack. The registry UID is the unique identifier of the registry. This attribute is required if there is more than one registry that contains a pack with the same name. If `uid` is not provided, this field is required along with `name` and `tag` to resolve the pack UID internally. Either `registry_uid` or `registry_name` can be specified, but not both.
- `tag` (String) The tag of the pack. The tag is the version of the pack. This attribute is required if the pack type is `spectro` or `helm`. If `uid` is not provided, this field is required along with `name` and `registry_uid` (or `registry_name`) to resolve the pack UID internally.
- `type` (String) The type of the pack. Allowed values are `spectro`, `manifest`, `helm`, or `oci`. The default value is spectro. If using an OCI registry for pack, set the type to `oci`.
- `uid` (String) The unique identifier of the pack. The value can be looked up using the [`spectrocloud_pack`](https://registry.terraform.io/providers/spectrocloud/spectrocloud/latest/docs/data-sources/pack) data source. This value is required if the pack type is `spectro` and for `helm` if the chart is from a public helm registry. If not provided, all of `name`, `tag`, and `registry_uid` must be specified to resolve the pack UID internally.
- `values` (String) The values of the pack. The values are the configuration values of the pack. The values are specified in YAML format.

<a id="nestedblock--cluster_profile--pack--manifest"></a>
### Nested Schema for `cluster_profile.pack.manifest`

Required:

- `content` (String) The content of the manifest. The content is the YAML content of the manifest.
- `name` (String) The name of the manifest. The name must be unique within the pack.

Read-Only:

- `uid` (String)




<a id="nestedblock--cluster_rbac_binding"></a>
### Nested Schema for `cluster_rbac_binding`

Required:

- `type` (String) The type of the RBAC binding. Can be one of the following values: `RoleBinding`, or `ClusterRoleBinding`.

Optional:

- `namespace` (String) The Kubernetes namespace of the RBAC binding. Required if 'type' is set to 'RoleBinding'.
- `role` (Map of String) The role of the RBAC binding. Required if 'type' is set to 'RoleBinding'.
- `subjects` (Block List) (see [below for nested schema](#nestedblock--cluster_rbac_binding--subjects))

<a id="nestedblock--cluster_rbac_binding--subjects"></a>
### Nested Schema for `cluster_rbac_binding.subjects`

Required:

- `name` (String) The name of the subject. Required if 'type' is set to 'User' or 'Group'.
- `type` (String) The type of the subject. Can be one of the following values: `User`, `Group`, or `ServiceAccount`.

Optional:

- `namespace` (String) The Kubernetes namespace of the subject. Required if 'type' is set to 'ServiceAccount'.



<a id="nestedblock--host_config"></a>
### Nested Schema for `host_config`

Optional:

- `external_traffic_policy` (String) The external traffic policy for the cluster.
- `host_endpoint_type` (String) The type of endpoint for the cluster. Can be either 'Ingress' or 'LoadBalancer'. The default is 'Ingress'.
- `ingress_host` (String) The host for the Ingress endpoint. Required if 'host_endpoint_type' is set to 'Ingress'.
- `load_balancer_source_ranges` (String) The source ranges for the load balancer. Required if 'host_endpoint_type' is set to 'LoadBalancer'.


<a id="nestedblock--namespaces"></a>
### Nested Schema for `namespaces`

Required:

- `name` (String) Name of the namespace. This is the name of the Kubernetes namespace in the cluster.
- `resource_allocation` (Map of String) Resource allocation for the namespace. This is a map containing the resource type and the resource value. For example, `{cpu_cores: '2', memory_MiB: '2048', gpu_limit: '1', gpu_provider: 'nvidia'}`


<a id="nestedblock--scan_policy"></a>
### Nested Schema for `scan_policy`

Required:

- `configuration_scan_schedule` (String) The schedule for configuration scan.
- `conformance_scan_schedule` (String) The schedule for conformance scan.
- `penetration_scan_schedule` (String) The schedule for penetration scan.


<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `delete` (String)
- `update` (String)


<a id="nestedatt--location_config"></a>
### Nested Schema for `location_config`

Read-Only:

- `country_code` (String)
- `country_name` (String)
- `latitude` (Number)
- `longitude` (Number)
- `region_code` (String)
- `region_name` (String)


---
page_title: "spectrocloud_cluster_apache_cloudstack Resource - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  Resource for managing Apache CloudStack clusters in Spectro Cloud through Palette.
---

# spectrocloud_cluster_apache_cloudstack (Resource)

  Resource for managing Apache CloudStack clusters in Spectro Cloud through Palette.

## Example Usage

### Basic Apache CloudStack Cluster

```terraform
data "spectrocloud_cloudaccount_apache_cloudstack" "account" {
  name = "apache-cloudstack-account-1"
}

data "spectrocloud_cluster_profile" "profile" {
  name = "cloudstack-k8s-profile"
}

resource "spectrocloud_cluster_apache_cloudstack" "cluster" {
  name             = "apache-cloudstack-cluster-1"
  tags             = ["dev", "cloudstack", "department:engineering"]
  cloud_account_id = data.spectrocloud_cloudaccount_apache_cloudstack.account.id

  cloud_config {
    ssh_key_name = "my-ssh-key"
    
    zone {
      name = "Zone1"
      
      network {
        id   = "network-id"  # or use name instead
        name = "DefaultNetwork"
      }
    }
  }

  cluster_profile {
    id = data.spectrocloud_cluster_profile.profile.id
  }

  # Control Plane Pool
  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "cp-pool"
    count                   = 1
    offering                = "Medium Instance"  # CloudStack compute offering

    network {
      network_name = "DefaultNetwork"
    }

    additional_labels = {
      "role" = "control-plane"
    }
  }

  # Worker Pool
  machine_pool {
    name     = "worker-pool"
    count    = 2
    offering = "Large Instance"  # CloudStack compute offering

    network {
      network_name = "DefaultNetwork"
    }

    additional_labels = {
      "role" = "worker"
    }
  }
}
```

### Apache CloudStack Cluster with Advanced Configuration

```terraform
data "spectrocloud_cloudaccount_apache_cloudstack" "account" {
  name = "apache-cloudstack-account-1"
}

data "spectrocloud_cluster_profile" "profile" {
  name = "cloudstack-k8s-profile"
}

data "spectrocloud_backup_storage_location" "bsl" {
  name = "s3-backup-location"
}

resource "spectrocloud_cluster_apache_cloudstack" "advanced_cluster" {
  name             = "apache-cloudstack-production"
  tags             = ["prod", "cloudstack", "department:platform"]
  cloud_account_id = data.spectrocloud_cloudaccount_apache_cloudstack.account.id

  # Update all worker pools simultaneously for faster updates
  update_worker_pools_in_parallel = true

  cloud_config {
    ssh_key_name = "production-ssh-key"
    
    # Optional: Specify CloudStack project
    project {
      id   = "project-uuid"        # CloudStack project ID
      name = "ProductionProject"   # CloudStack project name
    }
    
    zone {
      name = "Zone1"
      
      network {
        id   = "network-id"
        name = "ProductionNetwork"
      }
    }
  }

  cluster_profile {
    id = data.spectrocloud_cluster_profile.profile.id

    # Override cluster profile variables
    variables = {
      "cluster_size" = "large"
      "environment"  = "production"
    }
  }

  # Backup Policy
  backup_policy {
    schedule                  = "0 2 * * *"  # Daily at 2 AM
    backup_location_id        = data.spectrocloud_backup_storage_location.bsl.id
    prefix                    = "cloudstack-prod-backup"
    expiry_in_hour            = 7200  # 300 days
    include_disks             = true
    include_cluster_resources = true
  }

  # Scan Policy
  scan_policy {
    configuration_scan_schedule = "0 0 * * SUN"
    penetration_scan_schedule   = "0 0 * * SUN"
    conformance_scan_schedule   = "0 0 1 * *"
  }

  # Control Plane Pool with Custom Instance Configuration
  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "cp-pool"
    count                   = 3

    placement {
      zone         = "Zone1"
      compute      = "Medium Instance"
      network_name = "ProductionNetwork"
    }

    # Custom instance configuration
    instance_config {
      disk_gib   = 100
      memory_mib = 8192
      num_cpus   = 4
    }

    additional_labels = {
      "role"        = "control-plane"
      "environment" = "production"
    }

    taints {
      key    = "master"
      value  = "true"
      effect = "NoSchedule"
    }
  }

  # Worker Pool with Autoscaling
  machine_pool {
    name = "worker-pool-autoscale"
    count = 3
    min  = 2
    max  = 10

    placement {
      zone         = "Zone1"
      compute      = "Large Instance"
      network_name = "ProductionNetwork"
    }

    instance_config {
      disk_gib   = 200
      memory_mib = 16384
      num_cpus   = 8
    }

    additional_labels = {
      "role"        = "worker"
      "scalable"    = "true"
      "environment" = "production"
    }

    # Optional: Additional annotations for worker pool nodes
    additional_annotations = {
      "custom.io/annotation"                        = "production-workload"
      "cluster-autoscaler.kubernetes.io/enabled"    = "true"
    }

    # Optional: Override kubeadm configuration for worker nodes
    # This is only supported for worker pools (not control plane)
    override_kubeadm_configuration = <<-EOT
      kubeletExtraArgs:
        node-labels: "env=production,tier=backend"
        max-pods: "110"
      preKubeadmCommands:
        - echo 'Starting node customization'
        - sysctl -w net.ipv4.ip_forward=1
      postKubeadmCommands:
        - echo 'Node customization complete'
        - systemctl restart kubelet
    EOT

    update_strategy      = "RollingUpdateScaleOut"
    node_repave_interval = 90
  }

  timeouts {
    create = "45m"
    update = "45m"
    delete = "30m"
  }
}
```

### Apache CloudStack Cluster with Static IP Pool

```terraform
data "spectrocloud_cloudaccount_apache_cloudstack" "account" {
  name = "apache-cloudstack-account-1"
}

data "spectrocloud_cluster_profile" "profile" {
  name = "cloudstack-k8s-profile"
}

resource "spectrocloud_cluster_apache_cloudstack" "cluster_static_ip" {
  name             = "apache-cloudstack-static-ip"
  tags             = ["prod", "static-ip"]
  cloud_account_id = data.spectrocloud_cloudaccount_apache_cloudstack.account.id

  cloud_config {
    ssh_key_name = "prod-ssh-key"
    
    zone {
      name = "Zone1"
      
      network {
        id   = "network-id"
        name = "StaticIPNetwork"
      }
    }
  }

  cluster_profile {
    id = data.spectrocloud_cluster_profile.profile.id
  }

  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "cp-pool"
    count                   = 1

    placement {
      zone              = "Zone1"
      compute           = "Medium Instance"
      network_name      = "StaticIPNetwork"
      static_ip_pool_id = "static-ip-pool-uuid"
    }
  }

  machine_pool {
    name  = "worker-pool"
    count = 2

    placement {
      zone              = "Zone1"
      compute           = "Large Instance"
      network_name      = "StaticIPNetwork"
      static_ip_pool_id = "static-ip-pool-uuid"
    }
  }
}
```

### Apache CloudStack Cluster with Template Override

```terraform
data "spectrocloud_cloudaccount_apache_cloudstack" "account" {
  name = "apache-cloudstack-account-1"
}

data "spectrocloud_cluster_profile" "profile" {
  name = "cloudstack-k8s-profile"
}

resource "spectrocloud_cluster_apache_cloudstack" "cluster_custom_template" {
  name             = "apache-cloudstack-custom-template"
  tags             = ["dev", "custom-os"]
  cloud_account_id = data.spectrocloud_cloudaccount_apache_cloudstack.account.id

  cloud_config {
    ssh_key_name = "dev-ssh-key"
    
    zone {
      name = "Zone1"
      
      network {
        id   = "network-id"
        name = "DefaultNetwork"
      }
    }
  }

  cluster_profile {
    id = data.spectrocloud_cluster_profile.profile.id
  }

  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "cp-pool"
    count                   = 1

    placement {
      zone         = "Zone1"
      compute      = "Medium Instance"
      network_name = "DefaultNetwork"
    }

    # Override CloudStack template for this machine pool
    template {
      name = "ubuntu-22.04-custom-template"
    }
  }

  machine_pool {
    name  = "worker-pool"
    count = 2

    placement {
      zone         = "Zone1"
      compute      = "Large Instance"
      network_name = "DefaultNetwork"
    }

    # Different template for worker nodes
    template {
      name = "ubuntu-22.04-gpu-template"
    }

    instance_config {
      disk_gib   = 500
      memory_mib = 32768
      num_cpus   = 16
    }
  }
}
```

### Apache CloudStack Cluster with Cluster Template

This example shows how to create a cluster using a cluster template instead of a cluster profile. Cluster templates provide a standardized way to deploy clusters with predefined configurations.

```terraform
data "spectrocloud_cloudaccount_apache_cloudstack" "account" {
  name = "apache-cloudstack-account-1"
}

data "spectrocloud_cluster_config_template" "template" {
  name = "apache-cloudstack-standard-template"
}

resource "spectrocloud_cluster_apache_cloudstack" "cluster_from_template" {
  name             = "apache-cloudstack-template-cluster"
  tags             = ["prod", "template"]
  cloud_account_id = data.spectrocloud_cloudaccount_apache_cloudstack.account.id

  cloud_config {
    ssh_key_name = "prod-ssh-key"
    
    zone {
      name = "Zone1"
      
      network {
        id   = "network-id"
        name = "ProductionNetwork"
      }
    }
  }

  # Use cluster_template instead of cluster_profile
  cluster_template {
    id = data.spectrocloud_cluster_config_template.template.id

    # Optional: Override variables for specific profiles in the template
    cluster_profile {
      id = "profile-uid-1"
      variables = {
        "k8s_version" = "1.28.0"
        "replicas"    = "3"
      }
    }

    cluster_profile {
      id = "profile-uid-2"
      variables = {
        "namespace" = "production"
      }
    }
  }

  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "cp-pool"
    count                   = 3
    offering                = "Large Instance"

    network {
      network_name = "ProductionNetwork"
    }
  }

  machine_pool {
    name     = "worker-pool"
    count    = 5
    min      = 3
    max      = 10
    offering = "XLarge Instance"

    network {
      network_name = "ProductionNetwork"
    }

    additional_labels = {
      "workload" = "production"
      "tier"     = "backend"
    }

    # Optional: Additional annotations
    additional_annotations = {
      "team.company.io/owner" = "platform-engineering"
    }
    
    update_strategy = "RollingUpdateScaleOut"
  }
}
```

### Apache CloudStack Cluster with Kubeadm Configuration Override

This example demonstrates how to customize kubeadm configuration for worker nodes. This is useful for setting custom kubelet arguments, running pre/post kubeadm commands, and other node-level customizations.

```terraform
data "spectrocloud_cloudaccount_apache_cloudstack" "account" {
  name = "apache-cloudstack-account-1"
}

data "spectrocloud_cluster_profile" "profile" {
  name = "cloudstack-k8s-profile"
}

resource "spectrocloud_cluster_apache_cloudstack" "cluster_custom_kubeadm" {
  name             = "apache-cloudstack-custom-kubeadm"
  tags             = ["prod", "customized"]
  cloud_account_id = data.spectrocloud_cloudaccount_apache_cloudstack.account.id

  cloud_config {
    ssh_key_name = "prod-ssh-key"
    
    zone {
      name = "Zone1"
      
      network {
        id   = "network-id"
        name = "ProductionNetwork"
      }
    }
  }

  cluster_profile {
    id = data.spectrocloud_cluster_profile.profile.id
  }

  machine_pool {
    control_plane           = true
    control_plane_as_worker = false
    name                    = "cp-pool"
    count                   = 3

    placement {
      zone         = "Zone1"
      compute      = "Large Instance"
      network_name = "ProductionNetwork"
    }

    additional_labels = {
      "role" = "control-plane"
    }
  }

  # Worker pool with custom kubeadm configuration
  machine_pool {
    name  = "worker-pool-frontend"
    count = 3

    placement {
      zone         = "Zone1"
      compute      = "XLarge Instance"
      network_name = "ProductionNetwork"
    }

    additional_labels = {
      "role" = "worker"
      "tier" = "frontend"
    }

    additional_annotations = {
      "custom.io/workload-type" = "web-facing"
    }

    # Customize kubeadm configuration for this worker pool
    # Note: override_kubeadm_configuration is only supported for worker pools
    override_kubeadm_configuration = <<-EOT
      kubeletExtraArgs:
        node-labels: "tier=frontend,zone=dmz"
        max-pods: "110"
        eviction-hard: "memory.available<500Mi,nodefs.available<10%"
        feature-gates: "RotateKubeletServerCertificate=true"
      preKubeadmCommands:
        - echo 'Configuring frontend worker node'
        - sysctl -w net.core.somaxconn=32768
        - sysctl -w net.ipv4.ip_local_port_range="1024 65535"
        - sysctl -w net.ipv4.tcp_tw_reuse=1
      postKubeadmCommands:
        - echo 'Frontend worker node setup complete'
        - systemctl restart kubelet
    EOT

    update_strategy      = "RollingUpdateScaleOut"
    node_repave_interval = 60  # Repave nodes every 60 days
  }

  # Worker pool for backend services
  machine_pool {
    name  = "worker-pool-backend"
    count = 5
    min   = 3
    max   = 10

    placement {
      zone         = "Zone1"
      compute      = "XXLarge Instance"
      network_name = "ProductionNetwork"
    }

    instance_config {
      disk_gib   = 500
      memory_mib = 65536
      num_cpus   = 16
    }

    additional_labels = {
      "role" = "worker"
      "tier" = "backend"
    }

    additional_annotations = {
      "custom.io/workload-type"                     = "backend-services"
      "cluster-autoscaler.kubernetes.io/enabled"    = "true"
    }

    # Different kubeadm configuration for backend workers
    override_kubeadm_configuration = <<-EOT
      kubeletExtraArgs:
        node-labels: "tier=backend,compute=high"
        max-pods: "200"
        kube-reserved: "cpu=1,memory=2Gi,ephemeral-storage=1Gi"
        system-reserved: "cpu=500m,memory=1Gi,ephemeral-storage=1Gi"
      preKubeadmCommands:
        - echo 'Configuring backend worker node'
        - sysctl -w vm.max_map_count=262144
        - sysctl -w fs.file-max=2097152
      postKubeadmCommands:
        - echo 'Backend worker node setup complete'
    EOT

    update_strategy      = "RollingUpdateScaleIn"
    node_repave_interval = 90
  }

  timeouts {
    create = "60m"
    update = "60m"
    delete = "30m"
  }
}
```

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import)
to import the resource spectrocloud_cluster_apache_cloudstack by using its `id` with the Palette `context` separated by a colon. For example:

```terraform
import {
  to = spectrocloud_cluster_apache_cloudstack.example
  id = "example_id:context"
}
```

Using `terraform import`, import the cluster using the `cluster_apache_cloudstack_name` or  `id` colon separated with `context`. For example:

```console
terraform import spectrocloud_cluster_apache_cloudstack.{cluster_uid}/{cluster_name}:project
```


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
- `cluster_profile` (Block Set) (see [below for nested schema](#nestedblock--cluster_profile))
- `cluster_rbac_binding` (Block List) The RBAC binding for the cluster. (see [below for nested schema](#nestedblock--cluster_rbac_binding))
- `cluster_template` (Block List, Max: 1) The cluster template of the cluster. (see [below for nested schema](#nestedblock--cluster_template))
- `cluster_timezone` (String) Defines the time zone used by this cluster to interpret scheduled operations. Maintenance tasks like upgrades will follow this time zone to ensure they run at the appropriate local time for the cluster. Must be in IANA timezone format (e.g., 'America/New_York', 'Asia/Kolkata', 'Europe/London').
- `context` (String) The context of the CloudStack configuration. Allowed values are `project` or `tenant`. Default is `project`. If  the `project` context is specified, the project name will sourced from the provider configuration parameter [`project_name`](https://registry.terraform.io/providers/spectrocloud/spectrocloud/latest/docs#schema).
- `description` (String) The description of the cluster. Default value is empty string.
- `force_delete` (Boolean) If set to `true`, the cluster will be force deleted and user has to manually clean up the provisioned cloud resources.
- `force_delete_delay` (Number) Delay duration in minutes to before invoking cluster force delete. Default and minimum is 20.
- `host_config` (Block List) The host configuration for the cluster. (see [below for nested schema](#nestedblock--host_config))
- `namespaces` (Block List) The namespaces for the cluster. (see [below for nested schema](#nestedblock--namespaces))
- `os_patch_after` (String) The date and time after which to patch the cluster. Prefix the time value with the respective RFC. Ex: `RFC3339: 2006-01-02T15:04:05Z07:00`
- `os_patch_on_boot` (Boolean) Whether to apply OS patch on boot. Default is `false`.
- `os_patch_schedule` (String) Cron schedule for OS patching. This must be in the form of `0 0 * * *`.
- `pause_agent_upgrades` (String) The pause agent upgrades setting allows to control the automatic upgrade of the Palette component and agent for an individual cluster. The default value is `unlock`, meaning upgrades occur automatically. Setting it to `lock` pauses automatic agent upgrades for the cluster.
- `review_repave_state` (String) To authorize the cluster repave, set the value to `Approved` for approval and `""` to decline. Default value is `""`.
- `scan_policy` (Block List, Max: 1) The scan policy for the cluster. (see [below for nested schema](#nestedblock--scan_policy))
- `skip_completion` (Boolean) If `true`, the cluster will be created asynchronously. Default value is `false`.
- `tags` (Set of String) A list of tags to be applied to the cluster. Tags must be in the form of `key:value`.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `update_worker_pools_in_parallel` (Boolean) Controls whether worker pool updates occur in parallel or sequentially. When set to `true` (default), all worker pools are updated simultaneously. When `false`, worker pools are updated one at a time, reducing cluster disruption but taking longer to complete updates.

### Read-Only

- `admin_kube_config` (String) Admin Kube-config for the cluster. This can be used to connect to the cluster using `kubectl`, With admin privilege.
- `cloud_config_id` (String, Deprecated) ID of the cloud config used for the cluster. This cloud config must be of type `cloudstack`.
- `id` (String) The ID of this resource.
- `kubeconfig` (String) Kubeconfig for the cluster. This can be used to connect to the cluster using `kubectl`.
- `location_config` (List of Object) The location of the cluster. (see [below for nested schema](#nestedatt--location_config))

<a id="nestedblock--cloud_config"></a>
### Nested Schema for `cloud_config`

Required:

- `zone` (Block List, Min: 1) List of CloudStack zones for multi-AZ deployments. If only one zone is specified, it will be treated as single-zone deployment. (see [below for nested schema](#nestedblock--cloud_config--zone))

Optional:

- `control_plane_endpoint` (String) Endpoint IP to be used for the API server. Should only be set for static CloudStack networks.
- `project` (Block List, Max: 1) CloudStack project configuration (optional). If not specified, the cluster will be created in the domain's default project. (see [below for nested schema](#nestedblock--cloud_config--project))
- `ssh_key_name` (String) SSH key name for accessing cluster nodes.
- `sync_with_cks` (Boolean) Determines if an external managed CKS (CloudStack Kubernetes Service) cluster should be created. Default is `false`.

<a id="nestedblock--cloud_config--zone"></a>
### Nested Schema for `cloud_config.zone`

Required:

- `name` (String) CloudStack zone name where the cluster will be deployed.

Optional:

- `id` (String) CloudStack zone ID. Either `id` or `name` can be used to identify the zone. If both are specified, `id` takes precedence.
- `network` (Block List, Max: 1) Network configuration for this zone. (see [below for nested schema](#nestedblock--cloud_config--zone--network))

<a id="nestedblock--cloud_config--zone--network"></a>
### Nested Schema for `cloud_config.zone.network`

Required:

- `name` (String) Network name in this zone.

Optional:

- `gateway` (String) Gateway IP address for the network.
- `id` (String) Network ID in CloudStack. Either `id` or `name` can be used to identify the network. If both are specified, `id` takes precedence.
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
- `id` (String) VPC ID. Either `id` or `name` can be used to identify the VPC. If both are specified, `id` takes precedence.
- `offering` (String) VPC offering name.




<a id="nestedblock--cloud_config--project"></a>
### Nested Schema for `cloud_config.project`

Optional:

- `id` (String) CloudStack project ID.
- `name` (String) CloudStack project name.



<a id="nestedblock--machine_pool"></a>
### Nested Schema for `machine_pool`

Required:

- `count` (Number) Number of nodes in the machine pool.
- `name` (String) Name of the machine pool.
- `offering` (String) Apache CloudStack compute offering (instance type/size) name.

Optional:

- `additional_annotations` (Map of String) Additional annotation to be applied to the machine pool. annotation must be in the form of `key:value`.
- `additional_labels` (Map of String) Additional labels to be applied to the machine pool. Labels must be in the form of `key:value`.
- `control_plane` (Boolean) Whether this machine pool is a control plane. Defaults to `false`.
- `control_plane_as_worker` (Boolean) Whether this machine pool is a control plane and a worker. Defaults to `false`.
- `max` (Number) Maximum number of nodes in the machine pool. This is used for autoscaling.
- `min` (Number) Minimum number of nodes in the machine pool. This is used for autoscaling.
- `network` (Block List) Network configuration for the machine pool instances. (see [below for nested schema](#nestedblock--machine_pool--network))
- `node` (Block List) (see [below for nested schema](#nestedblock--machine_pool--node))
- `node_repave_interval` (Number) Minimum number of seconds node should be Ready, before the next node is selected for repave. Default value is `0`, Applicable only for worker pools.
- `override_kubeadm_configuration` (String) YAML config for kubeletExtraArgs, preKubeadmCommands, postKubeadmCommands. Overrides pack-level settings. Worker pools only.
- `override_scaling` (Block List, Max: 1) Rolling update strategy for the machine pool. (see [below for nested schema](#nestedblock--machine_pool--override_scaling))
- `taints` (Block List) (see [below for nested schema](#nestedblock--machine_pool--taints))
- `template` (Block List, Max: 1) Apache CloudStack template override for this machine pool. If not specified, inherits cluster default. (see [below for nested schema](#nestedblock--machine_pool--template))
- `update_strategy` (String) Update strategy for the machine pool. Valid values are `RollingUpdateScaleOut`, `RollingUpdateScaleIn` and `OverrideScaling`. If `OverrideScaling` is used, `override_scaling` must be specified with both `max_surge` and `max_unavailable`.

Read-Only:

- `instance_config` (List of Object) Instance configuration details returned by the CloudStack API. This is a computed field based on the selected offering. (see [below for nested schema](#nestedatt--machine_pool--instance_config))

<a id="nestedblock--machine_pool--network"></a>
### Nested Schema for `machine_pool.network`

Required:

- `network_name` (String) Network name to attach to the machine pool.

Optional:

- `ip_address` (String, Deprecated) Static IP address to assign. **DEPRECATED**: This field is no longer supported by CloudStack and will be ignored.


<a id="nestedblock--machine_pool--node"></a>
### Nested Schema for `machine_pool.node`

Required:

- `action` (String) The action to perform on the node. Valid values are: `cordon`, `uncordon`.
- `node_id` (String) The node_id of the node, For example `i-07f899a33dee624f7`


<a id="nestedblock--machine_pool--override_scaling"></a>
### Nested Schema for `machine_pool.override_scaling`

Optional:

- `max_surge` (String) Max extra nodes during rolling update. Integer or percentage (e.g., '1' or '20%'). Only valid when type=OverrideScaling. Both maxSurge and maxUnavailable are required.
- `max_unavailable` (String) Max unavailable nodes during rolling update. Integer or percentage (e.g., '0' or '10%'). Only valid when type=OverrideScaling. Both maxSurge and maxUnavailable are required.


<a id="nestedblock--machine_pool--taints"></a>
### Nested Schema for `machine_pool.taints`

Required:

- `effect` (String) The effect of the taint. Allowed values are: `NoSchedule`, `PreferNoSchedule` or `NoExecute`.
- `key` (String) The key of the taint.
- `value` (String) The value of the taint.


<a id="nestedblock--machine_pool--template"></a>
### Nested Schema for `machine_pool.template`

Optional:

- `id` (String) Template ID. Either ID or name must be provided.
- `name` (String) Template name. Either ID or name must be provided.


<a id="nestedatt--machine_pool--instance_config"></a>
### Nested Schema for `machine_pool.instance_config`

Read-Only:

- `category` (String)
- `cpu_set` (Number)
- `disk_gib` (Number)
- `memory_mib` (Number)
- `name` (String)
- `num_cpus` (Number)



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
- `role` (Map of String) The role of the RBAC binding. Required if 'type' is set to 'RoleBinding'. Must include 'name' and 'kind' fields.
- `subjects` (Block List) (see [below for nested schema](#nestedblock--cluster_rbac_binding--subjects))

<a id="nestedblock--cluster_rbac_binding--subjects"></a>
### Nested Schema for `cluster_rbac_binding.subjects`

Required:

- `name` (String) The name of the subject. Required if 'type' is set to 'User' or 'Group'.
- `type` (String) The type of the subject. Can be one of the following values: `User`, `Group`, or `ServiceAccount`.

Optional:

- `namespace` (String) The Kubernetes namespace of the subject. Required if 'type' is set to 'ServiceAccount'.



<a id="nestedblock--cluster_template"></a>
### Nested Schema for `cluster_template`

Required:

- `id` (String) The ID of the cluster template.

Optional:

- `cluster_profile` (Block Set) The cluster profile of the cluster template. (see [below for nested schema](#nestedblock--cluster_template--cluster_profile))

Read-Only:

- `name` (String) The name of the cluster template.

<a id="nestedblock--cluster_template--cluster_profile"></a>
### Nested Schema for `cluster_template.cluster_profile`

Required:

- `id` (String) The UID of the cluster profile.

Optional:

- `variables` (Map of String) A map of cluster profile variables, specified as key-value pairs. For example: `priority = "5"`.



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


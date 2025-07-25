---
page_title: "spectrocloud_cluster_vsphere Resource - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  A resource to manage a vSphere cluster in Palette.
---

# spectrocloud_cluster_vsphere (Resource)

  A resource to manage a vSphere cluster in Palette.

## Example Usage

```terraform
data "spectrocloud_cluster_profile" "vmware_profile" {
  name    = "vsphere-picard-2"
  version = "1.0.0"
  context = "project"
}
data "spectrocloud_cloudaccount_vsphere" "vmware_account" {
  name = var.shared_vmware_cloud_account_name
}


resource "spectrocloud_cluster_vsphere" "cluster" {
  name = "vsphere-picard-3"
  # For Force Delete enforcement
  # force_delete = true
  # force_delete_delay = 25
  cloud_account_id = data.spectrocloud_cloudaccount_vsphere.vmware_account.id
  cluster_profile {
    id = data.spectrocloud_cluster_profile.vmware_profile.id
  }
  cloud_config {
    ssh_key = var.cluster_ssh_public_key

    datacenter = var.vsphere_datacenter
    folder     = var.vsphere_folder
    // For Dynamic DNS (network_type & network_search_domain value should set for DDNS)
    network_type          = "DDNS"
    network_search_domain = var.cluster_network_search
    // For Static (By Default static_ip is false, for static provisioning, it is set to be true. Not required to specify network_type & network_search_domain)
    # static_ip = true
  }

  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "cp-pool"
    count                   = 1
    placement {
      cluster       = var.vsphere_cluster
      resource_pool = var.vsphere_resource_pool
      datastore     = var.vsphere_datastore
      network       = var.vsphere_network
    }
    instance_type {
      disk_size_gb = 40
      memory_mb    = 4096
      cpu          = 2
    }
  }

  machine_pool {
    name                 = "worker-basic"
    count                = 1
    node_repave_interval = 30
    placement {
      cluster       = var.vsphere_cluster
      resource_pool = var.vsphere_resource_pool
      datastore     = var.vsphere_datastore
      network       = var.vsphere_network
    }
    instance_type {
      disk_size_gb = 40
      memory_mb    = 8192
      cpu          = 4
    }
  }
}
```

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import)
to import the resource spectrocloud_cluster_vsphere by using its `id` with the Palette `context` separated by a colon. For example:

```terraform
import {
  to = spectrocloud_cluster_vsphere.example
  id = "example_id:context"
}
```

Using `terraform import`, import the cluster using the `id` colon separated with `context`. For example:

```console
terraform import spectrocloud_cluster_vsphere.example example_id:project
```

Refer to the [Import section](/docs#import) to learn more.

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `cloud_account_id` (String) ID of the cloud account to be used for the cluster. This cloud account must be of type `vsphere`.
- `cloud_config` (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--cloud_config))
- `machine_pool` (Block Set, Min: 1) (see [below for nested schema](#nestedblock--machine_pool))
- `name` (String) The name of the cluster.

### Optional

- `apply_setting` (String) The setting to apply the cluster profile. `DownloadAndInstall` will download and install packs in one action. `DownloadAndInstallLater` will only download artifact and postpone install for later. Default value is `DownloadAndInstall`.
- `backup_policy` (Block List, Max: 1) The backup policy for the cluster. If not specified, no backups will be taken. (see [below for nested schema](#nestedblock--backup_policy))
- `cluster_meta_attribute` (String) `cluster_meta_attribute` can be used to set additional cluster metadata information, eg `{'nic_name': 'test', 'env': 'stage'}`
- `cluster_profile` (Block List) (see [below for nested schema](#nestedblock--cluster_profile))
- `cluster_rbac_binding` (Block List) The RBAC binding for the cluster. (see [below for nested schema](#nestedblock--cluster_rbac_binding))
- `context` (String) The context of the VMware cluster. Allowed values are `project` or `tenant`. Default is `project`. If  the `project` context is specified, the project name will sourced from the provider configuration parameter [`project_name`](https://registry.terraform.io/providers/spectrocloud/spectrocloud/latest/docs#schema).
- `description` (String) The description of the cluster. Default value is empty string.
- `force_delete` (Boolean) If set to `true`, the cluster will be force deleted and user has to manually clean up the provisioned cloud resources.
- `force_delete_delay` (Number) Delay duration in minutes to before invoking cluster force delete. Default and minimum is 20.
- `host_config` (Block List) The host configuration for the cluster. (see [below for nested schema](#nestedblock--host_config))
- `location_config` (Block List) (see [below for nested schema](#nestedblock--location_config))
- `namespaces` (Block List) The namespaces for the cluster. (see [below for nested schema](#nestedblock--namespaces))
- `os_patch_after` (String) The date and time after which to patch the cluster. Prefix the time value with the respective RFC. Ex: `RFC3339: 2006-01-02T15:04:05Z07:00`
- `os_patch_on_boot` (Boolean) Whether to apply OS patch on boot. Default is `false`.
- `os_patch_schedule` (String) The cron schedule for OS patching. This must be in the form of cron syntax. Ex: `0 0 * * *`.
- `pause_agent_upgrades` (String) The pause agent upgrades setting allows to control the automatic upgrade of the Palette component and agent for an individual cluster. The default value is `unlock`, meaning upgrades occur automatically. Setting it to `lock` pauses automatic agent upgrades for the cluster.
- `review_repave_state` (String) To authorize the cluster repave, set the value to `Approved` for approval and `""` to decline. Default value is `""`.
- `scan_policy` (Block List, Max: 1) The scan policy for the cluster. (see [below for nested schema](#nestedblock--scan_policy))
- `skip_completion` (Boolean) If `true`, the cluster will be created asynchronously. Default value is `false`.
- `tags` (Set of String) A list of tags to be applied to the cluster. Tags must be in the form of `key:value`.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `admin_kube_config` (String) Admin Kube-config for the cluster. This can be used to connect to the cluster using `kubectl`, With admin privilege.
- `cloud_config_id` (String, Deprecated) ID of the cloud config used for the cluster. This cloud config must be of type `azure`.
- `id` (String) The ID of this resource.
- `kubeconfig` (String) Kubeconfig for the cluster. This can be used to connect to the cluster using `kubectl`.

<a id="nestedblock--cloud_config"></a>
### Nested Schema for `cloud_config`

Required:

- `datacenter` (String) The name of the datacenter in vSphere. This is the name of the datacenter as it appears in vSphere.
- `folder` (String) The name of the folder in vSphere. This is the name of the folder as it appears in vSphere.

Optional:

- `host_endpoint` (String) The host endpoint to use for the cluster. This can be `IP` or `FQDN(External/DDNS)`.
- `image_template_folder` (String) The name of the image template folder in vSphere. This is the name of the folder as it appears in vSphere.
- `network_search_domain` (String) The search domain to use for the cluster in case of DHCP.
- `network_type` (String) The type of network to use for the cluster. This can be `VIP` or `DDNS`.
- `ntp_servers` (Set of String) A list of NTP servers to be used by the cluster.
- `ssh_key` (String, Deprecated) The SSH key to be used for the cluster. This is the public key that will be used to access the cluster nodes. `ssh_key & ssh_keys` are mutually exclusive.
- `ssh_keys` (Set of String) List of public SSH (Secure Shell) keys to establish, administer, and communicate with remote clusters, `ssh_key & ssh_keys` are mutually exclusive.
- `static_ip` (Boolean) Whether to use static IP addresses for the cluster. If `true`, the cluster will use static IP addresses. If `false`, the cluster will use DDNS. Default is `false`.


<a id="nestedblock--machine_pool"></a>
### Nested Schema for `machine_pool`

Required:

- `count` (Number) Number of nodes in the machine pool.
- `instance_type` (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--machine_pool--instance_type))
- `name` (String) The name of the machine pool. This is used to identify the machine pool in the cluster.
- `placement` (Block List, Min: 1) (see [below for nested schema](#nestedblock--machine_pool--placement))

Optional:

- `additional_labels` (Map of String)
- `control_plane` (Boolean) Whether this machine pool is a control plane. Defaults to `false`.
- `control_plane_as_worker` (Boolean) Whether this machine pool is a control plane and a worker. Defaults to `false`.
- `max` (Number) Maximum number of nodes in the machine pool. This is used for autoscaling the machine pool.
- `min` (Number) Minimum number of nodes in the machine pool. This is used for autoscaling the machine pool.
- `node` (Block List) (see [below for nested schema](#nestedblock--machine_pool--node))
- `node_repave_interval` (Number) Minimum number of seconds node should be Ready, before the next node is selected for repave. Default value is `0`, Applicable only for worker pools.
- `taints` (Block List) (see [below for nested schema](#nestedblock--machine_pool--taints))
- `update_strategy` (String) Update strategy for the machine pool. Valid values are `RollingUpdateScaleOut` and `RollingUpdateScaleIn`.

<a id="nestedblock--machine_pool--instance_type"></a>
### Nested Schema for `machine_pool.instance_type`

Required:

- `cpu` (Number) The number of CPUs.
- `disk_size_gb` (Number) The size of the disk in GB.
- `memory_mb` (Number) The amount of memory in MB.


<a id="nestedblock--machine_pool--placement"></a>
### Nested Schema for `machine_pool.placement`

Required:

- `cluster` (String) The name of the cluster to use for the machine pool. As it appears in the vSphere.
- `datastore` (String) The name of the datastore to use for the machine pool. As it appears in the vSphere.
- `network` (String) The name of the network to use for the machine pool. As it appears in the vSphere.
- `resource_pool` (String) The name of the resource pool to use for the machine pool. As it appears in the vSphere.

Optional:

- `static_ip_pool_id` (String) The ID of the static IP pool to use for the machine pool in case of static cluster placement.

Read-Only:

- `id` (String) The ID of this resource.


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
- `registry_uid` (String) The registry UID of the pack. The registry UID is the unique identifier of the registry. This attribute is required if there is more than one registry that contains a pack with the same name. If `uid` is not provided, this field is required along with `name` and `tag` to resolve the pack UID internally.
- `tag` (String) The tag of the pack. The tag is the version of the pack. This attribute is required if the pack type is `spectro` or `helm`. If `uid` is not provided, this field is required along with `name` and `registry_uid` to resolve the pack UID internally.
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


<a id="nestedblock--location_config"></a>
### Nested Schema for `location_config`

Required:

- `latitude` (Number) The latitude coordinates value.
- `longitude` (Number) The longitude coordinates value.

Optional:

- `country_code` (String) The country code of the country the cluster is located in.
- `country_name` (String) The name of the country.
- `region_code` (String) The region code of where the cluster is located in.
- `region_name` (String) The name of the region.


<a id="nestedblock--namespaces"></a>
### Nested Schema for `namespaces`

Required:

- `name` (String) Name of the namespace. This is the name of the Kubernetes namespace in the cluster.
- `resource_allocation` (Map of String) Resource allocation for the namespace. This is a map containing the resource type and the resource value. For example, `{cpu_cores: '2', memory_MiB: '2048'}`

Optional:

- `images_blacklist` (List of String) List of images to disallow for the namespace. For example, `['nginx:latest', 'redis:latest']`


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
---
page_title: "spectrocloud_cluster_maas Resource - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  
---

# Resource `spectrocloud_cluster_maas`



## Example Usage

```terraform
data "spectrocloud_cloudaccount_maas" "account" {
  # id = <uid>
  name = var.cluster_cloud_account_name
}

data "spectrocloud_cluster_profile" "profile" {
  # id = <uid>
  name = var.cluster_cluster_profile_name
}

data "spectrocloud_backup_storage_location" "bsl" {
  name = var.backup_storage_location_name
}

resource "spectrocloud_cluster_maas" "cluster" {
  name             = var.cluster_name
  tags             = ["dev", "department:devops", "owner:bob"]
  cloud_account_id = data.spectrocloud_cloudaccount_maas.account.id

  cloud_config {
    subscription_id = "subscription-id"
    resource_group  = "dev"
    ssh_key         = "ssh key value"
    region          = "centralus"
  }

  cluster_profile {
    id = data.spectrocloud_cluster_profile.profile.id

    # To override or specify values for a cluster:

    # pack {
    #   name   = "spectro-byo-manifest"
    #   tag    = "1.0.x"
    #   values = <<-EOT
    #     manifests:
    #       byo-manifest:
    #         contents: |
    #           # Add manifests here
    #           apiVersion: v1
    #           kind: Namespace
    #           metadata:
    #             labels:
    #               app: wordpress
    #               app2: wordpress2
    #             name: wordpress
    #   EOT
    # }
  }

  backup_policy {
    schedule                  = "0 0 * * SUN"
    backup_location_id        = data.spectrocloud_backup_storage_location.bsl.id
    prefix                    = "prod-backup"
    expiry_in_hour            = 7200
    include_disks             = true
    include_cluster_resources = true
  }

  scan_policy {
    configuration_scan_schedule = "0 0 * * SUN"
    penetration_scan_schedule   = "0 0 * * SUN"
    conformance_scan_schedule   = "0 0 1 * *"
  }

  machine_pool {
    name                 = "worker-basic"
    count                = 1
    instance_type        = "Standard_DS4"
    disk_size_gb         = 60
    is_system_node_pool  = true
    storage_account_type = "Standard_LRS"
  }
}
```

## Schema

### Required

- **cloud_config** (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--cloud_config))
- **machine_pool** (Block Set, Min: 1) (see [below for nested schema](#nestedblock--machine_pool))
- **name** (String)

### Optional

- **backup_policy** (Block List, Max: 1) (see [below for nested schema](#nestedblock--backup_policy))
- **cloud_account_id** (String)
- **cluster_profile** (Block List) (see [below for nested schema](#nestedblock--cluster_profile))
- **id** (String) The ID of this resource.
- **os_patch_after** (String)
- **os_patch_on_boot** (Boolean)
- **os_patch_schedule** (String)
- **scan_policy** (Block List, Max: 1) (see [below for nested schema](#nestedblock--scan_policy))
- **tags** (Set of String)
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-only

- **cloud_config_id** (String)
- **kubeconfig** (String)

<a id="nestedblock--cloud_config"></a>
### Nested Schema for `cloud_config`

Required:

- **domain** (String)


<a id="nestedblock--machine_pool"></a>
### Nested Schema for `machine_pool`

Required:

- **azs** (Set of String)
- **count** (Number)
- **instance_type** (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--machine_pool--instance_type))
- **name** (String)
- **placement** (Block List, Min: 1) (see [below for nested schema](#nestedblock--machine_pool--placement))

Optional:

- **control_plane** (Boolean)
- **control_plane_as_worker** (Boolean)
- **update_strategy** (String)

<a id="nestedblock--machine_pool--instance_type"></a>
### Nested Schema for `machine_pool.instance_type`

Required:

- **min_cpu** (Number)
- **min_memory_mb** (Number)


<a id="nestedblock--machine_pool--placement"></a>
### Nested Schema for `machine_pool.placement`

Required:

- **resource_pool** (String)

Read-only:

- **id** (String) The ID of this resource.



<a id="nestedblock--backup_policy"></a>
### Nested Schema for `backup_policy`

Required:

- **backup_location_id** (String)
- **expiry_in_hour** (Number)
- **prefix** (String)
- **schedule** (String)

Optional:

- **include_cluster_resources** (Boolean)
- **include_disks** (Boolean)
- **namespaces** (Set of String)


<a id="nestedblock--cluster_profile"></a>
### Nested Schema for `cluster_profile`

Required:

- **id** (String) The ID of this resource.

Optional:

- **pack** (Block List) (see [below for nested schema](#nestedblock--cluster_profile--pack))

<a id="nestedblock--cluster_profile--pack"></a>
### Nested Schema for `cluster_profile.pack`

Required:

- **name** (String)
- **tag** (String)
- **values** (String)



<a id="nestedblock--scan_policy"></a>
### Nested Schema for `scan_policy`

Required:

- **configuration_scan_schedule** (String)
- **conformance_scan_schedule** (String)
- **penetration_scan_schedule** (String)


<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **create** (String)
- **delete** (String)
- **update** (String)



---
page_title: "spectrocloud_cluster_azure Resource - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  
---

# Resource `spectrocloud_cluster_azure`



## Example Usage

```terraform
data "spectrocloud_cloudaccount_azure" "account" {
  # id = <uid>
  name = var.cluster_cloud_account_name
}

data "spectrocloud_cluster_profile" "profile" {
  # id = <uid>
  name = var.cluster_cluster_profile_name
}

resource "spectrocloud_cluster_azure" "cluster" {
  name               = var.cluster_name
  cluster_profile_id = data.spectrocloud_cluster_profile.profile.id
  cloud_account_id   = data.spectrocloud_cloudaccount_azure.account.id

  cloud_config {
    subscription_id = var.azure_subscription_id
    resource_group  = var.azure_resource_group
    region          = var.azure_region
    ssh_key         = var.cluster_ssh_public_key
  }

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

  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "master-pool"
    count                   = 1
    instance_type           = "Standard_D2_v3"
    disk {
      size_gb = 65
      type    = "Standard_LRS"
    }
  }

  machine_pool {
    name          = "worker-basic"
    count         = 1
    instance_type = "Standard_D2_v3"
  }

}
```

## Schema

### Required

- **cloud_account_id** (String)
- **cloud_config** (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--cloud_config))
- **cluster_profile_id** (String)
- **machine_pool** (Block Set, Min: 1) (see [below for nested schema](#nestedblock--machine_pool))
- **name** (String)

### Optional

- **id** (String) The ID of this resource.
- **pack** (Block Set) (see [below for nested schema](#nestedblock--pack))
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- **os_patch_on_boot** (Boolean, Optional) OS Patch on boot when set, updates security patch of host OS 
of all nodes and monitors new nodes (which gets created when cluster is scaled up or cluster k8s version is upgraded) for security patch
- **os_patch_schedule** (String, Optional) Cron schedule to patch security updates on host OS for all nodes. Please see https://en.wikipedia.org/wiki/Cron for valid cron syntax
- **os_patch_after** (String, Optional) On demand security patch on host OS for all nodes. Please follow RFC3339 Date and Time Standards. Eg 2021-01-01T00:00:00.000Z

### Read-only

- **cloud_config_id** (String)
- **kubeconfig** (String)

<a id="nestedblock--cloud_config"></a>
### Nested Schema for `cloud_config`

Required:

- **region** (String)
- **resource_group** (String)
- **ssh_key** (String)
- **subscription_id** (String)


<a id="nestedblock--machine_pool"></a>
### Nested Schema for `machine_pool`

Required:

- **count** (Number)
- **instance_type** (String)
- **name** (String)

Optional:

- **control_plane** (Boolean)
- **control_plane_as_worker** (Boolean)
- **disk** (Block List, Max: 1) (see [below for nested schema](#nestedblock--machine_pool--disk))
- **update_strategy** (String)

<a id="nestedblock--machine_pool--disk"></a>
### Nested Schema for `machine_pool.disk`

Required:

- **size_gb** (Number)
- **type** (String)



<a id="nestedblock--pack"></a>
### Nested Schema for `pack`

Required:

- **name** (String)
- **tag** (String)
- **values** (String)


<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **create** (String)
- **delete** (String)
- **update** (String)



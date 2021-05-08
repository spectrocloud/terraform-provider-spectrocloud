---
page_title: "spectrocloud_cluster_gcp Resource - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  
---

# Resource `spectrocloud_cluster_gcp`



## Example Usage

```terraform
data "spectrocloud_cloudaccount_gcp" "account" {
  # id = <uid>
  name = var.cluster_cloud_account_name
}

data "spectrocloud_cluster_profile" "profile" {
  # id = <uid>
  name = var.cluster_cluster_profile_name
}


resource "spectrocloud_cluster_gcp" "cluster" {
  name               = var.cluster_name
  cluster_profile_id = data.spectrocloud_cluster_profile.profile.id
  cloud_account_id   = data.spectrocloud_cloudaccount_gcp.account.id

  cloud_config {
    network = var.gcp_network
    project = var.gcp_project
    region  = var.gcp_region
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
    instance_type           = "e2-standard-2"
    disk_size_gb            = 62
    azs                     = ["us-west3-a"]
  }

  machine_pool {
    name          = "worker-basic"
    count         = 1
    instance_type = "e2-standard-2"
    azs           = ["us-west3-a"]
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
- **os_patch_after** (String)
- **os_patch_on_boot** (Boolean)
- **os_patch_schedule** (String)
- **pack** (Block List) (see [below for nested schema](#nestedblock--pack))
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-only

- **cloud_config_id** (String)
- **kubeconfig** (String)

<a id="nestedblock--cloud_config"></a>
### Nested Schema for `cloud_config`

Required:

- **project** (String)
- **region** (String)

Optional:

- **network** (String)


<a id="nestedblock--machine_pool"></a>
### Nested Schema for `machine_pool`

Required:

- **azs** (Set of String)
- **count** (Number)
- **instance_type** (String)
- **name** (String)

Optional:

- **control_plane** (Boolean)
- **control_plane_as_worker** (Boolean)
- **disk_size_gb** (Number)
- **update_strategy** (String)


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



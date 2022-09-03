---
page_title: "spectrocloud_cluster_eks Resource - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  
---

# Resource `spectrocloud_cluster_eks`
[Spectro Cloud EKS Static Placement Example](examples/e2e/eks-static)
[Spectro Cloud EKS Dynamic Placement Example](examples/e2e/eks)

## Example Usage

```terraform
data "spectrocloud_cloudaccount_aws" "account" {
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

resource "spectrocloud_cluster_eks" "cluster" {
  name             = var.cluster_name
  tags             = ["dev", "department:devops", "owner:bob"]
  cloud_account_id = data.spectrocloud_cloudaccount_aws.account.id

  cloud_config {
    ssh_key_name = "default"
    region       = "us-west-2"
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
    name          = "worker-basic"
    count         = 1
    instance_type = "t3.large"
    az_subnets = {
      "us-west-2a" = "subnet-0d4978ddbff16c"
    }
  }

}
```

## Schema

### Required

- **cloud_account_id** (String)
- **cloud_config** (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--cloud_config))
- **machine_pool** (Block List, Min: 1) (see [below for nested schema](#nestedblock--machine_pool))
- **name** (String)

### Optional

- **apply_setting** (String)
- **backup_policy** (Block List, Max: 1) (see [below for nested schema](#nestedblock--backup_policy))
- **cluster_profile** (Block List) (see [below for nested schema](#nestedblock--cluster_profile))
- **cluster_profile_id** (String, Deprecated)
- **cluster_rbac_binding** (Block List) (see [below for nested schema](#nestedblock--cluster_rbac_binding))
- **fargate_profile** (Block List) (see [below for nested schema](#nestedblock--fargate_profile))
- **id** (String) The ID of this resource.
- **namespaces** (Block List) (see [below for nested schema](#nestedblock--namespaces))
- **os_patch_after** (String)
- **os_patch_on_boot** (Boolean)
- **os_patch_schedule** (String)
- **pack** (Block List) (see [below for nested schema](#nestedblock--pack))
- **scan_policy** (Block List, Max: 1) (see [below for nested schema](#nestedblock--scan_policy))
- **tags** (Set of String)
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-only

- **cloud_config_id** (String)
- **kubeconfig** (String)

<a id="nestedblock--cloud_config"></a>
### Nested Schema for `cloud_config`

Required:

- **region** (String)

Optional:

- **az_subnets** (Map of String) Mutually exclusive with `azs`. Use for Static provisioning.
- **azs** (List of String) Mutually exclusive with `az_subnets`. Use for Dynamic provisioning.
- **encryption_config_arn** (String)
- **endpoint_access** (String)
- **public_access_cidrs** (Set of String)
- **ssh_key_name** (String)
- **vpc_id** (String)


<a id="nestedblock--machine_pool"></a>
### Nested Schema for `machine_pool`

Required:

- **count** (Number)
- **disk_size_gb** (Number)
- **instance_type** (String)
- **name** (String)

Optional:

- **additional_labels** (Map of String)
- **az_subnets** (Map of String)
- **azs** (List of String)
- **capacity_type** (String)
- **max** (Number)
- **max_price** (String)
- **min** (Number)
- **taints** (Block List) (see [below for nested schema](#nestedblock--machine_pool--taints))
- **update_strategy** (String)

<a id="nestedblock--machine_pool--taints"></a>
### Nested Schema for `machine_pool.taints`

Required:

- **effect** (String)
- **key** (String)
- **value** (String)



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

Optional:

- **manifest** (Block List) (see [below for nested schema](#nestedblock--cluster_profile--pack--manifest))
- **registry_uid** (String)
- **tag** (String)
- **type** (String)
- **values** (String)

<a id="nestedblock--cluster_profile--pack--manifest"></a>
### Nested Schema for `cluster_profile.pack.manifest`

Required:

- **content** (String)
- **name** (String)




<a id="nestedblock--cluster_rbac_binding"></a>
### Nested Schema for `cluster_rbac_binding`

Required:

- **type** (String)

Optional:

- **namespace** (String)
- **role** (Map of String)
- **subjects** (Block List) (see [below for nested schema](#nestedblock--cluster_rbac_binding--subjects))

<a id="nestedblock--cluster_rbac_binding--subjects"></a>
### Nested Schema for `cluster_rbac_binding.subjects`

Required:

- **name** (String)
- **type** (String)

Optional:

- **namespace** (String)



<a id="nestedblock--fargate_profile"></a>
### Nested Schema for `fargate_profile`

Required:

- **name** (String)
- **selector** (Block List, Min: 1) (see [below for nested schema](#nestedblock--fargate_profile--selector))

Optional:

- **additional_tags** (Map of String)
- **subnets** (List of String)

<a id="nestedblock--fargate_profile--selector"></a>
### Nested Schema for `fargate_profile.selector`

Required:

- **namespace** (String)

Optional:

- **labels** (Map of String)



<a id="nestedblock--namespaces"></a>
### Nested Schema for `namespaces`

Required:

- **name** (String)
- **resource_allocation** (Map of String)


<a id="nestedblock--pack"></a>
### Nested Schema for `pack`

Required:

- **name** (String)
- **tag** (String)
- **values** (String)

Optional:

- **registry_uid** (String)


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



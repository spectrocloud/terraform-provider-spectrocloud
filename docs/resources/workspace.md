---
page_title: "spectrocloud_workspace Resource - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  
---

# Resource `spectrocloud_workspace`



## Example Usage

```terraform
data "spectrocloud_cluster" "cluster1" {
  name = "vsphere-picard-2"
}

resource "spectrocloud_workspace" "workspace" {
  name = "wsp-tf"

  clusters {
    uid = data.spectrocloud_cluster.cluster1.id
  }

  cluster_rbac_binding {
    type = "ClusterRoleBinding"

    role = {
      kind = "ClusterRole"
      name = "testrole3"
    }
    subjects {
      type = "User"
      name = "testRoleUser4"
    }
    subjects {
      type = "Group"
      name = "testRoleGroup4"
    }
    subjects {
      type      = "ServiceAccount"
      name      = "testrolesubject3"
      namespace = "testrolenamespace"
    }
  }

  cluster_rbac_binding {
    type      = "RoleBinding"
    namespace = "test5ns"
    role = {
      kind = "Role"
      name = "testrolefromns3"
    }
    subjects {
      type = "User"
      name = "testUserRoleFromNS3"
    }
    subjects {
      type = "Group"
      name = "testGroupFromNS3"
    }
    subjects {
      type      = "ServiceAccount"
      name      = "testrolesubject3"
      namespace = "testrolenamespace"
    }
  }

  namespaces {
    name = "test5ns"
    resource_allocation = {
      cpu_cores  = "2"
      memory_MiB = "2048"
    }

    images_blacklist = ["1", "2", "3"]
  }

  backup_policy {
    schedule                  = "0 0 * * SUN"
    backup_location_id        = data.spectrocloud_backup_storage_location.bsl.id
    prefix                    = "prod-backup"
    expiry_in_hour            = 7200
    include_disks             = false
    include_cluster_resources = true

    //namespaces = ["test5ns"]
    include_all_clusters = true
    cluster_uids         = [data.spectrocloud_cluster.cluster1.id]
  }

}

data "spectrocloud_backup_storage_location" "bsl" {
  name = "backups-nikolay"
}
```

## Schema

### Required

- **clusters** (Block Set, Min: 1) (see [below for nested schema](#nestedblock--clusters))
- **name** (String)

### Optional

- **backup_policy** (Block List, Max: 1) (see [below for nested schema](#nestedblock--backup_policy))
- **cluster_rbac_binding** (Block List) (see [below for nested schema](#nestedblock--cluster_rbac_binding))
- **description** (String)
- **id** (String) The ID of this resource.
- **namespaces** (Block List) (see [below for nested schema](#nestedblock--namespaces))
- **tags** (Set of String)

<a id="nestedblock--clusters"></a>
### Nested Schema for `clusters`

Required:

- **uid** (String)


<a id="nestedblock--backup_policy"></a>
### Nested Schema for `backup_policy`

Required:

- **backup_location_id** (String)
- **expiry_in_hour** (Number)
- **prefix** (String)
- **schedule** (String)

Optional:

- **cluster_uids** (Set of String)
- **include_all_clusters** (Boolean)
- **include_cluster_resources** (Boolean)
- **include_disks** (Boolean)
- **include_workspace_resources** (Boolean)
- **namespaces** (Set of String)


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



<a id="nestedblock--namespaces"></a>
### Nested Schema for `namespaces`

Required:

- **name** (String)
- **resource_allocation** (Map of String)

Optional:

- **images_blacklist** (List of String)


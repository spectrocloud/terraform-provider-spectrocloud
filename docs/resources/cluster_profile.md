---
page_title: "spectrocloud_cluster_profile Resource - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  
---

# Resource `spectrocloud_cluster_profile`



## Example Usage

```terraform
# If looking up a cluster profile instead of creating a new one
# data "spectrocloud_cluster_profile" "profile" {
#   # id = <uid>
#   name = var.cluster_cluster_profile_name
# }

# # Example of a Basic add-on profile
# resource "spectrocloud_cluster_profile" "cp-addon-vsphere" {
#   name        = "cp-basic"
#   description = "basic cp"
#   cloud       = "vsphere"
#   type        = "add-on"
#   pack {
#     name = "spectro-byo-manifest"
#     tag  = "1.0.x"
#     uid  = "5faad584f244cfe0b98cf489"
#     # layer  = ""
#     values = <<-EOT
#       manifests:
#         byo-manifest:
#           contents: |
#             # Add manifests here
#             apiVersion: v1
#             kind: Namespace
#             metadata:
#               labels:
#                 app: wordpress
#                 app3: wordpress3
#               name: wordpress
#     EOT
#   }
# }


data "spectrocloud_pack" "csi" {
  name = "csi-vsphere-volume"
  # version  = "1.0.x"
}

data "spectrocloud_pack" "cni" {
  name    = "cni-calico"
  version = "3.16.0"
}

data "spectrocloud_pack" "k8s" {
  name    = "kubernetes"
  version = "1.18.16"
}

data "spectrocloud_pack" "ubuntu" {
  name = "ubuntu-vsphere"
  # version  = "1.0.x"
}

resource "spectrocloud_cluster_profile" "profile" {
  name        = "vsphere-picard-3"
  description = "basic cp"
  cloud       = "vsphere"
  type        = "cluster"

  pack {
    name   = "ubuntu-vsphere"
    tag    = "LTS__18.4.x"
    uid    = data.spectrocloud_pack.ubuntu.id
    values = "foo: 1"
  }

  pack {
    name   = "kubernetes"
    tag    = "1.18.16"
    uid    = data.spectrocloud_pack.k8s.id
    values = data.spectrocloud_pack.k8s.values
  }

  pack {
    name   = "cni-calico"
    tag    = "3.16.x"
    uid    = data.spectrocloud_pack.cni.id
    values = data.spectrocloud_pack.cni.values
  }

  pack {
    name   = "csi-vsphere-volume"
    tag    = "1.0.x"
    uid    = data.spectrocloud_pack.csi.id
    values = data.spectrocloud_pack.csi.values
  }

  pack {
    name = "manifest-namespace"
    type = "manifest"
    manifest {
      name    = "manifest-namespace"
      content = <<-EOT
        apiVersion: v1
        kind: Namespace
        metadata:
          labels:
            app: wordpress
            app3: wordpress786
          name: wordpress
      EOT
    }
    #uid    = "spectro-manifest-pack"
  }
}
```

## Schema

### Required

- **name** (String)
- **pack** (Block List, Min: 1) (see [below for nested schema](#nestedblock--pack))

### Optional

- **cloud** (String)
- **description** (String)
- **id** (String) The ID of this resource.
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- **type** (String)

<a id="nestedblock--pack"></a>
### Nested Schema for `pack`

Required:

- **name** (String)

Optional:

- **manifest** (Block List) (see [below for nested schema](#nestedblock--pack--manifest))
- **tag** (String)
- **type** (String)
- **uid** (String)
- **values** (String)

<a id="nestedblock--pack--manifest"></a>
### Nested Schema for `pack.manifest`

Required:

- **content** (String)
- **name** (String)

Read-only:

- **uid** (String)



<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **create** (String)
- **delete** (String)
- **update** (String)



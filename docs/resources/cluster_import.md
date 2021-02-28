---
page_title: "spectrocloud_cluster_import Resource - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  
---

# Resource `spectrocloud_cluster_import`



## Example Usage

```terraform
data "spectrocloud_cluster_profile" "profile" {
  # id = <uid>
  name = var.cluster_cluster_profile_name
}
resource "spectrocloud_cluster_import" "cluster" {
  name = "vsphere-import-01"
  cloud = "vsphere"
  cluster_profile_id = spectrocloud_cluster_profile.profile.id
/*  pack {
     name   = "k8s-dashboard"
     tag    = "2.1.x"
     values = <<-EOT
        manifests:
          k8s-dashboard:
            #Namespace to install kubernetes-dashboard
            namespace: "kubernetes-dashboard"
            #The ClusterRole to assign for kubernetes-dashboard. By default, a ready-only cluster role is provisioned
            clusterRole: "k8s-dashboard-readonly"
            #Self-Signed Certificate duration in hours
            certDuration: 9000h
            #Self-Signed Certificate renewal in hours
            certRenewal: 720h     #30d
            #The service type for dashboard. Supported values are ClusterIP / LoadBalancer / NodePort
            serviceType: ClusterIP
            #Flag to enable skip login option on the dashboard login page
            skipLogin: false
            #Ingress config
            ingress:
              enabled: false
     EOT
  }*/
}
```

## Schema

### Required

- **cloud** (String)
- **name** (String)

### Optional

- **cluster_profile_id** (String) The ID of this resource.
- **pack** (Block Set) (see [below for nested schema](#nestedblock--pack))
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-only

- **cloud_config_id** (String)
- **cluster_import_manifest_apply_command** (String)
- **cluster_import_manifest** (String)


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

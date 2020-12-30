---
page_title: "spectrocloud_cluster_profile Resource - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  
---

# Resource `spectrocloud_cluster_profile`



## Example Usage

```terraform
resource "spectrocloud_cluster_profile" "cp-addon-azure" {
  name        = "cp-basic"
  description = "basic cp"
  cloud       = "azure"
  type        = "add-on"

  pack {
    name = "spectro-byo-manifest"
    tag  = "1.0.x"
    uid  = "5faad584f244cfe0b98cf489"
    # layer  = ""
    values = <<-EOT
      manifests:
        byo-manifest:
          contents: |
            # Add manifests here
            apiVersion: v1
            kind: Namespace
            metadata:
              labels:
                app: wordpress
                app3: wordpress3
              name: wordpress
    EOT
  }

}
```

## Schema

### Required

- **cloud** (String)
- **name** (String)
- **pack** (Block List, Min: 1) (see [below for nested schema](#nestedblock--pack))

### Optional

- **description** (String)
- **id** (String) The ID of this resource.
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- **type** (String)

<a id="nestedblock--pack"></a>
### Nested Schema for `pack`

Required:

- **name** (String)
- **tag** (String)
- **uid** (String)
- **values** (String)


<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **create** (String)
- **delete** (String)
- **update** (String)



---
page_title: "spectrocloud_cluster_profile Resource - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  The Cluster Profile resource allows you to create and manage cluster profiles.
---

# spectrocloud_cluster_profile (Resource)

  The Cluster Profile resource allows you to create and manage cluster profiles.

## Example Usage



### Example of a Cluster Profile with Data Resources and Custom YAML

In the following example, the cluster profile references data resources for the pack information. The pack information is used to define the cluster profile. The cluster profile also references a YAML file that contains the manifest information.

```terraform

data "spectrocloud_registry" "public_registry" {
  name = "Public Repo"
}

data "spectrocloud_cloudaccount_aws" "account" {
  count = var.deploy-aws ? 1 : 0
  name  = var.aws-cloud-account-name
}

data "spectrocloud_pack" "aws_csi" {
  name         = "csi-aws-ebs"
  version      = "1.22.0"
  registry_uid = data.spectrocloud_registry.public_registry.id
}

data "spectrocloud_pack" "aws_cni" {
  name         = "cni-calico"
  version      = "3.26.1"
  registry_uid = data.spectrocloud_registry.public_registry.id
}

data "spectrocloud_pack" "aws_k8s" {
  name         = "kubernetes"
  version      = "1.27.5"
  registry_uid = data.spectrocloud_registry.public_registry.id
}

data "spectrocloud_pack" "aws_ubuntu" {
  name         = "ubuntu-aws"
  version      = "22.04"
  registry_uid = data.spectrocloud_registry.public_registry.id
}

data "spectrocloud_cluster" "aws_cluster_api" {
  count = var.deploy-aws ? 1 : 0

  name    = "aws-cluster-api"
  context = "project"

  depends_on = [spectrocloud_cluster_aws.aws-cluster-api]
}


resource "spectrocloud_cluster_profile" "aws-profile" {
  count = var.deploy-aws ? 1 : 0

  name        = "tf-aws-profile"
  description = "A basic cluster profile for AWS"
  tags        = concat(var.tags, ["env:aws"])
  cloud       = "aws"
  type        = "cluster"
  version     = "1.0.0"

  pack {
    name   = data.spectrocloud_pack.aws_ubuntu.name
    tag    = data.spectrocloud_pack.aws_ubuntu.version
    uid    = data.spectrocloud_pack.aws_ubuntu.id
    values = data.spectrocloud_pack.aws_ubuntu.values
  }

  pack {
    name   = data.spectrocloud_pack.aws_k8s.name
    tag    = data.spectrocloud_pack.aws_k8s.version
    uid    = data.spectrocloud_pack.aws_k8s.id
    values = data.spectrocloud_pack.aws_k8s.values
  }

  pack {
    name   = data.spectrocloud_pack.aws_cni.name
    tag    = data.spectrocloud_pack.aws_cni.version
    uid    = data.spectrocloud_pack.aws_cni.id
    values = data.spectrocloud_pack.aws_cni.values
  }

  pack {
    name   = data.spectrocloud_pack.aws_csi.name
    tag    = data.spectrocloud_pack.aws_csi.version
    uid    = data.spectrocloud_pack.aws_csi.id
    values = data.spectrocloud_pack.aws_csi.values
  }

  pack {
    name   = "hello-universe"
    type   = "manifest"
    tag    = "1.0.0"
    values = ""
    manifest {
      name    = "hello-universe"
      content = file("manifests/hello-universe.yaml")
    }
  }
}
```

### Inline YAML Example

An example of a cluster profile using inline YAML.

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
  name    = "csi-vsphere-csi"
  version = "2.3.0"
}

data "spectrocloud_pack" "cni" {
  name    = "cni-calico"
  version = "3.16.0"
}

data "spectrocloud_pack" "k8s" {
  name    = "kubernetes"
  version = "1.21.5"
}

data "spectrocloud_pack" "ubuntu" {
  name    = "ubuntu-vsphere"
  version = "18.04"
}

locals {
  proxy_val = <<-EOT
        manifests:
          spectro-proxy:
            namespace: "cluster-{{ .spectro.system.cluster.uid }}"

            server: "{{ .spectro.system.reverseproxy.server }}"

            # Cluster UID - DO NOT CHANGE (new3)
            clusterUid: "{{ .spectro.system.cluster.uid }}"
            subdomain: "cluster-{{ .spectro.system.cluster.uid }}"
  EOT
}

resource "spectrocloud_cluster_profile" "profile" {
  name        = "vsphere-picard-3"
  description = "basic cp"
  tags        = ["dev", "department:devops", "owner:bob"]
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
    tag    = "1.21.5"
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
    name   = "csi-vsphere-csi"
    tag    = "2.3.x"
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

  pack {
    name   = "spectro-proxy"
    tag    = "1.0.0"
    uid    = "60bd99ce9c10082ed8b314c9"
    values = local.proxy_val
  }
  /*
  # profile_variables are currently supported only for edge-native cloud type and add-on profile type only
  profile_variables{
    variable {
      name = "default_password"
      display_name = "Default Password"
      format = "string"
      hidden = true // For sensitive variables like passwords, setting hidden to true will mask the variable value.
    }
    variable {
      name = "default_version"
      display_name = "Version"
      format = "version"
      description = "description hard-version"
      default_value = "0.0.1"
      regex = "*.*"
      required = true
      immutable = false
    }
  }
  profile_variables {
    variable {
      default_value = "test2"
      description   = null
      display_name  = "Type List"
      format        = "string"
      hidden        = false
      immutable     = false
      input_type    = "dropdown"
      is_sensitive  = false
      name          = "type_list"
      regex         = null
      required      = false
      options {
        description = "test 1 description"
        label       = "test1"
        value       = "value1"
      }
      options {
        description = "test 2 description"
        label       = "test2"
        value       = "value2"
      }
    }
    variable {
      default_value = <<-EOT
      sdfsdfdsf
      sdfsdf
      sdfdsf
      EOT      
      description   = null
      display_name  = "Type Multiline"
      format        = "string"
      hidden        = false
      immutable     = false
      input_type    = "multiline"
      is_sensitive  = false
      name          = "test_multiline"
      regex         = null
      required      = false
    }
  }
  */
}
```


### Example of Providing Multiple Packs

You can provide multiple packs at once by leveraging a dynamic block.  

!> The order of the Packs must be taken into consideration so avoid any situations where the order of the Packs is re-arranged by Terraform, such as nested loops `for_each = { for pack in var.packs : pack => pack }`. We recommend creating a variable that contains the list of Packs arranged in order of the App Profile stack and their respective configuration.

```terraform
resource "spectrocloud_cluster_profile" "this" {
  name = "security-profile"
  dynamic "pack" {
    for_each = var.my-packs
  }
```

### Profile Variables Example

An example of a cluster profile with profile variables. Refer to the [Profile Variables](#nested-schema-for-profile_variablesvariable) section for more information on the nested schema for profile variables.

```terraform
resource "spectrocloud_cluster_profile" "profile" {
  name        = "vsphere-picard-4"
  description = "basic cp"
  tags        = ["dev", "department:devops", "owner:bob"]
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
    tag    = "1.21.5"
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
    name   = "csi-vsphere-csi"
    tag    = "2.3.x"
    uid    = data.spectrocloud_pack.csi.id
    values = data.spectrocloud_pack.csi.values
  }


  profile_variables{
    variable {
      name = "default_password"
      display_name = "Default Password"
      format = "string"
      is_sensitive = true // For sensitive variables like passwords, setting hidden to true will mask the variable value.
    }
    variable {
      name = "default_version"
      display_name = "Version"
      format = "version"
      description = "description hard-version"
      default_value = "0.0.1"
      regex = "*.*"
      required = true
      immutable = false
    }
  }
}
```


## Immutable versioning

This resource supports the standard Terraform Plugin SDK v2 pattern for immutable-versioned upstream resources, the same shape used by [`aws_lambda_layer_version`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/lambda_layer_version), `aws_db_snapshot`, `azurerm_shared_image_version`, `google_compute_instance_template`, and similar resources. When the `immutable-clusterprofiles` feature_preview flag is enabled on the provider:

- Changes to `version` trigger a **Terraform resource replacement** (`ForceNew`) instead of an in-place update
- The new version is created by calling `POST /v1/clusterprofiles/{uid}/clone` against any existing version of the lineage, then overwriting the clone with the user's HCL content
- The previous version is preserved in Palette because `Delete` honors `skip_destroy = true`
- `create_before_destroy = true` ensures the new version exists before Terraform removes the old one from state

All three knobs (feature flag, `skip_destroy`, and `lifecycle`) are required together. Miss any one and the plan will either error out (missing `skip_destroy`) or the old version will be destroyed in Palette (missing `create_before_destroy`). This matches the standard SDK v2 pattern exactly.

### Enabling immutable versioning

```terraform
provider "spectrocloud" {
  host = "api.spectrocloud.com"

  feature_preview = {
    "immutable-clusterprofiles" = true
  }
}

resource "spectrocloud_cluster_profile" "addon" {
  name    = "example-addon"
  version = var.profile_version  # bump this via git tags / CI
  type    = "add-on"
  context = "project"

  pack {
    name = "example-manifest"
    type = "manifest"

    manifest {
      name    = "example"
      content = file("${path.module}/manifests/example.yaml")
    }
  }

  skip_destroy = true  # preserve old versions in Palette

  lifecycle {
    create_before_destroy = true  # create new version before destroying old
  }
}
```

Bumping `var.profile_version` from `1.0.0` to `1.1.0` and running `terraform apply`:

1. Terraform plans a replacement (ForceNew on `version`)
2. The new resource's `Create` runs first (because of `create_before_destroy`), calls the clone endpoint, applies the user's pack content
3. The old resource is "destroyed" from Terraform's perspective, but `skip_destroy` makes the Delete call a no-op; the `1.0.0` version stays in Palette untouched
4. State advances to `1.1.0`; Palette has both versions with distinct UIDs; `terraform output` returns the correct current uid without needing `terraform apply -refresh-only`

### Why all three knobs are required

| Knob | Why it's required | What happens if you skip it |
|------|------------------|----------------------------|
| `feature_preview.immutable-clusterprofiles = true` | Opts the resource into replacement-based versioning. Without it, the resource preserves its legacy in-place `PUT` behavior for backward compatibility. | Legacy in-place mutation; previous version is destroyed. |
| `skip_destroy = true` | Prevents `Delete` from calling the Palette DELETE API on the replaced resource, so the old version is preserved in Palette even though Terraform removes it from state. | Plan fails with an error at `terraform plan` time (the provider catches this and tells you exactly what to add). |
| `lifecycle { create_before_destroy = true }` | Ensures Terraform creates the new version before marking the old one for destruction. Without it, Terraform destroys first and creates second, causing a brief window where no version exists in Terraform state. | Terraform would attempt to `Delete` the old resource before `Create`ing the new one; because `skip_destroy = true` makes Delete a no-op, the Palette version still survives, but the planned operation order is wrong and some edge cases (state corruption on failure, for example) are riskier. Terraform core parses `lifecycle` blocks before the provider runs, so the provider cannot validate this automatically; this one is on the user to remember. |

### Backward compatibility

The `immutable-clusterprofiles` feature flag is opt-in. Without it, `spectrocloud_cluster_profile` behaves exactly as it did before the flag was introduced; version changes go through the in-place `PUT` path. Existing HCL, state files, and CI pipelines continue to work without modification. See the resource's `version` and `skip_destroy` field docs below for further details.


## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import)
to import the resource spectrocloud_cluster_profile by using its `id`. For example:

```terraform
import {
  to = spectrocloud_cluster_profile.example
  id = "example_id:context"
}
```

You can also use the Terraform CLI and the `terraform import`, command to import the cluster using by referencing the resource `id`. For example:

```console
% terraform import spectrocloud_cluster_profile.example example_id:project
```

Refer to the [Import section](/docs#import) to learn more.

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String)

### Optional

- `cloud` (String) Specify the infrastructure provider the cluster profile is for. Only Palette supported infrastructure providers can be used. The supported cloud types are - `all, aws, azure, gcp, vsphere, maas, virtual, baremetal, eks, aks, edge-native, generic, and gke` or any custom cloud provider registered in Palette, e.g., `nutanix`.If the value is set to `all`, then the type must be set to `add-on`. Otherwise, the cluster profile may be incompatible with other providers. Default value is `all`.
- `context` (String) The context of the cluster profile. Allowed values are `project` or `tenant`. Default value is `project`. If  the `project` context is specified, the project name will sourced from the provider configuration parameter [`project_name`](https://registry.terraform.io/providers/spectrocloud/spectrocloud/latest/docs#schema).
- `description` (String)
- `pack` (Block List) For packs of type `spectro`, `helm`, and `manifest`, at least one pack must be specified. (see [below for nested schema](#nestedblock--pack))
- `profile_variables` (Block List, Max: 1) List of variables for the cluster profile. (see [below for nested schema](#nestedblock--profile_variables))
- `skip_destroy` (Boolean) When `true`, `terraform destroy` removes the cluster profile from Terraform state without calling the Palette delete API, leaving the underlying profile version intact in Palette. 

This is the standard Terraform Plugin SDK v2 preservation pattern for immutable-versioned resources. Combined with the `immutable-clusterprofiles` feature_preview flag and `lifecycle { create_before_destroy = true }`, it lets you bump the `version` field as a normal in-HCL edit while every previous version stays preserved in Palette -- Terraform's state advances cleanly to the new version while older versions remain immutable in Palette. Defaults to `false`.
- `tags` (Set of String) A list of tags to be applied to the cluster. Tags must be in the form of `key:value`.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `type` (String) Specify the cluster profile type to use. Allowed values are `cluster`, `infra`, `add-on`, and `system`. These values map to the following User Interface (UI) labels. Use the value ' cluster ' for a **Full** cluster profile.For an Infrastructure cluster profile, use the value `infra`; for an Add-on cluster profile, use the value `add-on`.System cluster profiles can be specified using the value `system`. To learn more about cluster profiles, refer to the [Cluster Profile](https://docs.spectrocloud.com/cluster-profiles) documentation. Default value is `add-on`.
- `version` (String) Version of the cluster profile. Defaults to '1.0.0'. 

Default behavior (no feature flag set): changing this value on an existing profile updates the version in place via `PUT /v1/clusterprofiles/{uid}`, which destroys the previous version. This is the legacy behavior preserved for backward compatibility. 

When the `immutable-clusterprofiles` feature_preview flag is enabled, changing this value triggers a Terraform resource **replacement** (`ForceNew`) instead of an in-place update. This is the standard Terraform Plugin SDK v2 pattern for immutable-versioned resources. Combined with `skip_destroy = true` and `lifecycle { create_before_destroy = true }`, the new version is created by cloning from the existing Palette lineage while the previous version is preserved untouched in Palette. The Terraform resource id is set once at Create time and never mutates mid-update, so it respects the SDK v2 contract that a resource's primary id is stable across in-place updates -- outputs that reference `.id` always reflect the current version without needing `terraform apply -refresh-only`.

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--pack"></a>
### Nested Schema for `pack`

Required:

- `name` (String) The name of the pack. The name must be unique within the cluster profile.

Optional:

- `manifest` (Block List) (see [below for nested schema](#nestedblock--pack--manifest))
- `registry_name` (String) The registry name of the pack. The registry name is the human-readable name of the registry. This attribute can be used instead of `registry_uid` for better readability. If `uid` is not provided, this field can be used along with `name` and `tag` to resolve the pack UID internally. Either `registry_uid` or `registry_name` can be specified, but not both.
- `registry_uid` (String) The registry UID of the pack. The registry UID is the unique identifier of the registry. This attribute is required if there is more than one registry that contains a pack with the same name. If `uid` is not provided, this field is required along with `name` and `tag` to resolve the pack UID internally. Either `registry_uid` or `registry_name` can be specified, but not both.
- `tag` (String) The tag of the pack. The tag is the version of the pack. This attribute is required if the pack type is `spectro` or `helm`. If `uid` is not provided, this field is required along with `name` and `registry_uid` (or `registry_name`) to resolve the pack UID internally.
- `type` (String) The type of the pack. Allowed values are `spectro`, `manifest`, `helm`, or `oci`. The default value is spectro. If using an OCI registry for pack, set the type to `oci`.
- `uid` (String) The unique identifier of the pack. The value can be looked up using the [`spectrocloud_pack`](https://registry.terraform.io/providers/spectrocloud/spectrocloud/latest/docs/data-sources/pack) data source. This value is required if the pack type is `spectro` and for `helm` if the chart is from a public helm registry. If not provided, all of `name`, `tag`, and `registry_uid` must be specified to resolve the pack UID internally.
- `values` (String) The values of the pack. The values are the configuration values of the pack. The values are specified in YAML format.

<a id="nestedblock--pack--manifest"></a>
### Nested Schema for `pack.manifest`

Required:

- `content` (String) The content of the manifest. The content is the YAML content of the manifest.
- `name` (String) The name of the manifest. The name must be unique within the pack.

Read-Only:

- `uid` (String)



<a id="nestedblock--profile_variables"></a>
### Nested Schema for `profile_variables`

Required:

- `variable` (Block List, Min: 1) (see [below for nested schema](#nestedblock--profile_variables--variable))

<a id="nestedblock--profile_variables--variable"></a>
### Nested Schema for `profile_variables.variable`

Required:

- `display_name` (String) The display name of the variable should be unique among variables.
- `name` (String) The name of the variable should be unique among variables.

Optional:

- `default_value` (String) The default value of the variable. If the format is `multiline`, then the default value should be a multi-line string. If the input type is `dropdown`, then the default value should be a option label.
- `description` (String) The description of the variable.
- `format` (String) The format of the variable. Default is `string`, `format` field can only be set during cluster profile creation. Allowed formats include `string`, `number`, `boolean`, `ipv4`, `ipv4cidr`, `ipv6`, `version`, `base64`.
- `hidden` (Boolean) If `hidden` is set to `true`, then variable will be hidden for overriding the value. By default the `hidden` flag will be set to `false`.
- `immutable` (Boolean) If `immutable` is set to `true`, then variable value can't be editable. By default the `immutable` flag will be set to `false`.
- `input_type` (String) The input type of the variable. Defaults to `text` for backward compatibility. Allowed input types include `text`, `dropdown`, `multiline`.
- `is_sensitive` (Boolean) If `is_sensitive` is set to `true`, then default value will be masked. By default the `is_sensitive` flag will be set to false.
- `options` (Block List) The options of the variable. Only applicable for dropdown input type. (see [below for nested schema](#nestedblock--profile_variables--variable--options))
- `regex` (String) Regular expression pattern which the variable value must match.
- `required` (Boolean) The `required` to specify if the variable is optional or mandatory. If it is mandatory then default value must be provided.

<a id="nestedblock--profile_variables--variable--options"></a>
### Nested Schema for `profile_variables.variable.options`

Required:

- `label` (String) The label of the option.
- `value` (String) The value of the option.

Optional:

- `description` (String) The description of the option.

Read-Only:

- `default` (Boolean) The default value of the option.




<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `delete` (String)
- `update` (String)
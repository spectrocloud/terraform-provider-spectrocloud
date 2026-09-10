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

# Day-2 mutability: `cloud` and `type` are ForceNew - changing either recreates the profile.
# `version` is not a static ForceNew field, but a CustomizeDiff clones the profile into a new
# version whenever `version` changes, which behaves like a replacement from the caller's point of
# view. `name`, `tags`, `description`, `context`, `pack`, and `profile_variables` all update in
# place - note `name` is NOT ForceNew here, unlike most other resources in this provider.
resource "spectrocloud_cluster_profile" "profile" {
  name        = "vsphere-picard-3"
  description = "basic cp"
  tags        = ["dev", "department:devops", "owner:bob"]
  cloud       = "vsphere"
  type        = "cluster"

  pack {
    name   = "ubuntu-vsphere"
    tag    = data.spectrocloud_pack.ubuntu.version
    uid    = data.spectrocloud_pack.ubuntu.id
    values = "foo: 1"
  }

  pack {
    name   = "kubernetes"
    tag    = data.spectrocloud_pack.k8s.version
    uid    = data.spectrocloud_pack.k8s.id
    values = data.spectrocloud_pack.k8s.values
  }

  pack {
    name   = "cni-calico"
    tag    = data.spectrocloud_pack.cni.version
    uid    = data.spectrocloud_pack.cni.id
    values = data.spectrocloud_pack.cni.values
  }

  pack {
    name   = "csi-vsphere-csi"
    tag    = data.spectrocloud_pack.csi.version
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
  # profile_variables lets Day-2 consumers (e.g. spectrocloud_cluster's cluster_profile.variables,
  # or spectrocloud_cluster_config_template's per-cluster overrides) supply values that get
  # templated into pack manifests via `{{ .spectro.var.<name> }}`, without editing the profile
  # itself. At most one profile_variables block is allowed per profile - all variables go inside
  # its single `variable` list, not as multiple profile_variables blocks.
  profile_variables {
    variable {
      name         = "default_password"
      display_name = "Default Password"
      format       = "string"
      # For sensitive variables like passwords, hidden = true masks the value from being
      # overridden/viewed at the point of use.
      hidden = true
    }
    variable {
      name          = "default_version"
      display_name  = "Version"
      format        = "version"
      description   = "description hard-version"
      default_value = "0.0.1"
      regex         = "^\\d+\\.\\d+\\.\\d+$"
      required      = true
      immutable     = false
    }
    variable {
      default_value = "test2"
      display_name  = "Type List"
      format        = "string"
      # input_type = "dropdown" requires at least one options block; default_value must match
      # one of the option labels below.
      input_type = "dropdown"
      name       = "type_list"
      required   = false
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
      line one of the default value
      line two of the default value
      EOT
      display_name  = "Type Multiline"
      format        = "string"
      input_type    = "multiline"
      name          = "test_multiline"
      required      = false
    }
  }
}

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

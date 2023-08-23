# If looking up a cluster profile instead of creating a new one
# data "spectrocloud_cluster_profile" "profile" {
#   # id = <uid>
#   name = var.cluster_cluster_profile_name
# }

# # Example of a Basic add-on profile
# resource "spectrocloud_cluster_profile" "cp-addon-aws" {
#   name        = "cp-basic"
#   description = "basic cp"
#   cloud       = "aws"
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
data "spectrocloud_registry_pack" "registry" {
  name = "Public Repo"
}

data "spectrocloud_pack" "byom" {
  name = "spectro-byo-manifest"
  version  = "1.0.0"
  registry_uid = data.spectrocloud_registry_pack.registry.id
}

data "spectrocloud_pack" "csi" {
  name = "csi-aws-ebs"
  version  = "1.12.0"
  registry_uid = data.spectrocloud_registry_pack.registry.id
}

data "spectrocloud_pack" "cni" {
  name    = "cni-calico"
  version = "3.24.5"
  registry_uid = data.spectrocloud_registry_pack.registry.id
}

data "spectrocloud_pack" "k8s" {
  name    = "kubernetes"
  version = "1.26.1"
  registry_uid = data.spectrocloud_registry_pack.registry.id
}

data "spectrocloud_pack" "ubuntu" {
  name = "ubuntu-aws"
  version  = "22.04"
  registry_uid = data.spectrocloud_registry_pack.registry.id
}

resource "spectrocloud_cluster_profile" "profile" {
  name        = "aws-sample-2"
  description = "basic cp"
  tags        = ["dev", "department:devops", "owner:bob"]
  cloud       = "aws"
  type        = "cluster"
  context = "tenant"

  pack {
    name   = "spectro-byo-manifest"
    tag    = "1.0.x"
    uid    = data.spectrocloud_pack.byom.id
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

  pack {
    name   = "csi-aws-ebs"
    tag    = "1.12.x"
    uid    = data.spectrocloud_pack.csi.id
    values = data.spectrocloud_pack.csi.values
  }

  pack {
    name   = "cni-calico"
    tag    = "3.24.x"
    uid    = data.spectrocloud_pack.cni.id
    values = data.spectrocloud_pack.cni.values
  }

  pack {
    name   = "kubernetes"
    tag    = "1.26.x"
    uid    = data.spectrocloud_pack.k8s.id
    values = data.spectrocloud_pack.k8s.values
  }

  pack {
    name   = "ubuntu-aws"
    tag    = "LTS__22.4.x"
    uid    = data.spectrocloud_pack.ubuntu.id
    values = data.spectrocloud_pack.ubuntu.values
  }
}

data "spectrocloud_registry" "registry" {
  name = "Public Repo"
}

data "spectrocloud_pack" "csi" {
  name         = "csi-vsphere-csi"
  registry_uid = data.spectrocloud_registry.registry.id
  version      = "3.0.2"
}

data "spectrocloud_pack" "cni" {
  name         = "cni-cilium-oss"
  registry_uid = data.spectrocloud_registry.registry.id
  version      = "1.14.1"
}

data "spectrocloud_pack" "k8s" {
  name         = "kubernetes"
  registry_uid = data.spectrocloud_registry.registry.id
  version      = "1.27.5"
}

data "spectrocloud_pack" "ubuntu" {
  name         = "ubuntu-vsphere"
  registry_uid = data.spectrocloud_registry.registry.id
  version      = "22.04"
}

resource "spectrocloud_cluster_profile" "profile" {
  name        = "vsphere-profile-tf"
  description = "basic cp"
  cloud       = "vsphere"
  type        = "cluster"

  pack {
    name   = "ubuntu-vsphere"
    tag    = data.spectrocloud_pack.ubuntu.version
    uid    = data.spectrocloud_pack.ubuntu.id
    values = data.spectrocloud_pack.ubuntu.values
  }

  pack {
    name   = "kubernetes"
    tag    = data.spectrocloud_pack.k8s.version
    uid    = data.spectrocloud_pack.k8s.id
    values = data.spectrocloud_pack.k8s.values
  }

  pack {
    name   = "cni-cilium-oss"
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

  /*  pack {
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
  }*/


}

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


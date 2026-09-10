####################################
# Data resources for the profile
####################################
data "spectrocloud_registry" "public_registry" {
  name = "Public Repo"
}

data "spectrocloud_registry" "community_registry" {
  name = "Palette Registry"
}

####################################
# Core Infrastructure Layers
# The following core infrastructure layers are configured for deployment to vSphere.
####################################
data "spectrocloud_pack" "ubuntu" {
  name         = "ubuntu-vsphere"
  version      = "22.04"
  registry_uid = data.spectrocloud_registry.public_registry.id
}

data "spectrocloud_pack" "k8s" {
  name         = "kubernetes"
  version      = "1.32.3"
  registry_uid = data.spectrocloud_registry.public_registry.id
}

data "spectrocloud_pack" "cni" {
  name         = "cni-calico"
  version      = "3.29.3"
  registry_uid = data.spectrocloud_registry.public_registry.id
}

data "spectrocloud_pack" "csi" {
  name         = "csi-vsphere-csi"
  version      = "3.3.1"
  registry_uid = data.spectrocloud_registry.public_registry.id
}

data "spectrocloud_pack" "metallb" {
  name         = "lb-metallb-helm"
  version      = "0.14.9"
  registry_uid = data.spectrocloud_registry.public_registry.id
}

####################################
# Add-On Layers
####################################

data "spectrocloud_pack" "hellouniverse" {
  name         = "hello-universe"
  version      = "1.2.0"
  registry_uid = data.spectrocloud_registry.community_registry.id
}

####################################
# Data resources for the cluster
####################################
data "spectrocloud_cloudaccount_vsphere" "account" {
  name = var.pcg_name
}

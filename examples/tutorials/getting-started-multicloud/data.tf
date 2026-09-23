########################################
# Data resources for the cluster profile
########################################
data "spectrocloud_registry" "public_registry" {
  name = "Public Repo"
}

data "spectrocloud_registry" "community_registry" {
  name = "Palette Community Registry"
}

#############
# AWS
#############

data "spectrocloud_cloudaccount_aws" "account" {
  count = var.deploy-aws ? 1 : 0
  name  = var.aws-cloud-account-name
}

data "spectrocloud_pack" "aws_csi" {
  name         = "csi-aws-ebs"
  version      = "1.53.0"
  registry_uid = data.spectrocloud_registry.public_registry.id
}

data "spectrocloud_pack" "aws_cni" {
  name         = "cni-calico"
  version      = "3.30.3"
  registry_uid = data.spectrocloud_registry.public_registry.id
}

data "spectrocloud_pack" "aws_k8s" {
  name         = "kubernetes"
  version      = "1.32.10"
  registry_uid = data.spectrocloud_registry.public_registry.id
}

data "spectrocloud_pack" "aws_ubuntu" {
  name         = "ubuntu-aws"
  version      = "22.04"
  registry_uid = data.spectrocloud_registry.public_registry.id
}

#############
# Azure
#############
data "spectrocloud_cloudaccount_azure" "account" {
  count = var.deploy-azure ? 1 : 0
  name  = var.azure-cloud-account-name
}

data "spectrocloud_pack" "azure_csi" {
  name         = "csi-azure"
  version      = "1.32.0"
  registry_uid = data.spectrocloud_registry.public_registry.id
}

data "spectrocloud_pack" "azure_cni" {
  name         = "cni-calico-azure"
  version      = "3.29.3"
  registry_uid = data.spectrocloud_registry.public_registry.id
}

data "spectrocloud_pack" "azure_k8s" {
  name         = "kubernetes"
  version      = "1.32.3"
  registry_uid = data.spectrocloud_registry.public_registry.id
}

data "spectrocloud_pack" "azure_ubuntu" {
  name         = "ubuntu-azure"
  version      = "22.04"
  registry_uid = data.spectrocloud_registry.public_registry.id
}
#############
# GCP
#############
data "spectrocloud_cloudaccount_gcp" "account" {
  count = var.deploy-gcp ? 1 : 0
  name  = var.gcp-cloud-account-name
}

data "spectrocloud_pack" "gcp_csi" {
  name         = "csi-gcp-driver"
  version      = "1.15.4"
  registry_uid = data.spectrocloud_registry.public_registry.id
}

data "spectrocloud_pack" "gcp_cni" {
  name         = "cni-calico"
  version      = "3.29.3"
  registry_uid = data.spectrocloud_registry.public_registry.id
}

data "spectrocloud_pack" "gcp_k8s" {
  name         = "kubernetes"
  version      = "1.32.3"
  registry_uid = data.spectrocloud_registry.public_registry.id
}

data "spectrocloud_pack" "gcp_ubuntu" {
  name         = "ubuntu-gcp"
  version      = "22.04"
  registry_uid = data.spectrocloud_registry.public_registry.id
}

#############
# VMware
#############

data "spectrocloud_cloudaccount_vsphere" "account" {
  count = var.deploy-vmware ? 1 : 0
  name  = var.pcg_name
}

data "spectrocloud_pack" "vmware_ubuntu" {
  name         = "ubuntu-vsphere"
  version      = "22.04"
  registry_uid = data.spectrocloud_registry.public_registry.id
}

data "spectrocloud_pack" "vmware_k8s" {
  name         = "kubernetes"
  version      = "1.32.3"
  registry_uid = data.spectrocloud_registry.public_registry.id
}

data "spectrocloud_pack" "vmware_cni" {
  name         = "cni-calico"
  version      = "3.29.3"
  registry_uid = data.spectrocloud_registry.public_registry.id
}

data "spectrocloud_pack" "vmware_csi" {
  name         = "csi-vsphere-csi"
  version      = "3.3.1"
  registry_uid = data.spectrocloud_registry.public_registry.id
}

data "spectrocloud_pack" "vmware_metallb" {
  name         = "lb-metallb-helm"
  version      = "0.14.9"
  registry_uid = data.spectrocloud_registry.public_registry.id
}

# Required for static IP placement
data "spectrocloud_private_cloud_gateway" "pcg" {
  count = var.deploy-vmware-static ? 1 : 0
  name  = var.pcg_name
}

#####################
# Hello Universe Pack
#####################

data "spectrocloud_pack" "hellouniverse" {
  name         = "hello-universe"
  version      = "1.2.0"
  registry_uid = data.spectrocloud_registry.community_registry.id
}

#####################
# Kubecost Pack
#####################

data "spectrocloud_pack" "kubecost" {
  name         = "cost-analyzer"
  version      = "1.103.3"
  registry_uid = data.spectrocloud_registry.community_registry.id
}

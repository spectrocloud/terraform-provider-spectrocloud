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

data "spectrocloud_pack" "aws_ubuntu" {
  name         = "ubuntu-aws"
  version      = "22.04"
  registry_uid = data.spectrocloud_registry.public_registry.id
}

data "spectrocloud_pack" "aws_k8s" {
  name         = "kubernetes"
  version      = "1.35.2"
  registry_uid = data.spectrocloud_registry.public_registry.id
}

data "spectrocloud_pack" "aws_cni" {
  name         = "cni-calico"
  version      = "3.31.4"
  registry_uid = data.spectrocloud_registry.public_registry.id
}

data "spectrocloud_pack" "aws_csi" {
  name         = "csi-aws-ebs"
  version      = "1.59.0"
  registry_uid = data.spectrocloud_registry.public_registry.id
}

#############
# Azure
#############

data "spectrocloud_cloudaccount_azure" "account" {
  count = var.deploy-azure ? 1 : 0
  name  = var.azure-cloud-account-name
}

data "spectrocloud_pack" "azure_ubuntu" {
  name         = "ubuntu-azure"
  version      = "22.04"
  registry_uid = data.spectrocloud_registry.public_registry.id
}

data "spectrocloud_pack" "azure_k8s" {
  name         = "kubernetes"
  version      = "1.35.2"
  registry_uid = data.spectrocloud_registry.public_registry.id
}

data "spectrocloud_pack" "azure_cni" {
  name         = "cni-calico-azure"
  version      = "3.31.4"
  registry_uid = data.spectrocloud_registry.public_registry.id
}

data "spectrocloud_pack" "azure_csi" {
  name         = "csi-azure"
  version      = "1.34.2"
  registry_uid = data.spectrocloud_registry.public_registry.id
}

#####################
# Hello Universe Pack
#####################

data "spectrocloud_pack" "hellouniverse" {
  name         = "hello-universe"
  version      = "1.3.0"
  registry_uid = data.spectrocloud_registry.community_registry.id
}

###########
# Kubecost
###########

data "spectrocloud_pack" "kubecost" {
  count        = var.create_new_profile_version ? 1 : 0
  name         = "cost-analyzer"
  version      = "1.103.3"
  registry_uid = data.spectrocloud_registry.community_registry.id
}

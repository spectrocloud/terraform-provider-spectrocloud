####################################
# Data resources for the profile
####################################
data "spectrocloud_registry" "public_registry" {
  name = "Public Repo"
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
  version      = "1.41.0"
  registry_uid = data.spectrocloud_registry.public_registry.id
}

data "spectrocloud_pack" "aws_cni" {
  name         = "cni-calico"
  version      = "3.29.3"
  registry_uid = data.spectrocloud_registry.public_registry.id
}

data "spectrocloud_pack" "aws_k8s" {
  name         = "kubernetes"
  version      = "1.32.3"
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

data "spectrocloud_cluster" "azure_cluster_api" {
  count = var.deploy-azure ? 1 : 0

  name    = "azure-cluster-api"
  context = "project"

  depends_on = [spectrocloud_cluster_azure.azure-cluster-api]
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

data "spectrocloud_cluster" "gcp_cluster_api" {
  count = var.deploy-gcp ? 1 : 0

  name    = "gcp-cluster-api"
  context = "project"

  depends_on = [spectrocloud_cluster_gcp.gcp-cluster-api]
}

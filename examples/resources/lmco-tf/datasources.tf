# Note: These entities should be pre-created in Palette

# The Cloud account to lookup from Palette
data "spectrocloud_cloudaccount_aws" "demo_cloudaccount" {
  name = "aws-gov-dev"
}

# The Infra Cluster profile to lookup from Palette
data "spectrocloud_cluster_profile" "infra_demo" {
  name = "aws-pxke-infra"
  version = "1.27.2"
}

# The addon profile to lookup from Palette
data "spectrocloud_pack" "argo-cd" {
  name = var.argocd_name
  version  = var.argocd_version
  # registry_uid = data.spectrocloud_registry.private-demo.id
}

#data "spectrocloud_registry" "private-demo" {
#  name = "Public Repo"
#}
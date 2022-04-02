
data "spectrocloud_cluster_profile" "profile" {
  for_each = toset(values(var.SpectroCloudClusterProfiles))
  name     = each.value
}

data "spectrocloud_cloudaccount_aws" "account" {
  # id = <uid>
  name = var.SpectroCloudAccount
}

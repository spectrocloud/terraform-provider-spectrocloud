data "spectrocloud_cloudaccount_coxedge" "account" {
  name = var.shared_coxedge_cloud_account_name
}

resource "spectrocloud_cloudaccount_coxedge" "account" {
  name            = "coxedge-account"
  organization_id = var.organization_id
  environment     = var.environment
  service         = var.service
  api_base_url    = var.api_base_url
  api_key         = var.api_key
}

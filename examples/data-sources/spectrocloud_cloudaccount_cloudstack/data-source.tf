terraform {
  required_providers {
    spectrocloud = {
      version = ">= 0.1"
      source  = "spectrocloud/spectrocloud"
    }
  }
}

# Data source to fetch a CloudStack cloud account by name
data "spectrocloud_cloudaccount_cloudstack" "cloudstack_account_by_name" {
  name = "cloudstack-prod-account"
}

# Data source to fetch a CloudStack cloud account by ID
data "spectrocloud_cloudaccount_cloudstack" "cloudstack_account_by_id" {
  id = "your-cloudstack-account-id"
}

# Data source with explicit context
data "spectrocloud_cloudaccount_cloudstack" "cloudstack_account_with_context" {
  name    = "cloudstack-prod-account"
  context = "project" # or "tenant"
}

# Output the account details
output "account_id" {
  value = data.spectrocloud_cloudaccount_cloudstack.cloudstack_account_by_name.id
}

output "account_name" {
  value = data.spectrocloud_cloudaccount_cloudstack.cloudstack_account_by_name.name
}


terraform {
  required_providers {
    spectrocloud = {
      version = ">= 0.1"
      source  = "spectrocloud/spectrocloud"
    }
  }
}

# CloudStack Cloud Account Resource
resource "spectrocloud_cloudaccount_cloudstack" "cloudstack_account" {
  name = "cloudstack-prod-account"

  # Context - either 'project' or 'tenant'
  context = "project"

  # Private Cloud Gateway (PCG) ID
  private_cloud_gateway_id = "your-pcg-id-here"

  # CloudStack API Configuration
  api_url    = "https://cloudstack.example.com:8080/client/api"
  api_key    = "your-api-key-here"
  secret_key = "your-secret-key-here"

  # SSL Configuration (Optional)
  # Set insecure to true to skip SSL certificate verification
  # Note: CloudStack must have valid SSL certificates from a trusted CA if insecure is false
  insecure = false
}

# Output the cloud account ID
output "cloudstack_account_id" {
  value = spectrocloud_cloudaccount_cloudstack.cloudstack_account.id
}


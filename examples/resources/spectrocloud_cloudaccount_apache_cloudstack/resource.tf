# Nothing on this resource is ForceNew - every attribute below updates in place.

data "spectrocloud_private_cloud_gateway" "pcg" {
  name = "System Private Gateway"
}

resource "spectrocloud_cloudaccount_apache_cloudstack" "cloudstack_account" {
  # Required.
  name = "ran-tf-cloudstack-account"
  # Optional, default "project". Allowed: "project", "tenant".
  context = "project"
  # Required. The ID of the Private Cloud Gateway used to reach this CloudStack environment.
  private_cloud_gateway_id = data.spectrocloud_private_cloud_gateway.pcg.id

  # Required. The CloudStack management server's API endpoint.
  api_url = var.cloudstack_api_url # gitleaks:allow
  # Required, sensitive. CloudStack API key.
  api_key = var.cloudstack_api_key # gitleaks:allow
  # Required, sensitive. CloudStack secret key.
  secret_key = var.cloudstack_secret_key # gitleaks:allow

  # Optional, default "" (the ROOT domain). Set for multi-domain CloudStack environments.
  domain = "ROOT"
  # Optional, default false. Skips SSL certificate verification - only use this for
  # development/testing; CloudStack must have valid CA-signed certificates otherwise.
  insecure = true
}

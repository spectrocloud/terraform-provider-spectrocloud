# Creates and manages a dedicated Azure cloud account for this end-to-end example, rather than
# looking up one that already exists - so this folder is fully self-contained.
#
# Day-2 mutability: nothing on this resource is ForceNew - every attribute updates in place.
resource "spectrocloud_cloudaccount_azure" "account" {
  # Required.
  name = "e2e-azure-account"
  # Optional, default "project". Allowed: "project", "tenant".
  context = "project"
  # Optional. Required only for a private cluster endpoint reachable via a Private Cloud
  # Gateway; leave unset for a standard public Azure account.
  # private_cloud_gateway_id = var.private_cloud_gateway_id

  # Required. Azure AD application credentials.
  azure_tenant_id     = var.azure_tenant_id
  azure_client_id     = var.azure_client_id
  azure_client_secret = var.azure_client_secret

  # Optional. Display name of the Azure tenant.
  tenant_name = "e2e-demo-tenant"
  # Optional, default false. Whether to disable the properties request Palette normally makes
  # against this account.
  disable_properties_request = false

  # Optional, default "AzurePublicCloud". Allowed: "AzurePublicCloud",
  # "AzureUSGovernmentCloud", "AzureUSSecretCloud".
  cloud = "AzurePublicCloud"

  # Optional. Only allowed (and required) when cloud = "AzureUSSecretCloud".
  # tls_cert = file("azure-secret-cloud-cert.pem")
}

# Alternative to creating a new account: look up one that's already registered in Palette.
# data "spectrocloud_cloudaccount_azure" "account" {
#   name = "some-existing-account-name"
# }

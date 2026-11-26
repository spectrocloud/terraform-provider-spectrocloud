# Example 1: Basic Azure cloud account for public cloud
resource "spectrocloud_cloudaccount_azure" "azure_public" {
  name                = "azure-public-account"
  azure_tenant_id     = var.azure_tenant_id
  azure_client_id     = var.azure_client_id
  azure_client_secret = var.azure_client_secret

  # Optional: Context (defaults to "project")
  context = "project"

  # Optional: Cloud environment (defaults to "AzurePublicCloud")
  cloud = "AzurePublicCloud"

  # Optional: Tenant name
  tenant_name = "My Azure Tenant"

  # Optional: Disable properties request (defaults to false)
  disable_properties_request = false

  # Optional: Private cloud gateway ID for private cluster connectivity
  # private_cloud_gateway_id = "pcg-12345"
}

# Example 2: Azure US Government Cloud account
resource "spectrocloud_cloudaccount_azure" "azure_government" {
  name                = "azure-government-account"
  azure_tenant_id     = var.azure_gov_tenant_id
  azure_client_id     = var.azure_gov_client_id
  azure_client_secret = var.azure_gov_client_secret

  cloud   = "AzureUSGovernmentCloud"
  context = "project"
}

# Example 3: Azure US Secret Cloud account with TLS certificate
resource "spectrocloud_cloudaccount_azure" "azure_secret" {
  name                = "azure-secret-account"
  azure_tenant_id     = var.azure_secret_tenant_id
  azure_client_id     = var.azure_secret_client_id
  azure_client_secret = var.azure_secret_client_secret

  cloud   = "AzureUSSecretCloud"
  context = "project"

  # TLS certificate is only allowed when cloud is set to "AzureUSSecretCloud"
  tls_cert = var.azure_secret_tls_cert

  tenant_name = "Secret Cloud Tenant"
}

# Example 4: Azure account that uses a user-assigned managed identity attached to a private cloud gateway
resource "spectrocloud_cloudaccount_azure" "azure_msi" {
  name                     = "azure-msi-account"
  identity_type            = "UserAssignedMSI"
  azure_tenant_id          = var.azure_tenant_id
  azure_client_id          = var.azure_msi_client_id # client ID of the managed identity
  private_cloud_gateway_id = var.private_cloud_gateway_id

  # Do not set azure_client_secret: Palette rejects a managed identity account that has one.
  # A managed identity can be used by cloud accounts in only one project of the tenant.
}

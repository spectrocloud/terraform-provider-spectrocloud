resource "spectrocloud_cloudaccount_azure" "azure-1" {
  name                = "azure-1"
  azure_tenant_id     = var.azure_tenant_id
  azure_client_id     = var.azure_client_id
  azure_client_secret = var.azure_client_secret
}

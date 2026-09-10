# Spectro Cloud credentials
sc_host         = "{Enter Spectro Cloud API Host}" #e.g: api.spectrocloud.com (for SaaS)
sc_api_key      = "{Enter Spectro Cloud API Key}"
sc_project_name = "{Enter Spectro Cloud Project Name}" #e.g: Default

# Azure cloud account (created by this example, not looked up)
azure_tenant_id     = "{enter Azure AD tenant ID}"
azure_client_id     = "{enter Azure AD application client ID}"
azure_client_secret = "{enter Azure AD application client secret}"

# Cluster placement
azure_subscription_id  = "{enter Azure subscription ID}"
azure_resource_group   = "{enter Azure resource group}"
azure_region           = "eastus"
cluster_ssh_public_key = <<-EOT
  {enter SSH Public Key}
EOT

# Backup storage location (Azure Blob), created by this example
backup_azure_storage_name   = "{enter Azure storage account name}"
backup_azure_container_name = "backups"
backup_azure_resource_group = "{enter resource group of the storage account}"

# Where to write the local kubeconfig files (defaults to the current directory)
# kubeconfig_output_dir = "."

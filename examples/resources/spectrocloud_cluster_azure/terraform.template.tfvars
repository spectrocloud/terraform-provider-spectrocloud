# Spectro Cloud credentials
sc_host         = "{enter Spectro Cloud Host}" #e.g: api.spectrocloud.com (for SaaS)
sc_api_key      = "{enter Spectro Cloud API endpoint}"
sc_project_name = "{enter Spectro Cloud project Name}" #e.g: Default

# Azure Cloud Account credentials
# Follow the Spectro Cloud documentation to provision a new Azure Enterprise Application.
# https://docs.spectrocloud.com/clusters/?clusterType=azure_cluster#creatinganazurecloudaccount
azure_tenant_id     = "{enter Azure Tenant Id}"
azure_client_id     = "{enter Azure Client Id}"
azure_client_secret = "{enter Azure Client Secret}"

# SSH public key to inject into all K8s nodes
# Insert your public key between the EOT markers
# The public key starts with "ssh-rsa ...."
cluster_ssh_public_key = <<-EOT
  {enter SSH Public Key}
EOT


# Cluster Placement properties
# https://azure.microsoft.com/en-us/global-infrastructure/geographies/#geographies
# The region names are lowercase with spaces removed, e.g: "West US" -> westus
azure_subscription_id = "{enter Azure Subscription Id}"
azure_resource_group  = "{enter Azure Resource Group}"
azure_region          = "{enter Azure Region}" #e.g: westus

# Credentials
sc_host         = "api.spectrocloud.com"
sc_username     = "<...>"
sc_password     = "<...>"
sc_project_name = "Default"

# Cloud Account credentials
shared_vmware_cloud_account_name = "..."

# Cluster
cluster_ssh_public_key = <<-EOT
  ssh-rsa AAA...
EOT

# For DHCP, the search domain
cluster_network_search = "..." #e.g spectrocloud.local

vsphere_datacenter = "..."
vsphere_folder = "..."

vsphere_cluster = "..."
vsphere_resource_pool = ""
vsphere_datastore = "..."
vsphere_network = "..."

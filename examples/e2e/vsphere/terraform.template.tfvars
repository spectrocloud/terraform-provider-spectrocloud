# Spectro Cloud credentials
sc_host         = "{enter Spectro Cloud API endpoint}" #e.g: api.spectrocloud.com (for SaaS)
sc_username     = "{enter Spectro Cloud username}"     #e.g: user1@abc.com
sc_password     = "{enter Spectro Cloud password}"     #e.g: supereSecure1!
sc_project_name = "{enter Spectro Cloud project Name}" #e.g: Default

# Cloud Account credentials
shared_vmware_cloud_account_name = "..."

# Cluster
cluster_ssh_public_key = <<-EOT
  ssh-rsa AAA...
EOT

# For DHCP, the search domain
cluster_network_search = "..." #e.g spectrocloud.local

vsphere_datacenter = "..."
vsphere_folder     = "..."

vsphere_cluster       = "..."
vsphere_resource_pool = ""
vsphere_datastore     = "..."
vsphere_network       = "..."

# Spectro Cloud credentials
sc_host         = "{enter Spectro Cloud API endpoint}" #e.g: api.spectrocloud.com (for SaaS)
sc_username     = "{enter Spectro Cloud username}"     #e.g: user1@abc.com
sc_password     = "{enter Spectro Cloud password}"     #e.g: supereSecure1!
sc_project_name = "{enter Spectro Cloud project Name}" #e.g: Default

# Cloud Account lookup by name
# See README.md for instructions how to obtain this name
shared_vmware_cloud_account_name = "{enter Spectro Cloud VMware Cloud Account name}"

# SSH public key to inject into all K8s nodes
# Insert your public key between the EOT markers
# The public key starts with "ssh-rsa ...."
cluster_ssh_public_key = <<-EOT
  {enter SSH Public Key}
EOT

# For DHCP, the search domain
cluster_network_search = "{enter DHCP Search domain}" #e.g spectrocloud.local

# VMware cluster placement properties
# All fields except _vsphere\_resource\_pool_ are required fields
vsphere_datacenter = "{enter vSphere Datacenter}"
vsphere_folder     = "{enter vSphere Folder}"

vsphere_cluster       = "{enter vSphere ESX Cluster}"
vsphere_resource_pool = "{enter vSphere Resource Pool}" # Leave "" blank for Cluster Resource pool
vsphere_datastore     = "{enter vSphere Datastore}"
vsphere_network       = "{enter vSphere Network}"

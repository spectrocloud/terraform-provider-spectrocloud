# Spectro Cloud credentials
sc_host         = "{Enter Spectro Cloud API Host}" #e.g: api.spectrocloud.com (for SaaS)
sc_api_key      = "{Enter Spectro Cloud API Key}"
sc_project_name = "{Enter Spectro Cloud Project Name}" #e.g: Default

# UID of the registered Edge host (spectrocloud_appliance) this cluster is deployed on
edge_host_uid = "{enter Edge host UID}"

# SSH public key to inject into all K8s nodes
# Insert your public key between the EOT markers
# The public key starts with "ssh-rsa ...."
cluster_ssh_public_key = <<-EOT
  {enter SSH Public Key}
EOT

# Virtual IP for the Kubernetes control plane endpoint
cluster_vip = "{enter control plane VIP}"

# VMware cluster placement properties
# All fields except _vsphere\_resource\_pool_ are required fields
vsphere_datacenter = "{enter vSphere Datacenter}"
vsphere_folder     = "{enter vSphere Folder}"

vsphere_cluster       = "{enter vSphere ESX Cluster}"
vsphere_resource_pool = "{enter vSphere Resource Pool}" # Leave "" blank for Cluster Resource pool
vsphere_datastore     = "{enter vSphere Datastore}"
vsphere_network       = "{enter vSphere Network}"

# Backup storage location lookup by name (only needed if you keep the backup_policy block)
backup_storage_location_name = "{enter Backup Storage Location name}"

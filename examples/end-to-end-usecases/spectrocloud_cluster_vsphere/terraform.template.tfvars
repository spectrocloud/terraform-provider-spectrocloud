# Spectro Cloud credentials
sc_host         = "{Enter Spectro Cloud API Host}" #e.g: api.spectrocloud.com (for SaaS)
sc_api_key      = "{Enter Spectro Cloud API Key}"
sc_project_name = "{Enter Spectro Cloud Project Name}" #e.g: Default

# vSphere cloud account (created by this example, not looked up)
private_cloud_gateway_id      = "{enter Private Cloud Gateway UID}"
vsphere_vcenter               = "{enter vCenter address, e.g. vcenter.corp.example.com}"
vsphere_username              = "{enter vCenter username}"
vsphere_password              = "{enter vCenter password}"
vsphere_ignore_insecure_error = false

# SSH public key to inject into all K8s nodes
cluster_ssh_public_key = <<-EOT
  {enter SSH Public Key}
EOT

# VMware cluster placement properties
vsphere_datacenter = "{enter vSphere Datacenter}"
vsphere_folder     = "{enter vSphere Folder}"

vsphere_cluster       = "{enter vSphere ESX Cluster}"
vsphere_resource_pool = "{enter vSphere Resource Pool}" # Leave "" blank for Cluster Resource pool
vsphere_datastore     = "{enter vSphere Datastore}"
vsphere_network       = "{enter vSphere Network}"

# Backup storage location (S3), created by this example
backup_s3_region      = "{enter AWS region, e.g. us-east-1}"
backup_s3_bucket_name = "{enter S3 bucket name}"
backup_s3_access_key  = "{enter AWS access key}"
backup_s3_secret_key  = "{enter AWS secret key}"

# Where to write the local kubeconfig files (defaults to the current directory)
# kubeconfig_output_dir = "."

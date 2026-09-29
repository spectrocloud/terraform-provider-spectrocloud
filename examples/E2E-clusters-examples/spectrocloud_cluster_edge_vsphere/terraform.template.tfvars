# Spectro Cloud credentials
sc_host         = "{Enter Spectro Cloud API Host}" #e.g: api.spectrocloud.com (for SaaS)
sc_api_key      = "{Enter Spectro Cloud API Key}"
sc_project_name = "{Enter Spectro Cloud Project Name}" #e.g: Default

# Edge appliance (paired in Palette under Clusters > Edge Hosts), registered by this example
edge_host_uid = "{Enter edge appliance UID}"

# Cluster placement
cluster_ssh_public_key = "ssh-rsa AAAA..."
vsphere_datacenter     = "{Enter vSphere Datacenter}"
vsphere_folder         = "{Enter vSphere Folder}"
vsphere_cluster        = "{Enter vSphere Cluster}"
vsphere_resource_pool  = "{Enter vSphere Resource Pool}"
vsphere_datastore      = "{Enter vSphere Datastore}"
vsphere_network        = "{Enter vSphere Network}"
cluster_vip            = "{Enter control plane VIP}"

# Backup storage location (Minio, S3-compatible), created by this example
backup_minio_endpoint   = "https://minio.mycompany.com"
backup_minio_access_key = "{Enter Minio Access Key}"
backup_minio_secret_key = "{Enter Minio Secret Key}"

# Where to write the local kubeconfig files (defaults to the current directory)
# kubeconfig_output_dir = "."

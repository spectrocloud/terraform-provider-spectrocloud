# Spectro Cloud credentials
sc_host         = "{Enter Spectro Cloud API Host}" #e.g: api.spectrocloud.com (for SaaS)
sc_api_key      = "{Enter Spectro Cloud API Key}"
sc_project_name = "{Enter Spectro Cloud Project Name}" #e.g: Default

# Edge appliances (paired in Palette under Clusters > Edge Hosts), registered by this example
control_plane_edge_host_uid = "{Enter control-plane edge appliance UID}"
worker_edge_host_uid        = "{Enter worker edge appliance UID}"

# Cluster placement
cluster_ssh_public_keys = ["ssh-rsa AAAA...", ]

# Backup storage location (Minio, S3-compatible), created by this example
backup_minio_endpoint   = "https://minio.mycompany.com"
backup_minio_access_key = "{Enter Minio Access Key}"
backup_minio_secret_key = "{Enter Minio Secret Key}"

# Where to write the local kubeconfig files (defaults to the current directory)
# kubeconfig_output_dir = "."

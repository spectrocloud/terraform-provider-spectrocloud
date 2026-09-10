# Spectro Cloud credentials
sc_host         = "{Enter Spectro Cloud API Host}" #e.g: api.spectrocloud.com (for SaaS)
sc_api_key      = "{Enter Spectro Cloud API Key}"
sc_project_name = "{Enter Spectro Cloud Project Name}" #e.g: Default

# Private Cloud Gateway (must already be installed and reachable from your MAAS environment)
private_cloud_gateway_id = "{Enter Private Cloud Gateway UID}"

# MAAS cloud account (created by this example, not looked up)
maas_api_endpoint = "http://maas.mycompany.com:5240/MAAS"
maas_api_key      = "{Enter MAAS API Key}"

# Cluster placement
maas_domain             = "maas.mycompany.com"
maas_resource_pool      = "Medium-Generic"
cluster_ssh_public_keys = ["ssh-rsa AAAA...", ]

# Backup storage location (Minio, S3-compatible), created by this example
backup_minio_endpoint   = "https://minio.mycompany.com"
backup_minio_access_key = "{Enter Minio Access Key}"
backup_minio_secret_key = "{Enter Minio Secret Key}"

# Where to write the local kubeconfig files (defaults to the current directory)
# kubeconfig_output_dir = "."

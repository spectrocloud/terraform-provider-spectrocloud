# Spectro Cloud credentials
sc_host         = "{Enter Spectro Cloud API Host}" #e.g: api.spectrocloud.com (for SaaS)
sc_api_key      = "{Enter Spectro Cloud API Key}"
sc_project_name = "{Enter Spectro Cloud Project Name}" #e.g: Default

# Private Cloud Gateway (must already be installed and reachable from your CloudStack environment)
private_cloud_gateway_id = "{Enter Private Cloud Gateway UID}"

# CloudStack cloud account (created by this example, not looked up)
cloudstack_api_url    = "{Enter CloudStack API URL}" #e.g: https://cloudstack.mycompany.com/client/api
cloudstack_api_key    = "{Enter CloudStack API Key}"
cloudstack_secret_key = "{Enter CloudStack Secret Key}"

# Cluster placement
cloudstack_zone_name       = "{Enter CloudStack Zone Name}"
cloudstack_network_name    = "{Enter CloudStack Network Name}"
cloudstack_ssh_key_name    = "{Enter CloudStack SSH Key Name}"
cloudstack_cp_offering     = "{Enter CloudStack Compute Offering for control plane}"
cloudstack_worker_offering = "{Enter CloudStack Compute Offering for workers}"

# Backup storage location (Minio, S3-compatible), created by this example
backup_minio_endpoint   = "https://minio.mycompany.com"
backup_minio_access_key = "{Enter Minio Access Key}"
backup_minio_secret_key = "{Enter Minio Secret Key}"

# Where to write the local kubeconfig files (defaults to the current directory)
# kubeconfig_output_dir = "."

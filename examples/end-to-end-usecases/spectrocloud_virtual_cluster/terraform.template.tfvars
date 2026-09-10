# Spectro Cloud credentials
sc_host         = "{Enter Spectro Cloud API Host}" #e.g: api.spectrocloud.com (for SaaS)
sc_api_key      = "{Enter Spectro Cloud API Key}"
sc_project_name = "{Enter Spectro Cloud Project Name}" #e.g: Default

# Existing Palette-managed cluster this virtual cluster is created on
host_cluster_uid = "{Enter host cluster UID}"

# Backup storage location (Minio, S3-compatible), created by this example
backup_minio_endpoint   = "https://minio.mycompany.com"
backup_minio_access_key = "{Enter Minio Access Key}"
backup_minio_secret_key = "{Enter Minio Secret Key}"

# Where to write the local kubeconfig files (defaults to the current directory)
# kubeconfig_output_dir = "."

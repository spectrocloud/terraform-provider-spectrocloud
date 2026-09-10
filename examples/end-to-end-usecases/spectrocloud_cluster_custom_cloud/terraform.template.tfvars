# Spectro Cloud credentials
sc_host         = "{Enter Spectro Cloud API Host}" #e.g: api.spectrocloud.com (for SaaS)
sc_api_key      = "{Enter Spectro Cloud API Key}"
sc_project_name = "{Enter Spectro Cloud Project Name}" #e.g: Default

# Private Cloud Gateway (must already be installed and reachable from your Nutanix environment)
private_cloud_gateway_id = "{Enter Private Cloud Gateway UID}"

# Nutanix cloud account (created by this example, not looked up)
nutanix_user     = "{Enter Nutanix Prism Central username}"
nutanix_password = "{Enter Nutanix Prism Central password}"
nutanix_endpoint = "{Enter Nutanix Prism Central endpoint}"
nutanix_port     = "9440"

# Cluster placement
control_plane_endpoint_ip           = "{Enter control plane VIP}"
nutanix_ssh_authorized_key          = "ssh-rsa AAAA..."
nutanix_prism_element_cluster_name  = "{Enter Nutanix Prism Element cluster name}"
nutanix_machine_template_image_name = "{Enter Nutanix VM image template name}"
nutanix_subnet_name                 = "{Enter Nutanix subnet name}"

# Backup storage location (Minio, S3-compatible), created by this example
backup_minio_endpoint   = "https://minio.mycompany.com"
backup_minio_access_key = "{Enter Minio Access Key}"
backup_minio_secret_key = "{Enter Minio Secret Key}"

# Where to write the local kubeconfig files (defaults to the current directory)
# kubeconfig_output_dir = "."

# Spectro Cloud credentials
sc_host         = "{Enter Spectro Cloud API Host}" #e.g: api.spectrocloud.com (for SaaS)
sc_api_key      = "{Enter Spectro Cloud API Key}"
sc_project_name = "{Enter Spectro Cloud Project Name}" #e.g: Default

# GCP cloud account (created by this example, not looked up)
gcp_json_credentials = <<-EOT
  {enter GCP service account JSON key}
EOT

# Cluster placement
gcp_project_id = "{enter GCP project ID}"
gcp_region     = "us-east1"
gcp_network    = "default"

# Backup storage location (GCS), created by this example
backup_gcp_project_id       = "{enter GCP project ID for backups}"
backup_gcp_json_credentials = <<-EOT
  {enter GCP service account JSON key for backups}
EOT

# Where to write the local kubeconfig files (defaults to the current directory)
# kubeconfig_output_dir = "."

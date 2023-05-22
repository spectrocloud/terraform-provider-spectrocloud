# Spectro Cloud credentials
sc_host         = "{Enter Spectro Cloud API Host}" #e.g: api.spectrocloud.com (for SaaS)
sc_api_key      = "{Enter Spectro Cloud API Key}"
sc_project_name = "{Enter Spectro Cloud Project Name}" #e.g: Default

cluster_cloud_account_name   = "gcp-1"
cluster_cluster_profile_name = "ProdGoogle"
gcp_serviceaccount_json = <<-EOT
  {
    "type": "service_account",
    "project_id": "gcp-project-1",
    ...
  }
EOT

# Spectro Cloud credentials
sc_host         = "{enter Spectro Cloud Host}" #e.g: api.spectrocloud.com (for SaaS)
sc_api_key      = "{enter Spectro Cloud API endpoint}"
sc_project_name = "{enter Spectro Cloud project Name}" #e.g: Default

gcp_serviceaccount_json = <<-EOT
  {
    "type": "service_account",
    "project_id": "gcp-project-1",
    ...
  }
EOT

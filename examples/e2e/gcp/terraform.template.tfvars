# Credentials
sc_host         = "api.spectrocloud.com"
sc_username     = "<...>"
sc_password     = "<...>"
sc_project_name = "Default"

# Cloud Account credentials
gcp_serviceaccount_json = <<-EOT
  {
    "type": "service_account",
    ....
  }
EOT

# Cluster
gcp_network = ""
gcp_project = "..."
gcp_region  = "us-west3"

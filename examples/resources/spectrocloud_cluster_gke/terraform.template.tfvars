# Spectro Cloud credentials
sc_host         = "{Enter Spectro Cloud API Host}" #e.g: api.spectrocloud.com (for SaaS)
sc_api_key      = "{Enter Spectro Cloud API Key}"
sc_project_name = "{Enter Spectro Cloud Project Name}" #e.g: Default

# Google Cloud account credentials
# Create a new GCP service account with the Editor role mapping
# https://cloud.google.com/iam/docs/creating-managing-service-account-keys
#
# Paste the service account JSON key contents inside the yaml heredoc EOT markers.
gcp_serviceaccount_json = <<-EOT
  {enter GCP service account json}
EOT

# GCP Cluster Placement properties
#
gcp_network = "{enter GCP network}" #e.g: "" (this one can be blank)
gcp_project = "{enter GCP project}"
gcp_region  = "{enter GCP region}" #e.g: us-west3

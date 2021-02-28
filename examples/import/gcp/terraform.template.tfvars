# Spectro Cloud credentials
sc_host         = "{enter Spectro Cloud API endpoint}" #e.g: api.spectrocloud.com (for SaaS)
sc_username     = "{enter Spectro Cloud username}"     #e.g: user1@abc.com
sc_password     = "{enter Spectro Cloud password}"     #e.g: supereSecure1!
sc_project_name = "{enter Spectro Cloud project Name}" #e.g: Default

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

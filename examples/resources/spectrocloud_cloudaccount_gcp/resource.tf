# Nothing on this resource is ForceNew - every attribute below updates in place.
#
# Attributes:
#   name                  - Required.
#   context               - Optional, default "project". Allowed: "project", "tenant".
#   gcp_json_credentials  - Required, sensitive. The GCP service account credentials, in JSON
#                           format.
resource "spectrocloud_cloudaccount_gcp" "gcp-1" {
  name                 = "gcp-1"
  context              = "project"
  gcp_json_credentials = var.gcp_serviceaccount_json
}

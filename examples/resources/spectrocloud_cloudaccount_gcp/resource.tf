# Nothing on this resource is ForceNew - both attributes update in place.
resource "spectrocloud_cloudaccount_gcp" "gcp-1" {
  # Required.
  name = "gcp-1"
  # Optional, default "project". Allowed: "project", "tenant".
  context = "project"
  # Required, sensitive. The GCP service account credentials, in JSON format.
  gcp_json_credentials = var.gcp_serviceaccount_json
}

# Creates and manages a dedicated GCP cloud account for this end-to-end example, rather than
# looking up one that already exists - so this folder is fully self-contained.
#
# Day-2 mutability: nothing on this resource is ForceNew - every attribute updates in place.
resource "spectrocloud_cloudaccount_gcp" "account" {
  # Required.
  name = "e2e-gcp-account"
  # Optional, default "project". Allowed: "project", "tenant".
  context = "project"
  # Required, sensitive (credential material). GCP service account key JSON.
  gcp_json_credentials = var.gcp_json_credentials
}

# Alternative to creating a new account: look up one that's already registered in Palette.
# data "spectrocloud_cloudaccount_gcp" "account" {
#   name = "some-existing-account-name"
# }

# Creates and manages a dedicated custom cloud account for this end-to-end example, rather than
# looking up one that already exists - so this folder is fully self-contained. spectrocloud_
# cloudaccount_custom is generic: `cloud` names the registered custom cloud provider (Nutanix
# here), and `credentials` holds whatever provider-specific key/value pairs that provider needs.
#
# Day-2 mutability: `cloud`, `private_cloud_gateway_id`, and `context` are ForceNew. `name` and
# `credentials` update in place.
#
# Prerequisite: a Private Cloud Gateway (PCG) must already be installed and reachable from this
# Nutanix Prism Central before this account can be validated.
resource "spectrocloud_cloudaccount_custom" "account" {
  # Required. Display name shown in the Palette UI.
  name = "e2e-custom-cloud-account"
  # Required, ForceNew. The custom cloud provider name.
  cloud = "nutanix"
  # Required, ForceNew. UID of the Private Cloud Gateway this account routes traffic through.
  private_cloud_gateway_id = var.private_cloud_gateway_id
  # Optional, default "project", ForceNew. Allowed: "project", "tenant".
  context = "project"
  # Optional, sensitive. Provider-specific credential key/value pairs - the required keys depend
  # on the custom cloud provider (Nutanix Prism Central credentials, shown here).
  credentials = {
    "NUTANIX_USER"     = var.nutanix_user
    "NUTANIX_PASSWORD" = var.nutanix_password
    "NUTANIX_ENDPOINT" = var.nutanix_endpoint
    "NUTANIX_PORT"     = var.nutanix_port
    "NUTANIX_INSECURE" = "no"
  }
}

# Alternative to creating a new account: look up one that's already registered in Palette.
# data "spectrocloud_cloudaccount_custom" "account" {
#   name  = "some-existing-account-name"
#   cloud = "nutanix"
# }

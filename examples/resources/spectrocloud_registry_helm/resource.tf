# Day-2 mutability: only `name` is ForceNew - changing it recreates the registry. `endpoint`,
# `is_synchronization`, `credentials` (including nested tls_config), and `wait_for_sync` all
# update in place.
resource "spectrocloud_registry_helm" "r1" {
  # Required, ForceNew. Must be unique across Helm registries.
  name = "us-artifactory"
  # Required. URL of the Helm chart repository.
  endpoint = "https://123456.dkr.ecr.us-west-1.amazonaws.com"

  # Optional/Computed, mutually exclusive with the deprecated `is_private` (set only one).
  # false = private/not synchronized by Palette (requires auth, as configured below); true =
  # public and synchronized by Palette.
  is_synchronization = false

  credentials {
    # Required. "noAuth" (no credentials), "basic" (username/password, used here), or "token".
    credential_type = "basic"
    # Required when credential_type = "basic".
    username = "abc"
    # Required when credential_type = "basic" (credential material).
    password = "def"
    # Required when credential_type = "token" instead of "basic" (credential material, omit
    # username/password in that case).
    # token = var.registry_token

    # Optional: TLS configuration for the connection to the registry.
    # tls_config {
    #   enabled     = true
    #   ca          = file("ca.pem")
    #   certificate = file("client-cert.pem")
    #   key         = file("client-key.pem")
    #   # insecure_skip_verify = true  # only for trusted networks with self-signed certs
    # }
  }

  # Optional, default false. If true, Terraform waits for the registry's initial
  # synchronization to complete before marking create/update as done.
  # wait_for_sync = true
}
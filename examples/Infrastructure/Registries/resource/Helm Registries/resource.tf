# Day-2 mutability: only `name` is ForceNew - changing it recreates the registry. `endpoint`,
# `is_synchronization`, `credentials` (including nested tls_config), and `wait_for_sync` all
# update in place.
#
# Attributes:
#   name               - Required, ForceNew. Must be unique across Helm registries.
#   endpoint            - Required. URL of the Helm chart repository.
#   is_synchronization - Optional/Computed, mutually exclusive with the deprecated `is_private`
#                        (set only one). false = private/not synchronized by Palette (requires
#                        auth, as configured below); true = public and synchronized by Palette.
#   wait_for_sync      - Optional, default false. If true, Terraform waits for the registry's
#                        initial synchronization to complete before marking create/update as done.
#
# credentials block (required, at most one):
#   credential_type - Required. "noAuth" (no credentials), "basic" (username/password, used
#                     here), or "token".
#   username         - Required when credential_type = "basic".
#   password         - Required when credential_type = "basic" (credential material).
#   token            - Required when credential_type = "token" instead of "basic" (credential
#                      material, omit username/password in that case).
#   tls_config       - Optional nested block: TLS configuration for the connection to the
#                      registry (enabled, ca, certificate, key, insecure_skip_verify - the last
#                      only for trusted networks with self-signed certs).
resource "spectrocloud_registry_helm" "r1" {
  name               = "us-artifactory"
  endpoint           = "https://123456.dkr.ecr.us-west-1.amazonaws.com"
  is_synchronization = false

  credentials {
    credential_type = "basic"
    username        = "abc"
    password        = "def"
    # token = var.registry_token

    # tls_config {
    #   enabled     = true
    #   ca          = file("ca.pem")
    #   certificate = file("client-cert.pem")
    #   key         = file("client-key.pem")
    #   # insecure_skip_verify = true  # only for trusted networks with self-signed certs
    # }
  }

  # wait_for_sync = true
}
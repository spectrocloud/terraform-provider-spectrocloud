# Day-2 mutability: `name` and `type` are ForceNew - changing either recreates the registry.
# `is_private`, `is_synchronization`, `endpoint`, `endpoint_suffix`, `base_content_path`,
# `provider_type`, `wait_for_sync`, and `credentials` all update in place.
#
# Attributes:
#   name               - Required, ForceNew.
#   type               - Required, ForceNew. "ecr" for Amazon ECR (this example) or "basic" for
#                        other OCI registries.
#   endpoint           - Required. URL of the registry endpoint.
#   is_private         - Required. true = private registry requiring the credentials block
#                        below; false = public.
#   is_synchronization - Optional, default false. If true, Palette synchronizes the registry's
#                        contents; requires `base_content_path` to also be set. Not used here.
#   base_content_path  - See is_synchronization above.
#   endpoint_suffix    - Optional. Suffix appended to `endpoint` - some registries (e.g. JFrog)
#                        require this.
#   provider_type      - Optional, default "helm". Allowed: "helm", "zarf" (only with
#                        type = "basic"), "pack" (only when is_synchronization = true).
#   wait_for_sync      - Optional, default false. Wait for initial sync before marking
#                        create/update complete. Applicable only when provider_type is "helm" or
#                        "zarf".
#
# credentials block (required, at most one):
#   credential_type - Required. "secret" (access_key/secret_key), "sts" (arn/external_id, used
#                     here for ECR), "basic" (username/password), or "noAuth".
#   arn             - Required when credential_type = "sts". IAM role ARN to assume for ECR
#                     access.
#   external_id     - Required when credential_type = "sts". External ID for the AWS STS
#                     AssumeRole call.
#   access_key/secret_key - Used when credential_type = "secret" (mutually exclusive with
#                     arn/external_id/username/password above).
#   username/password     - Used when credential_type = "basic" (mutually exclusive with the
#                     other credential_type fields).
#   tls_config      - Optional nested block: TLS configuration for the connection to the
#                     registry (certificate, insecure_skip_verify).
resource "spectrocloud_registry_oci" "r1" {
  name       = "test-nik2"
  type       = "ecr" # basic
  endpoint   = "123456.dkr.ecr.us-west-1.amazonaws.com"
  is_private = true

  # is_synchronization = true
  # base_content_path  = "/"

  # endpoint_suffix = "/v2"

  # provider_type = "helm"

  # wait_for_sync = true

  credentials {
    credential_type = "sts"
    arn             = "arn:aws:iam::123456:role/stage-demo-ecr"
    external_id     = "sofiwhgowbrgiornM="

    # access_key = "AKIA..."       # credential_type = "secret"
    # secret_key = "..."           # credential_type = "secret"
    # username   = "registry-user" # credential_type = "basic"
    # password   = "..."           # credential_type = "basic"

    # tls_config {
    #   certificate          = file("ca.pem")
    #   insecure_skip_verify = false
    # }
  }
}
# 
# Import by Name:
# terraform import spectrocloud_registry_oci.example "test-nik2"
# terraform import spectrocloud_registry_oci.example "REGISTRY-NAME"

# Import by UID:
# terraform import spectrocloud_registry_oci.example "IMPORT-UID"

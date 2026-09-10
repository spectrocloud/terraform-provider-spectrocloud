# Day-2 mutability: `name` and `type` are ForceNew - changing either recreates the registry.
# `is_private`, `is_synchronization`, `endpoint`, `endpoint_suffix`, `base_content_path`,
# `provider_type`, `wait_for_sync`, and `credentials` all update in place.
resource "spectrocloud_registry_oci" "r1" {
  # Required, ForceNew.
  name = "test-nik2"
  # Required, ForceNew. "ecr" for Amazon ECR (this example) or "basic" for other OCI registries.
  type = "ecr" # basic
  # Required. URL of the registry endpoint.
  endpoint = "123456.dkr.ecr.us-west-1.amazonaws.com"
  # Required. true = private registry requiring the credentials block below; false = public.
  is_private = true

  # Optional, default false. If true, Palette synchronizes the registry's contents; requires
  # `base_content_path` to also be set. Not used in this example.
  # is_synchronization = true
  # base_content_path  = "/"

  # Optional. Suffix appended to `endpoint` - some registries (e.g. JFrog) require this.
  # endpoint_suffix = "/v2"

  # Optional, default "helm". Allowed: "helm", "zarf" (only with type = "basic"), "pack" (only
  # when is_synchronization = true).
  # provider_type = "helm"

  # Optional, default false. Wait for initial sync before marking create/update complete.
  # Applicable only when provider_type is "helm" or "zarf".
  # wait_for_sync = true

  credentials {
    # Required. "secret" (access_key/secret_key), "sts" (arn/external_id, used here for ECR),
    # "basic" (username/password), or "noAuth".
    credential_type = "sts"
    # Required when credential_type = "sts". IAM role ARN to assume for ECR access.
    arn = "arn:aws:iam::123456:role/stage-demo-ecr"
    # Required when credential_type = "sts". External ID for the AWS STS AssumeRole call.
    external_id = "sofiwhgowbrgiornM="

    # Fields for the other credential_type values (mutually exclusive with arn/external_id
    # above - set only the ones matching your chosen credential_type):
    # access_key = "AKIA..."       # credential_type = "secret"
    # secret_key = "..."           # credential_type = "secret"
    # username   = "registry-user" # credential_type = "basic"
    # password   = "..."           # credential_type = "basic"

    # Optional: TLS configuration for the connection to the registry.
    # tls_config {
    #   certificate          = file("ca.pem")
    #   insecure_skip_verify = false
    # }
  }
}
# 
# Import by Name:
# terraform import spectrocloud_registry_oci.example "Pack Registry"
# terraform import spectrocloud_registry_oci.example "REGISTRY-NAME"

# Import by UID:
# terraform import spectrocloud_registry_oci.example "IMPORT-UID"

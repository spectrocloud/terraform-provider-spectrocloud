data "spectrocloud_team" "team" {
  name = "Tenant Admin"
}

# Manages tenant-wide SSO configuration (singleton per tenant), configured for OIDC. Day-2
# mutability: nothing here is ForceNew - every attribute, including the nested oidc block,
# updates in place.
#
# `sso_auth_type = "oidc"` requires the oidc block below and forbids a saml block.
#
# Flat attributes:
#   sso_auth_type  - Optional, default "none". Allowed: "none", "saml", "oidc".
#   domains        - Optional. Email domains that are routed through this SSO configuration.
#   auth_providers - Optional. External auth providers to also allow. Allowed values: "github",
#                    "google".
resource "spectrocloud_sso" "sso_setting" {
  sso_auth_type  = "oidc"
  domains        = ["test.com", "test-login.com"]
  auth_providers = ["github", "google"]

  # oidc block (Required when sso_auth_type = "oidc", at most one):
  #   insecure_skip_tls_verify - Only for trusted networks with self-signed certs.
  #   client_secret            - Credential material, sensitive.
  #
  #   user_info_endpoint block (Optional, at most one): query the OIDC userinfo endpoint for
  #   these claims when the ID token doesn't carry them.
  oidc {
    issuer_url                       = "https://login.microsoftonline.com/sd8/v2.0"
    identity_provider_ca_certificate = "test certificate content"
    insecure_skip_tls_verify         = false
    client_id                        = var.oidc_client_id
    client_secret                    = var.oidc_client_secret
    default_team_ids                 = [data.spectrocloud_team.team.id]
    scopes                           = ["profile", "email"]
    first_name                       = "test"
    last_name                        = "last"
    email                            = "test@test.com"
    spectro_team                     = "groups"

    user_info_endpoint {
      first_name   = "test"
      last_name    = "last"
      email        = "test@test.com"
      spectro_team = "groups"
    }
  }

  # SAML is mutually exclusive with oidc above - only one may be set, matching sso_auth_type.
  # saml {
  #   service_provider           = "Microsoft ADFS"
  #   identity_provider_metadata = "<note>test</note>"
  #   default_team_ids           = [data.spectrocloud_team.team.id]
  #   enable_single_logout       = true
  #   name_id_format             = "name_id_format"
  #   first_name                 = "testfirst"
  #   last_name                  = "testlast"
  #   email                      = "test@test.com"
  #   spectro_team               = "SpectroTeam"
  # }
}

# Import existing OIDC settings.
# import {
#   to = spectrocloud_sso.sso_setting
#   id = "5eea74e9teste0dtestd3f316:oidc" // tenant-uid:oidc
# }

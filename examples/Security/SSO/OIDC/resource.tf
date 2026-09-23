data "spectrocloud_team" "team" {
  name = "Tenant Admin"
}

# Manages tenant-wide SSO configuration (singleton per tenant), configured for OIDC. Day-2
# mutability: nothing here is ForceNew - every attribute, including the nested oidc block,
# updates in place.
#
# `sso_auth_type = "oidc"` requires the oidc block below and forbids a saml block. Flat
# attributes (sso_auth_type/domains/auth_providers) are the same as ../SAML/resource.tf.
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

  # saml is mutually exclusive with oidc above - only one may be set, matching sso_auth_type.
  # See ../SAML/resource.tf for a full saml block example.
}

# Import existing OIDC settings.
# import {
#   to = spectrocloud_sso.sso_setting
#   id = "5eea74e9teste0dtestd3f316:oidc" // tenant-uid:oidc
# }

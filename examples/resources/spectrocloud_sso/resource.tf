
data "spectrocloud_team" "team" {
  name = "Tenant Admin"
}

# Manages tenant-wide SSO configuration (singleton per tenant). Day-2 mutability: nothing here
# is ForceNew - every attribute, including the nested saml/oidc blocks, updates in place.
#
# `sso_auth_type` gates which of `saml`/`oidc` is required: "saml" requires the saml block and
# forbids oidc; "oidc" requires the oidc block and forbids saml; "none" forbids both.
resource "spectrocloud_sso" "sso_setting" {
  # Optional, default "none". Allowed: "none", "saml", "oidc".
  sso_auth_type = "saml"
  # Optional. Email domains that are routed through this SSO configuration.
  domains = ["test.com", "test-login.com"]
  # Optional. External auth providers to also allow. Allowed values: "github", "google".
  auth_providers = ["github", "google"]

  saml {
    # Required. Allowed: "Azure Active Directory", "Okta", "Keycloak", "OneLogin",
    # "Microsoft ADFS", "Others".
    service_provider = "Microsoft ADFS"
    # Required. Metadata XML from the SAML identity provider.
    identity_provider_metadata = "<note>test</note>"
    # Optional. Teams new SSO users are added to by default.
    default_team_ids = [data.spectrocloud_team.team.id]
    # Optional, default false. Enables SAML single logout.
    enable_single_logout = true
    # Required. NameID format expected in SAML responses.
    name_id_format = "name_id_format"
    # Optional, default "FirstName"/"LastName"/"Email"/"SpectroTeam". Names of the SAML claims
    # that carry the user's first name, last name, email, and team/group membership.
    first_name   = "testfirst"
    last_name    = "testlast"
    email        = "test@test.com"
    spectro_team = "SpectroTeam"

    # Computed, read-only: issuer, certificate, single_logout_url, entity_id, login_url,
    # service_provider_metadata are populated by Palette after apply - do not set them.
  }

  # OIDC is mutually exclusive with saml above - only one may be set, matching sso_auth_type.
  # oidc {
  #   issuer_url                       = "https://login.microsoftonline.com/sd8/v2.0"
  #   identity_provider_ca_certificate = "test certificate content"
  #   insecure_skip_tls_verify         = false # only for trusted networks with self-signed certs
  #   client_id                        = ""
  #   client_secret                   = ""    # credential material
  #   default_team_ids                = [data.spectrocloud_team.team.id]
  #   scopes                          = ["profile", "email"]
  #   first_name                      = "test"
  #   last_name                       = "last"
  #   email                           = "test@test.com"
  #   spectro_team                    = "groups"
  #   # Optional: query the OIDC userinfo endpoint for these claims when the ID token doesn't
  #   # carry them.
  #   user_info_endpoint {
  #     first_name   = "test"
  #     last_name    = "last"
  #     email        = "test@test.com"
  #     spectro_team = "groups"
  #   }
  # }
}

## import existing sso settings
## when importing either we can import saml or oidc
#import {
#  to = spectrocloud_sso.sso_setting
#  id = "5eea74e9teste0dtestd3f316:saml" // tenant-uid:saml or tenant-uid:oidc
#}
data "spectrocloud_team" "team" {
  name = "Tenant Admin"
}

# Manages tenant-wide SSO configuration (singleton per tenant), configured for SAML. Day-2
# mutability: nothing here is ForceNew - every attribute, including the nested saml block,
# updates in place.
#
# `sso_auth_type = "saml"` requires the saml block below and forbids an oidc block.
#
# Flat attributes:
#   sso_auth_type  - Optional, default "none". Allowed: "none", "saml", "oidc".
#   domains        - Optional. Email domains that are routed through this SSO configuration.
#   auth_providers - Optional. External auth providers to also allow. Allowed values: "github",
#                    "google".
resource "spectrocloud_sso" "sso_setting" {
  sso_auth_type  = "saml"
  domains        = ["test.com", "test-login.com"]
  auth_providers = ["github", "google"]

  # saml block (Required when sso_auth_type = "saml", at most one):
  #   service_provider           - Required. Allowed: "Azure Active Directory", "Okta", "Keycloak",
  #                                "OneLogin", "Microsoft ADFS", "Others".
  #   identity_provider_metadata - Required. Metadata XML from the SAML identity provider.
  #   default_team_ids           - Optional. Teams new SSO users are added to by default.
  #   enable_single_logout       - Optional, default false. Enables SAML single logout.
  #   name_id_format             - Required. NameID format expected in SAML responses.
  #   first_name, last_name, email, spectro_team - Optional, default "FirstName"/"LastName"/
  #                                "Email"/"SpectroTeam". Names of the SAML claims that carry the
  #                                user's first name, last name, email, and team/group membership.
  #
  #   Computed, read-only: issuer, certificate, single_logout_url, entity_id, login_url,
  #   service_provider_metadata are populated by Palette after apply - do not set them.
  saml {
    service_provider           = "Microsoft ADFS"
    identity_provider_metadata = "<note>test</note>"
    default_team_ids           = [data.spectrocloud_team.team.id]
    enable_single_logout       = true
    name_id_format             = "name_id_format"
    first_name                 = "testfirst"
    last_name                  = "testlast"
    email                      = "test@test.com"
    spectro_team               = "SpectroTeam"
  }
}

# Import existing SAML settings.
# import {
#   to = spectrocloud_sso.sso_setting
#   id = "5eea74e9teste0dtestd3f316:saml" // tenant-uid:saml
# }

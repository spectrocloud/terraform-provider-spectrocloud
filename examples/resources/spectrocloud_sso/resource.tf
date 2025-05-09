
data "spectrocloud_team" "team" {
  name = "Tenant Admin"
}

resource "spectrocloud_sso" "sso_setting" {
  sso_auth_type  = "saml" # oidc or none
  domains        = ["test.com", "test-login.com"]
  auth_providers = ["github", "google"]
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
  #  oidc {
  #    issuer_url = "https://login.microsoftonline.com/sd8/v2.0"
  #    identity_provider_ca_certificate = "test certificate content"
  #    insecure_skip_tls_verify = false
  #    client_id = ""
  #    client_secret = ""
  #    default_team_ids = [data.spectrocloud_team.team.id]
  #    scopes = ["profile", "email"]
  #    first_name = "test"
  #    last_name = "last"
  #    email = "test@test.com"
  #    spectro_team = "groups"
  #    user_info_endpoint {
  #      first_name = "test"
  #      last_name = "last"
  #      email = "test@test.com"
  #      spectro_team = "groups"
  #    }
  #  }
}

## import existing sso settings
## when importing either we can import saml or oidc
#import {
#  to = spectrocloud_sso.sso_setting
#  id = "5eea74e9teste0dtestd3f316:saml" // tenant-uid:saml or tenant-uid:oidc
#}
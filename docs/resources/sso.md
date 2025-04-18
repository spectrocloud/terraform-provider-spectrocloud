---
page_title: "spectrocloud_sso Resource - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  
---

# spectrocloud_sso (Resource)

  

Palette supports Single Sign-On (SSO) with a variety of Identity Providers (IDP). You can enable SSO in Palette by using the following protocols for authentication and authorization.[SSO Setting](https://docs.spectrocloud.com/user-management/saml-sso/) guide.

~> The spectrocloud_sso resource enforces Single Sign-On (SSO) settings. By default, it is configured with Palette’s default values. Users can customize settings as needed. Destroying the spectrocloud_sso resource SSO set to none.

## Example Usage

An example of managing an developer setting in Palette.

```hcl

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

```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `auth_providers` (Set of String) A set of external authentication providers such as GitHub and Google.
- `domains` (Set of String) A set of domains associated with the SSO configuration.
- `oidc` (Block List, Max: 1) (see [below for nested schema](#nestedblock--oidc))
- `saml` (Block List, Max: 1) Configuration for Security Assertion Markup Language (SAML) authentication. (see [below for nested schema](#nestedblock--saml))
- `sso_auth_type` (String) Defines the type of SSO authentication. Supported values: none, saml, oidc.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--oidc"></a>
### Nested Schema for `oidc`

Required:

- `client_id` (String) Client ID for OIDC authentication.
- `client_secret` (String, Sensitive) Client secret for OIDC authentication (sensitive).
- `email` (String) User's email address retrieved from identity provider.
- `first_name` (String) User's first name retrieved from identity provider.
- `issuer_url` (String) URL of the OIDC issuer.
- `last_name` (String) User's last name retrieved from identity provider.
- `scopes` (Set of String) Scopes requested during OIDC authentication.
- `spectro_team` (String) The SpectroCloud team the user belongs to.

Optional:

- `default_team_ids` (Set of String) A set of default team IDs assigned to users.
- `identity_provider_ca_certificate` (String) Certificate authority (CA) certificate for the identity provider.
- `insecure_skip_tls_verify` (Boolean) Boolean to skip TLS verification for identity provider communication.
- `user_info_endpoint` (Block List, Max: 1) To allow Palette to query the OIDC userinfo endpoint using the provided Issuer URL. Palette will first attempt to retrieve role and group information from userInfo endpoint. If unavailable, Palette will fall back to using Required Claims as specified above. Use the following fields to specify what Required Claims Palette will include when querying the userinfo endpoint. (see [below for nested schema](#nestedblock--oidc--user_info_endpoint))

Read-Only:

- `callback_url` (String) URL to which the identity provider redirects after authentication.
- `logout_url` (String) URL used for logging out of the OIDC session.

<a id="nestedblock--oidc--user_info_endpoint"></a>
### Nested Schema for `oidc.user_info_endpoint`

Required:

- `email` (String) User's email address retrieved from identity provider.
- `first_name` (String) User's first name retrieved from identity provider.
- `last_name` (String) User's last name retrieved from identity provider.
- `spectro_team` (String) The SpectroCloud team the user belongs to.



<a id="nestedblock--saml"></a>
### Nested Schema for `saml`

Required:

- `identity_provider_metadata` (String) Metadata XML of the SAML identity provider.
- `name_id_format` (String) Format of the NameID attribute in SAML responses.
- `service_provider` (String) The identity provider service used for SAML authentication.

Optional:

- `default_team_ids` (Set of String) A set of default team IDs assigned to users.
- `email` (String) User's email address retrieved from identity provider.
- `enable_single_logout` (Boolean) Boolean to enable SAML single logout feature.
- `first_name` (String) User's first name retrieved from identity provider.
- `last_name` (String) User's last name retrieved from identity provider.
- `spectro_team` (String) The SpectroCloud team the user belongs to.

Read-Only:

- `certificate` (String) Certificate for SAML authentication.
- `entity_id` (String) Entity ID used to identify the service provider.
- `issuer` (String) SAML identity provider issuer URL.
- `login_url` (String) Login URL for the SAML identity provider.
- `service_provider_metadata` (String) Metadata XML of the SAML service provider.
- `single_logout_url` (String) URL used for initiating SAML single logout.


<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `delete` (String)
- `update` (String)
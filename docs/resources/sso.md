---
page_title: "spectrocloud_sso Resource - terraform-provider-spectrocloud"
subcategory: "SSO"
description: |-
  Resource for managing tenant-level single sign-on configuration in Spectro Cloud.
---

# spectrocloud_sso (Resource)

  Resource for managing tenant-level single sign-on configuration in Spectro Cloud.

Palette supports Single Sign-On (SSO) with a variety of Identity Providers (IDP). You can enable SSO in Palette by using the following protocols for authentication and authorization. See the [SSO Setting](https://docs.spectrocloud.com/user-management/saml-sso/) guide.

~> The `spectrocloud_sso` resource enforces Single Sign-On (SSO) settings tenant-wide (singleton per tenant). By default, it is configured with Palette's default values. Destroying the `spectrocloud_sso` resource sets SSO back to `none`.

## Example Usage

`sso_auth_type` selects the protocol - `none` (default), `saml`, or `oidc` - and gates which of the `saml`/`oidc` blocks is required; the two are mutually exclusive.

### SAML

```terraform
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
```

### OIDC

```terraform
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
```

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import)
to import the tenant's SSO settings, using the tenant UID and the protocol currently configured:

```terraform
import {
  to = spectrocloud_sso.sso_setting
  id = "5eea74e9teste0dtestd3f316:saml" // tenant-uid:saml or tenant-uid:oidc
}
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
- `email` (String) The name of the claim that returns the user's email address from the identity provider.
- `first_name` (String) The name of the claim that returns the user's first name from the identity provider.
- `issuer_url` (String) URL of the OIDC issuer.
- `last_name` (String) The name of the claim that returns the user's last name from the identity provider.
- `scopes` (Set of String) Set of OIDC scope strings requested during authentication.
- `spectro_team` (String) The name of the claim that returns the user's group memberships from the Identity Provider. The values of this claim will map to SpectroCloud teams.

Optional:

- `default_team_ids` (Set of String) A set of default team IDs assigned to users.
- `identity_provider_ca_certificate` (String) Certificate authority (CA) certificate for the identity provider.
- `insecure_skip_tls_verify` (Boolean) Boolean to skip TLS verification for identity provider communication. ⚠️ WARNING: Setting this to true disables SSL certificate verification and makes connections vulnerable to man-in-the-middle attacks. Only use this when connecting to identity providers with self-signed certificates in trusted networks.
- `user_info_endpoint` (Block List, Max: 1) To allow Palette to query the OIDC userinfo endpoint using the provided Issuer URL. Palette will first attempt to retrieve role and group information from userInfo endpoint. If unavailable, Palette will fall back to using Required Claims as specified above. Use the following fields to specify what Required Claims Palette will include when querying the userinfo endpoint. (see [below for nested schema](#nestedblock--oidc--user_info_endpoint))

Read-Only:

- `callback_url` (String) URL to which the identity provider redirects after authentication.
- `logout_url` (String) URL used for logging out of the OIDC session.

<a id="nestedblock--oidc--user_info_endpoint"></a>
### Nested Schema for `oidc.user_info_endpoint`

Required:

- `first_name` (String) The name of the claim that returns the user's first name from the identity provider.
- `last_name` (String) The name of the claim that returns the user's last name from the identity provider.
- `spectro_team` (String) The name of the claim that returns the user's group memberships from the Identity Provider. The values of this claim will map to SpectroCloud teams.

Optional:

- `email` (String) The name of the claim that returns the user's email address from the identity provider.



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

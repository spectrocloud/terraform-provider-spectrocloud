package spectrocloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"time"
)

func resourceSSO() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSSOCreate,
		ReadContext:   resourceSSORead,
		UpdateContext: resourceSSOUpdate,
		DeleteContext: resourceSSODelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		SchemaVersion: 2,
		Schema: map[string]*schema.Schema{
			"sso_auth_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "none",
				ValidateFunc: validation.StringInSlice([]string{"none", "saml", "oidc"}, false),
				Description:  "Defines the type of SSO authentication. Supported values: none, saml, oidc.",
			},
			"domains": {
				Type:        schema.TypeSet,
				Optional:    true,
				Default:     "",
				Set:         schema.HashString,
				Description: "A set of domains associated with the SSO configuration.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"auth_providers": {
				Type:        schema.TypeSet,
				Optional:    true,
				Default:     "none",
				Set:         schema.HashString,
				Description: "A set of external authentication providers such as GitHub and Google.",
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{"GitHub", "Google"}, false),
				},
			},
			"oidc": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"issuer_url": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "https://www.spectrocloud.com",
							Description: "URL of the OIDC issuer.",
						},
						"identity_provider_ca_certificate": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "Certificate authority (CA) certificate for the identity provider.",
						},
						"insecure_skip_tls_verify": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Boolean to skip TLS verification for identity provider communication.",
						},
						"client_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Client ID for OIDC authentication.",
						},
						"client_secret": {
							Type:        schema.TypeString,
							Required:    true,
							Sensitive:   true,
							Description: "Client secret for OIDC authentication (sensitive).",
						},
						"callback_url": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "URL to which the identity provider redirects after authentication.",
						},
						"logout_url": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "URL used for logging out of the OIDC session.",
						},
						"default_team_ids": {
							Type:     schema.TypeSet,
							Optional: true,
							Set:      schema.HashString,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "A set of default team IDs assigned to users.",
						},
						"scopes": {
							Type:     schema.TypeSet,
							Optional: true,
							Set:      schema.HashString,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "Scopes requested during OIDC authentication.",
						},
						"first_name": {
							Type:        schema.TypeString,
							Required:    true,
							Default:     "",
							Description: "User's first name retrieved from identity provider.",
						},
						"last_name": {
							Type:        schema.TypeString,
							Required:    true,
							Default:     "",
							Description: "User's last name retrieved from identity provider.",
						},
						"email": {
							Type:        schema.TypeString,
							Required:    true,
							Default:     "",
							Description: "User's email address retrieved from identity provider.",
						},
						"spectro_team": {
							Type:        schema.TypeString,
							Required:    true,
							Default:     "",
							Description: "The SpectroCloud team the user belongs to.",
						},
						"user_info_endpoint": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "To allow Palette to query the OIDC userinfo endpoint using the provided Issuer URL. Palette will first attempt to retrieve role and group information from userInfo endpoint. If unavailable, Palette will fall back to using Required Claims as specified above. Use the following fields to specify what Required Claims Palette will include when querying the userinfo endpoint.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"first_name": {
										Type:        schema.TypeString,
										Required:    true,
										Default:     "",
										Description: "User's first name retrieved from identity provider.",
									},
									"last_name": {
										Type:        schema.TypeString,
										Required:    true,
										Default:     "",
										Description: "User's last name retrieved from identity provider.",
									},
									"email": {
										Type:        schema.TypeString,
										Required:    true,
										Default:     "",
										Description: "User's email address retrieved from identity provider.",
									},
									"spectro_team": {
										Type:        schema.TypeString,
										Required:    true,
										Default:     "",
										Description: "The SpectroCloud team the user belongs to.",
									},
								},
							},
						},
					},
				},
			},
			"saml": {
				Type:        schema.TypeSet,
				Optional:    true,
				MaxItems:    1,
				Description: "Configuration for Security Assertion Markup Language (SAML) authentication.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"issuer": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "https://www.spectrocloud.com",
							Description: "SAML identity provider issuer URL.",
						},
						"certificate": {
							Type:        schema.TypeString,
							Required:    true,
							Default:     "",
							Description: "Certificate for SAML authentication.",
						},
						"service": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"Azure Active Directory", "Okta", "Keycloak", "OneLogin", "Microsoft ADFS", "Others"}, false),
							Description:  "The identity provider service used for SAML authentication.",
						},
						"identity_provider_metadata": {
							Type:        schema.TypeString,
							Required:    true,
							Default:     "",
							Description: "Metadata XML of the SAML identity provider.",
						},
						"default_team_ids": {
							Type:     schema.TypeSet,
							Optional: true,
							Set:      schema.HashString,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "A set of default team IDs assigned to users.",
						},
						"enable_single_logout": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Boolean to enable SAML single logout feature.",
						},
						"single_logout_url": {
							Type:        schema.TypeString,
							Required:    true,
							Default:     "",
							Description: "URL used for initiating SAML single logout.",
						},
						"entity_id": {
							Type:        schema.TypeString,
							Required:    true,
							Default:     "https://www.spectrocloud.com",
							Description: "Entity ID used to identify the service provider.",
						},
						"name_id_format": {
							Type:        schema.TypeString,
							Required:    true,
							Default:     "",
							Description: "Format of the NameID attribute in SAML responses.",
						},
						"login_url": {
							Type:        schema.TypeString,
							Required:    true,
							Default:     "",
							Description: "Login URL for the SAML identity provider.",
						},
						"first_name": {
							Type:        schema.TypeString,
							Required:    true,
							Default:     "",
							Description: "User's first name retrieved from identity provider.",
						},
						"last_name": {
							Type:        schema.TypeString,
							Required:    true,
							Default:     "",
							Description: "User's last name retrieved from identity provider.",
						},
						"email": {
							Type:        schema.TypeString,
							Required:    true,
							Default:     "",
							Description: "User's email address retrieved from identity provider.",
						},
						"spectro_team": {
							Type:        schema.TypeString,
							Required:    true,
							Default:     "",
							Description: "The SpectroCloud team the user belongs to.",
						},
						"service_provider_metadata": {
							Type:        schema.TypeString,
							Required:    true,
							Default:     "",
							Description: "Metadata XML of the SAML service provider.",
						},
					},
				},
			},
		},
		CustomizeDiff: customDiffValidation,
	}
}

func customDiffValidation(ctx context.Context, diff *schema.ResourceDiff, v interface{}) error {
	ssoAuthType, ok := diff.GetOk("sso_auth_type")
	if !ok {
		return nil // No validation needed if not set
	}

	authType := ssoAuthType.(string)
	samlExists := diff.Get("saml") != nil
	oidcExists := diff.Get("oidc") != nil

	switch authType {
	case "none":
		if samlExists || oidcExists {
			return fmt.Errorf("sso_auth_type is set to 'none', so 'saml' and 'oidc' should not be defined")
		}
	case "saml":
		if oidcExists {
			return fmt.Errorf("sso_auth_type is set to 'saml', so 'oidc' should not be defined")
		}
	case "oidc":
		if samlExists {
			return fmt.Errorf("sso_auth_type is set to 'oidc', so 'saml' should not be defined")
		}
	}

	return nil
}

func toOIDC(d *schema.ResourceData) *models.V1TenantOidcClientSpec {

	oidcSpec := &models.V1TenantOidcClientSpec{
		CallbackURL:     "",
		ClientID:        "",
		ClientSecret:    "",
		DefaultTeams:    nil,
		IsSsoEnabled:    false,
		IssuerTLS:       nil,
		IssuerURL:       "",
		LogoutURL:       "",
		RequiredClaims:  nil,
		Scopes:          nil,
		ScopesDelimiter: "",
		SyncSsoTeams:    false,
		UserInfo:        nil,
	}
	return oidcSpec
}

func toSAML(d *schema.ResourceData) *models.V1TenantSamlRequestSpec {
	samlSpec := &models.V1TenantSamlRequestSpec{
		Attributes:            nil,
		DefaultTeams:          nil,
		FederationMetadata:    "",
		IdentityProvider:      "",
		IsSingleLogoutEnabled: false,
		IsSsoEnabled:          false,
		NameIDFormat:          "",
		SyncSsoTeams:          false,
	}
	return samlSpec
}

func toDomains(d *schema.ResourceData) *models.V1TenantDomains {
	domainSpec := &models.V1TenantDomains{
		Domains: []string{},
	}
	return domainSpec
}

func toAuthProviders(d *schema.ResourceData) *models.V1TenantSsoAuthProvidersEntity {
	authProviderSpec := &models.V1TenantSsoAuthProvidersEntity{
		IsEnabled: false,
		SsoLogins: nil,
	}
	return authProviderSpec
}

func resourceSSOCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "tenant")
	var diags diag.Diagnostics
	tenantUID, err := c.GetTenantUID()
	if err != nil {
		return diag.FromErr(err)
	}
	ssoType := d.Get("sso_auth_type").(string)
	switch ssoType {
	case "none":
		//
	case "saml":
		//
	case "oidc":

	}
	return diags
}

func resourceSSORead(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {

}

func resourceSSOUpdate(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {

}

func resourceSSODelete(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {

}

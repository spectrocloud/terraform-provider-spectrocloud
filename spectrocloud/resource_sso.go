package spectrocloud

import (
	"context"
	"encoding/base64"
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
				Set:         schema.HashString,
				Description: "A set of domains associated with the SSO configuration.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"auth_providers": {
				Type:        schema.TypeSet,
				Optional:    true,
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
							Computed:    true,
							Description: "SAML identity provider issuer URL.",
						},
						"certificate": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Certificate for SAML authentication.",
						},
						"service_provider": {
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
							Computed:    true,
							Description: "URL used for initiating SAML single logout.",
						},
						"entity_id": {
							Type:        schema.TypeString,
							Computed:    true,
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
							Computed:    true,
							Description: "Login URL for the SAML identity provider.",
						},
						"first_name": {
							Type:        schema.TypeString,
							Required:    true,
							Default:     "FirstName",
							Description: "User's first name retrieved from identity provider.",
						},
						"last_name": {
							Type:        schema.TypeString,
							Required:    true,
							Default:     "LastName",
							Description: "User's last name retrieved from identity provider.",
						},
						"email": {
							Type:        schema.TypeString,
							Required:    true,
							Default:     "Email",
							Description: "User's email address retrieved from identity provider.",
						},
						"spectro_team": {
							Type:        schema.TypeString,
							Required:    true,
							Default:     "SpectroTeam",
							Description: "The SpectroCloud team the user belongs to.",
						},
						"service_provider_metadata": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Metadata XML of the SAML service provider.",
						},
					},
				},
			},
		},
		CustomizeDiff: customDiffValidation,
	}
}

func resourceCommonUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, tenantString)
	var diags diag.Diagnostics
	tenantUID, err := c.GetTenantUID()
	if err != nil {
		return diag.FromErr(err)
	}
	ssoType := d.Get("sso_auth_type").(string)
	switch ssoType {
	case "none":
		// disable
		samlEntity := toSAML(d)
		samlEntity.IsSsoEnabled = false
		err = c.UpdateSAML(tenantUID, samlEntity)
		if err != nil {
			return diag.FromErr(err)
		}

		oidcEntity := toOIDC(d)
		oidcEntity.IsSsoEnabled = false
		err = c.UpdateOIDC(tenantUID, oidcEntity)
		if err != nil {
			return diag.FromErr(err)
		}
	case "saml":
		samlEntity := toSAML(d)
		err = c.UpdateSAML(tenantUID, samlEntity)
		if err != nil {
			return diag.FromErr(err)
		}
	case "oidc":
		oidcEntity := toOIDC(d)
		err = c.UpdateOIDC(tenantUID, oidcEntity)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if v, ok := d.GetOk("domains"); ok {
		err = c.UpdateDomain(tenantUID, toDomains(v.([]interface{})))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if v, ok := d.GetOk("auth_providers"); ok {
		err = c.UpdateProviders(tenantUID, toAuthProviders(v.([]interface{})))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func resourceSSOCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceCommonUpdate(ctx, d, m)
}

func resourceSSORead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, tenantString)
	var diags diag.Diagnostics
	tenantUID, err := c.GetTenantUID()
	if err != nil {
		return diag.FromErr(err)
	}

	samlEntity, err := c.GetSAML(tenantUID)
	if err != nil {
		return diag.FromErr(err)
	}
	err = flattenSAML(samlEntity, d)
	if err != nil {
		return diag.FromErr(err)
	}

	oidcEntity, err := c.GetOIDC(tenantUID)
	if err != nil {
		return diag.FromErr(err)
	}
	err = flattenOidc(oidcEntity, d)
	if err != nil {
		return diag.FromErr(err)
	}

	domainsEntity, err := c.GetDomains(tenantUID)
	if err != nil {
		return diag.FromErr(err)
	}
	err = flattenDomains(domainsEntity, d)
	if err != nil {
		return diag.FromErr(err)
	}

	authEntity, err := c.GetProviders(tenantUID)
	if err != nil {
		return diag.FromErr(err)
	}
	err = flattenAuthProviders(authEntity, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceSSOUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceCommonUpdate(ctx, d, m)
}

func resourceSSODelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, tenantString)
	var diags diag.Diagnostics
	tenantUID, err := c.GetTenantUID()
	if err != nil {
		return diag.FromErr(err)
	}

	// disable saml
	samlEntity := toSAML(d)
	samlEntity.IsSsoEnabled = false
	err = c.UpdateSAML(tenantUID, samlEntity)
	if err != nil {
		return diag.FromErr(err)
	}
	// disable oidc
	oidcEntity := toOIDC(d)
	oidcEntity.IsSsoEnabled = false
	err = c.UpdateOIDC(tenantUID, oidcEntity)
	if err != nil {
		return diag.FromErr(err)
	}
	// disable domain
	err = c.UpdateDomain(tenantUID, toDomains([]interface{}{}))
	if err != nil {
		return diag.FromErr(err)
	}
	// disable auth provider
	err = c.UpdateProviders(tenantUID, toAuthProviders([]interface{}{}))
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
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

func toStringSlice(input []interface{}) []string {
	result := make([]string, len(input))
	for i, v := range input {
		result[i] = v.(string)
	}
	return result
}

func toOIDC(d *schema.ResourceData) *models.V1TenantOidcClientSpec {
	oidcSpec := &models.V1TenantOidcClientSpec{}

	oidcSpec.CallbackURL = d.Get("callback_url").(string)
	oidcSpec.ClientID = d.Get("client_id").(string)
	oidcSpec.ClientSecret = d.Get("client_secret").(string)
	oidcSpec.DefaultTeams = toStringSlice(d.Get("default_team_ids").([]interface{}))
	oidcSpec.IsSsoEnabled = true
	oidcSpec.IssuerTLS = &models.V1OidcIssuerTLS{
		CaCertificateBase64: d.Get("identity_provider_ca_certificate").(string),
		InsecureSkipVerify:  BoolPtr(d.Get("insecure_skip_tls_verify").(bool)),
	}
	oidcSpec.IssuerURL = d.Get("issuer_url").(string)
	oidcSpec.LogoutURL = d.Get("logout_url").(string)
	oidcSpec.RequiredClaims = &models.V1TenantOidcClaims{
		Email:       d.Get("email").(string),
		FirstName:   d.Get("first_name").(string),
		LastName:    d.Get("last_name").(string),
		SpectroTeam: d.Get("spectro_team").(string),
	}
	oidcSpec.Scopes = toStringSlice(d.Get("scopes").([]interface{}))
	oidcSpec.ScopesDelimiter = ""
	oidcSpec.SyncSsoTeams = true

	if uie, ok := d.GetOk("user_info_endpoint"); ok {
		oidcSpec.UserInfo = &models.V1OidcUserInfo{
			Claims: &models.V1TenantOidcClaims{
				Email:       uie.([]interface{})[0].(map[string]interface{})["email"].(string),
				FirstName:   uie.([]interface{})[0].(map[string]interface{})["first_name"].(string),
				LastName:    uie.([]interface{})[0].(map[string]interface{})["last_name"].(string),
				SpectroTeam: uie.([]interface{})[0].(map[string]interface{})["spectro_team"].(string),
			},
			UseUserInfo: BoolPtr(true),
		}
	}
	return oidcSpec
}

func flattenOidc(oidcSpec *models.V1TenantOidcClientSpec, d *schema.ResourceData) error {
	var err error
	var oidc []interface{}
	var spec map[string]interface{}

	spec["callback_url"] = oidcSpec.CallbackURL
	spec["client_id"] = oidcSpec.ClientID
	spec["client_secret"] = oidcSpec.ClientSecret

	spec["default_team_ids"] = oidcSpec.DefaultTeams
	spec["identity_provider_ca_certificate"] = oidcSpec.IssuerTLS.CaCertificateBase64
	spec["insecure_skip_tls_verify"] = oidcSpec.IssuerTLS.InsecureSkipVerify
	spec["issuer_url"] = oidcSpec.IssuerURL
	spec["logout_url"] = oidcSpec.LogoutURL

	spec["email"] = oidcSpec.RequiredClaims.Email
	spec["first_name"] = oidcSpec.RequiredClaims.FirstName
	spec["last_name"] = oidcSpec.RequiredClaims.LastName
	spec["spectro_team"] = oidcSpec.RequiredClaims.SpectroTeam
	spec["scopes"] = oidcSpec.Scopes

	var userEndpoint []interface{}
	userEndpoint = append(userEndpoint, map[string]interface{}{
		"email":        oidcSpec.UserInfo.Claims.Email,
		"first_name":   oidcSpec.UserInfo.Claims.FirstName,
		"last_name":    oidcSpec.UserInfo.Claims.LastName,
		"spectro_team": oidcSpec.UserInfo.Claims.SpectroTeam,
	})
	spec["user_info_endpoint"] = userEndpoint

	oidc = append(oidc, spec)
	if err = d.Set("oidc", oidc); err != nil {
		return err
	}
	return err
}

func toSAML(d *schema.ResourceData) *models.V1TenantSamlRequestSpec {
	samlSpec := &models.V1TenantSamlRequestSpec{}

	// Attributes
	firstNameTemplate := &models.V1TenantSamlSpecAttribute{
		AttributeValue:  "[FIRST_NAME_OF_USER]",
		MappedAttribute: "FirstName",
		Name:            d.Get("first_name").(string),
		NameFormat:      "urn:oasis:names:tc:SAML:2.0:attrname-format:unspecified",
	}
	lastNameTemplate := &models.V1TenantSamlSpecAttribute{
		AttributeValue:  "[LAST_NAME_OF_USER]",
		MappedAttribute: "LastName",
		Name:            d.Get("last_name").(string),
		NameFormat:      "urn:oasis:names:tc:SAML:2.0:attrname-format:unspecified",
	}
	emailTemplate := &models.V1TenantSamlSpecAttribute{
		AttributeValue:  "[EMAIL_OF_USER]",
		MappedAttribute: "Email",
		Name:            d.Get("email").(string),
		NameFormat:      "urn:oasis:names:tc:SAML:2.0:attrname-format:unspecified",
	}
	spectroTeamTemplate := &models.V1TenantSamlSpecAttribute{
		AttributeValue:  "[SPECTRO_TEAM_OF_USER]",
		MappedAttribute: "SpectroTeam",
		Name:            d.Get("spectro_team").(string),
		NameFormat:      "urn:oasis:names:tc:SAML:2.0:attrname-format:unspecified",
	}

	attributes := make([]*models.V1TenantSamlSpecAttribute, 0)
	attributes = append(attributes, firstNameTemplate)
	attributes = append(attributes, lastNameTemplate)
	attributes = append(attributes, emailTemplate)
	attributes = append(attributes, spectroTeamTemplate)
	samlSpec.Attributes = attributes

	samlSpec.DefaultTeams = toStringSlice(d.Get("default_team_ids").([]interface{}))
	samlSpec.FederationMetadata = base64.StdEncoding.EncodeToString([]byte(d.Get("identity_provider_metadata").(string)))
	samlSpec.IdentityProvider = d.Get("service_provider").(string)
	samlSpec.IsSingleLogoutEnabled = d.Get("enable_single_logout").(bool)
	samlSpec.IsSsoEnabled = true
	samlSpec.NameIDFormat = d.Get("name_id_format").(string)
	samlSpec.SyncSsoTeams = false

	return samlSpec
}

func flattenSAML(saml *models.V1TenantSamlSpec, d *schema.ResourceData) error {
	var err error
	var samlData []interface{}
	var spec map[string]interface{}
	spec["issuer"] = saml.Issuer
	spec["certificate"] = saml.Certificate
	spec["service_provider"] = saml.IdentityProvider
	spec["identity_provider_metadata"] = saml.FederationMetadata
	spec["default_team_ids"] = saml.DefaultTeams
	spec["enable_single_logout"] = saml.IsSingleLogoutEnabled
	spec["single_logout_url"] = saml.SingleLogoutURL
	spec["entity_id"] = saml.EntityID
	spec["name_id_format"] = saml.NameIDFormat
	spec["login_url"] = saml.AudienceURL
	spec["service_provider_metadata"] = saml.ServiceProviderMetadata

	var attribute map[string]interface{}

	for _, a := range saml.Attributes {
		attribute[a.MappedAttribute] = a
	}
	spec["first_name"] = attribute["FirstName"].(*models.V1TenantSamlSpecAttribute).AttributeValue
	spec["last_name"] = attribute["LastName"].(*models.V1TenantSamlSpecAttribute).AttributeValue
	spec["email"] = attribute["Email"].(*models.V1TenantSamlSpecAttribute).AttributeValue
	spec["spectro_team"] = attribute["SpectroTeam"].(*models.V1TenantSamlSpecAttribute).AttributeValue
	samlData = append(samlData, spec)
	if err = d.Set("saml", saml); err != nil {
		return err
	}
	return err
}

func toDomains(domains []interface{}) *models.V1TenantDomains {
	domainSpec := &models.V1TenantDomains{
		Domains: toStringSlice(domains),
	}
	return domainSpec
}

func flattenDomains(domains *models.V1TenantDomains, d *schema.ResourceData) error {
	var err error
	if domains.Domains != nil {
		err = d.Set("domains", domains.Domains)
	}
	return err
}

func toAuthProviders(providers []interface{}) *models.V1TenantSsoAuthProvidersEntity {
	authProviderSpec := &models.V1TenantSsoAuthProvidersEntity{
		IsEnabled: true,
		SsoLogins: toStringSlice(providers),
	}
	return authProviderSpec
}

func flattenAuthProviders(authProviders *models.V1TenantSsoAuthProvidersEntity, d *schema.ResourceData) error {
	var err error
	if authProviders != nil {
		err = d.Set("auth_providers", authProviders.SsoLogins)
	}
	return err
}

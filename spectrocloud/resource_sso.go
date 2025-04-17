package spectrocloud

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
	"regexp"
	"strings"
	"time"
)

func resourceSSO() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSSOCreate,
		ReadContext:   resourceSSORead,
		UpdateContext: resourceSSOUpdate,
		DeleteContext: resourceSSODelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceSSOImport,
		},

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
					ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
						v := val.(string)

						if strings.TrimSpace(v) == "" {
							errs = append(errs, fmt.Errorf("%q must not be empty", key))
							return
						}

						// Basic domain regex (not exhaustive but covers common cases)
						domainRegex := `^([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$`
						if matched, _ := regexp.MatchString(domainRegex, v); !matched {
							errs = append(errs, fmt.Errorf("%q must be a valid domain name", key))
						}
						return
					},
				},
			},
			"auth_providers": {
				Type:        schema.TypeSet,
				Optional:    true,
				Set:         schema.HashString,
				Description: "A set of external authentication providers such as GitHub and Google.",
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{"", "github", "google"}, false),
				},
			},
			"oidc": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"issuer_url": {
							Type:        schema.TypeString,
							Required:    true,
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
							ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
								v := val.(string)
								if v == "" {
									errs = append(errs, fmt.Errorf("%q must not be empty", key))
								}
								return
							},
						},
						"client_secret": {
							Type:        schema.TypeString,
							Required:    true,
							Sensitive:   true,
							Description: "Client secret for OIDC authentication (sensitive).",
							ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
								v := val.(string)
								if v == "" {
									errs = append(errs, fmt.Errorf("%q must not be empty", key))
								}
								return
							},
						},
						"callback_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "URL to which the identity provider redirects after authentication.",
						},
						"logout_url": {
							Type:        schema.TypeString,
							Computed:    true,
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
							Required: true,
							Set:      schema.HashString,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "Scopes requested during OIDC authentication.",
						},
						"first_name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "User's first name retrieved from identity provider.",
						},
						"last_name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "User's last name retrieved from identity provider.",
						},
						"email": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "User's email address retrieved from identity provider.",
							ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
								v := val.(string)
								if v == "" {
									errs = append(errs, fmt.Errorf("%q must not be empty", key))
									return
								}
								emailRegex := `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
								matched, err := regexp.MatchString(emailRegex, v)
								if err != nil || !matched {
									errs = append(errs, fmt.Errorf("%q must be a valid email address", key))
								}
								return
							},
						},
						"spectro_team": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The SpectroCloud team the user belongs to.",
						},
						"user_info_endpoint": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "To allow Palette to query the OIDC userinfo endpoint using the provided Issuer URL. Palette will first attempt to retrieve role and group information from userInfo endpoint. If unavailable, Palette will fall back to using Required Claims as specified above. Use the following fields to specify what Required Claims Palette will include when querying the userinfo endpoint.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"first_name": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "User's first name retrieved from identity provider.",
									},
									"last_name": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "User's last name retrieved from identity provider.",
									},
									"email": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "User's email address retrieved from identity provider.",
										ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
											v := val.(string)
											if v == "" {
												errs = append(errs, fmt.Errorf("%q must not be empty", key))
												return
											}
											emailRegex := `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
											matched, err := regexp.MatchString(emailRegex, v)
											if err != nil || !matched {
												errs = append(errs, fmt.Errorf("%q must be a valid email address", key))
											}
											return
										},
									},
									"spectro_team": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The SpectroCloud team the user belongs to.",
									},
								},
							},
						},
					},
				},
			},
			"saml": {
				Type:        schema.TypeList,
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
							Description: "Format of the NameID attribute in SAML responses.",
						},
						"login_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Login URL for the SAML identity provider.",
						},
						"first_name": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "FirstName",
							Description: "User's first name retrieved from identity provider.",
						},
						"last_name": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "LastName",
							Description: "User's last name retrieved from identity provider.",
						},
						"email": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "Email",
							Description: "User's email address retrieved from identity provider.",
							ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
								v := val.(string)
								if v == "" {
									errs = append(errs, fmt.Errorf("%q must not be empty", key))
									return
								}
								emailRegex := `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
								matched, err := regexp.MatchString(emailRegex, v)
								if err != nil || !matched {
									errs = append(errs, fmt.Errorf("%q must be a valid email address", key))
								}
								return
							},
						},
						"spectro_team": {
							Type:        schema.TypeString,
							Optional:    true,
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

func resourceSSOImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c := getV1ClientWithResourceContext(m, tenantString)
	ssoIdParts := strings.Split(d.Id(), ":")

	if len(ssoIdParts) != 2 || (ssoIdParts[1] != "saml" && ssoIdParts[1] != "oidc") {
		return nil, fmt.Errorf("invalid sso type provided, kindly use saml/oidc")
	}

	givenTenantId, ssoType := ssoIdParts[0], ssoIdParts[1]

	actualTenantId, err := c.GetTenantUID()
	if err != nil {
		return nil, err
	}

	if givenTenantId != actualTenantId {
		return nil, fmt.Errorf("tenant id is not valid with current user or invalid tenant uid provided")
	}

	if err := d.Set("sso_auth_type", ssoType); err != nil {
		return nil, err
	}

	diags := resourceSSORead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read sso settings for import: %v", diags)
	}

	domainsEntity, err := c.GetDomains(givenTenantId)
	if err != nil {
		return nil, err
	}
	if err := flattenDomains(domainsEntity, d); err != nil {
		return nil, err
	}

	authEntity, err := c.GetProviders(givenTenantId)
	if err != nil {
		return nil, err
	}
	if err := flattenAuthProviders(authEntity, d); err != nil {
		return nil, err
	}

	d.SetId("sso_settings")
	return []*schema.ResourceData{d}, nil
}

func disableSSO(c *client.V1Client, tenantUID string) error {
	// disable
	samlSpec, err := c.GetSAML(tenantUID)
	if err != nil {
		return err
	}
	samlEntity := &models.V1TenantSamlRequestSpec{
		Attributes:            samlSpec.Attributes,
		DefaultTeams:          samlSpec.DefaultTeams,
		FederationMetadata:    samlSpec.FederationMetadata,
		IdentityProvider:      samlSpec.IdentityProvider,
		IsSingleLogoutEnabled: samlSpec.IsSingleLogoutEnabled,
		IsSsoEnabled:          false,
		NameIDFormat:          samlSpec.NameIDFormat,
		SyncSsoTeams:          samlSpec.SyncSsoTeams,
	}

	samlEntity.IsSsoEnabled = false
	err = c.UpdateSAML(tenantUID, samlEntity)
	if err != nil {
		return err
	}

	oidcEntity, err := c.GetOIDC(tenantUID) // toOIDC(d)
	if err != nil {
		return err
	}
	oidcEntity.IsSsoEnabled = false
	err = c.UpdateOIDC(tenantUID, oidcEntity)
	if err != nil {
		return err
	}

	// disable domain
	err = c.UpdateDomain(tenantUID, toDomains([]interface{}{}))
	if err != nil {
		return err
	}
	// disable auth provider
	authPro := toAuthProviders([]interface{}{})
	authPro.IsEnabled = false
	err = c.UpdateProviders(tenantUID, authPro)
	if err != nil {
		return err
	}
	return err
}

func resourceCommonUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, tenantString)
	var diags diag.Diagnostics
	var err error
	tenantUID, err := c.GetTenantUID()
	if err != nil {
		return diag.FromErr(err)
	}
	ssoType := d.Get("sso_auth_type").(string)
	switch ssoType {
	case "none":
		err = disableSSO(c, tenantUID)
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
		err = c.UpdateDomain(tenantUID, toDomains(v.(*schema.Set).List()))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if v, ok := d.GetOk("auth_providers"); ok {
		err = c.UpdateProviders(tenantUID, toAuthProviders(v.(*schema.Set).List()))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func resourceSSOCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	diags := resourceCommonUpdate(ctx, d, m)
	if !diags.HasError() {
		d.SetId("sso_settings")
	}

	return diags
}

func resourceSSORead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, tenantString)
	var diags diag.Diagnostics
	tenantUID, err := c.GetTenantUID()
	if err != nil {
		return diag.FromErr(err)
	}
	ssoType := d.Get("sso_auth_type").(string)
	if ssoType == "saml" {
		samlEntity, err := c.GetSAML(tenantUID)
		if err != nil {
			return diag.FromErr(err)
		}
		err = flattenSAML(samlEntity, d)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if ssoType == "oidc" {
		oidcEntity, err := c.GetOIDC(tenantUID)
		if err != nil {
			return diag.FromErr(err)
		}
		err = flattenOidc(oidcEntity, d)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if _, ok := d.GetOk("domains"); ok {
		domainsEntity, err := c.GetDomains(tenantUID)
		if err != nil {
			return diag.FromErr(err)
		}
		err = flattenDomains(domainsEntity, d)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if _, ok := d.GetOk("auth_providers"); ok {
		authEntity, err := c.GetProviders(tenantUID)
		if err != nil {
			return diag.FromErr(err)
		}
		err = flattenAuthProviders(authEntity, d)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func resourceSSOUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, tenantString)
	var diags diag.Diagnostics
	var err error
	tenantUID, err := c.GetTenantUID()
	if err != nil {
		return diag.FromErr(err)
	}
	diags = resourceCommonUpdate(ctx, d, m)
	if ok := d.HasChange("domains"); ok {
		v := d.Get("domains")
		err = c.UpdateDomain(tenantUID, toDomains(v.(*schema.Set).List()))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if ok := d.HasChange("auth_providers"); ok {
		v := d.Get("auth_providers")
		err = c.UpdateProviders(tenantUID, toAuthProviders(v.(*schema.Set).List()))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if !diags.HasError() {
		d.SetId("sso_settings")
	}
	return diags
}

func resourceSSODelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, tenantString)
	var diags diag.Diagnostics
	var err error
	tenantUID, err := c.GetTenantUID()
	if err != nil {
		return diag.FromErr(err)
	}
	err = disableSSO(c, tenantUID)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return diags
}

func customDiffValidation(ctx context.Context, d *schema.ResourceDiff, v interface{}) error {
	ssoAuthType, ok := d.GetOk("sso_auth_type")
	if !ok {
		return nil // No validation needed if not set
	}

	authType := ssoAuthType.(string)
	_, samlExists := d.GetOk("saml")
	_, oidcExists := d.GetOk("oidc")

	switch authType {
	case "none":
		if samlExists || oidcExists {
			return fmt.Errorf("sso_auth_type is set to 'none', so 'saml' and 'oidc' should not be defined")
		}
	case "saml":
		if oidcExists {
			return fmt.Errorf("sso_auth_type is set to 'saml', so 'oidc' should not be defined")
		}
		if !samlExists {
			return fmt.Errorf("sso_auth_type is set to 'saml', so 'saml' should be defined")
		}
	case "oidc":
		if samlExists {
			return fmt.Errorf("sso_auth_type is set to 'oidc', so 'saml' should not be defined")
		}
		if !oidcExists {
			return fmt.Errorf("sso_auth_type is set to 'oidc', so 'oidc' should be defined")
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

	oidc := d.Get("oidc").([]interface{})[0].(map[string]interface{})

	oidcSpec.CallbackURL = oidc["callback_url"].(string)
	oidcSpec.ClientID = oidc["client_id"].(string)
	oidcSpec.ClientSecret = oidc["client_secret"].(string)
	oidcSpec.DefaultTeams = toStringSlice(oidc["default_team_ids"].(*schema.Set).List())
	oidcSpec.IsSsoEnabled = true
	oidcSpec.IssuerTLS = &models.V1OidcIssuerTLS{
		CaCertificateBase64: oidc["identity_provider_ca_certificate"].(string),
		InsecureSkipVerify:  BoolPtr(oidc["insecure_skip_tls_verify"].(bool)),
	}
	oidcSpec.IssuerURL = oidc["issuer_url"].(string)
	oidcSpec.LogoutURL = oidc["logout_url"].(string)
	oidcSpec.RequiredClaims = &models.V1TenantOidcClaims{
		Email:       oidc["email"].(string),
		FirstName:   oidc["first_name"].(string),
		LastName:    oidc["last_name"].(string),
		SpectroTeam: oidc["spectro_team"].(string),
	}
	oidcSpec.Scopes = toStringSlice(oidc["scopes"].(*schema.Set).List())
	oidcSpec.ScopesDelimiter = ""
	oidcSpec.SyncSsoTeams = true

	if uie, ok := oidc["user_info_endpoint"]; ok {
		if len(uie.([]interface{})) > 0 {
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
	}
	return oidcSpec
}

func flattenOidc(oidcSpec *models.V1TenantOidcClientSpec, d *schema.ResourceData) error {
	var err error
	var oidc []interface{}
	spec := make(map[string]interface{})

	spec["callback_url"] = oidcSpec.CallbackURL
	spec["client_id"] = oidcSpec.ClientID
	spec["client_secret"] = oidcSpec.ClientSecret

	spec["default_team_ids"] = oidcSpec.DefaultTeams
	decodeCA, _ := base64.StdEncoding.DecodeString(oidcSpec.IssuerTLS.CaCertificateBase64)
	spec["identity_provider_ca_certificate"] = string(decodeCA)
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
	saml := d.Get("saml").([]interface{})[0].(map[string]interface{})
	// Attributes
	firstNameTemplate := &models.V1TenantSamlSpecAttribute{
		AttributeValue:  "[FIRST_NAME_OF_USER]",
		MappedAttribute: saml["first_name"].(string),
		Name:            "FirstName",
		NameFormat:      "urn:oasis:names:tc:SAML:2.0:attrname-format:unspecified",
	}
	lastNameTemplate := &models.V1TenantSamlSpecAttribute{
		AttributeValue:  "[LAST_NAME_OF_USER]",
		MappedAttribute: saml["last_name"].(string),
		Name:            "LastName",
		NameFormat:      "urn:oasis:names:tc:SAML:2.0:attrname-format:unspecified",
	}
	emailTemplate := &models.V1TenantSamlSpecAttribute{
		AttributeValue:  "[EMAIL_OF_USER]",
		MappedAttribute: saml["email"].(string),
		Name:            "Email",
		NameFormat:      "urn:oasis:names:tc:SAML:2.0:attrname-format:unspecified",
	}
	spectroTeamTemplate := &models.V1TenantSamlSpecAttribute{
		AttributeValue:  "[SPECTRO_TEAM_OF_USER]",
		MappedAttribute: saml["spectro_team"].(string),
		Name:            "SpectroTeam",
		NameFormat:      "urn:oasis:names:tc:SAML:2.0:attrname-format:unspecified",
	}

	attributes := make([]*models.V1TenantSamlSpecAttribute, 0)
	attributes = append(attributes, firstNameTemplate)
	attributes = append(attributes, lastNameTemplate)
	attributes = append(attributes, emailTemplate)
	attributes = append(attributes, spectroTeamTemplate)
	samlSpec.Attributes = attributes

	samlSpec.DefaultTeams = toStringSlice(saml["default_team_ids"].(*schema.Set).List())
	samlSpec.FederationMetadata = base64.StdEncoding.EncodeToString([]byte(saml["identity_provider_metadata"].(string)))
	samlSpec.IdentityProvider = saml["service_provider"].(string)
	samlSpec.IsSingleLogoutEnabled = saml["enable_single_logout"].(bool)
	samlSpec.IsSsoEnabled = true
	samlSpec.NameIDFormat = saml["name_id_format"].(string)
	samlSpec.SyncSsoTeams = true

	return samlSpec
}

func flattenSAML(saml *models.V1TenantSamlSpec, d *schema.ResourceData) error {
	var err error
	var samlData []interface{}
	spec := make(map[string]interface{})
	spec["issuer"] = saml.Issuer
	spec["certificate"] = saml.Certificate
	spec["service_provider"] = saml.IdentityProvider
	decodeCA, _ := base64.StdEncoding.DecodeString(saml.FederationMetadata)
	spec["identity_provider_metadata"] = string(decodeCA)
	spec["default_team_ids"] = saml.DefaultTeams
	spec["enable_single_logout"] = saml.IsSingleLogoutEnabled
	spec["single_logout_url"] = saml.SingleLogoutURL
	spec["entity_id"] = saml.EntityID
	spec["name_id_format"] = saml.NameIDFormat
	spec["login_url"] = saml.AudienceURL
	spec["service_provider_metadata"] = saml.ServiceProviderMetadata

	attribute := make(map[string]interface{})

	for _, a := range saml.Attributes {
		attribute[a.Name] = a
	}
	spec["first_name"] = attribute["FirstName"].(*models.V1TenantSamlSpecAttribute).MappedAttribute
	spec["last_name"] = attribute["LastName"].(*models.V1TenantSamlSpecAttribute).MappedAttribute
	spec["email"] = attribute["Email"].(*models.V1TenantSamlSpecAttribute).MappedAttribute
	spec["spectro_team"] = attribute["SpectroTeam"].(*models.V1TenantSamlSpecAttribute).MappedAttribute
	samlData = append(samlData, spec)
	if err = d.Set("saml", samlData); err != nil {
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
	if len(providers) == 0 {
		authProviderSpec := &models.V1TenantSsoAuthProvidersEntity{
			IsEnabled: false,
			SsoLogins: []string{""},
		}
		return authProviderSpec
	}
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

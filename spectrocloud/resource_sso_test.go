package spectrocloud

import (
	"encoding/base64"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestToStringSlice(t *testing.T) {
	cases := []struct {
		name   string
		input  []interface{}
		expect []string
	}{
		{
			name:   "empty slice",
			input:  []interface{}{},
			expect: []string{},
		},
		{
			name:   "single element",
			input:  []interface{}{"one"},
			expect: []string{"one"},
		},
		{
			name:   "multiple elements",
			input:  []interface{}{"one", "two", "three"},
			expect: []string{"one", "two", "three"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := toStringSlice(tc.input)
			assert.Equal(t, tc.expect, result)
		})
	}
}

func TestToOIDC(t *testing.T) {
	input := map[string]interface{}{
		"oidc": []interface{}{
			map[string]interface{}{
				"callback_url":                     "https://example.com/callback",
				"client_id":                        "client-id",
				"client_secret":                    "client-secret",
				"default_team_ids":                 schema.NewSet(schema.HashString, []interface{}{"team1", "team2"}),
				"identity_provider_ca_certificate": "cert",
				"insecure_skip_tls_verify":         true,
				"issuer_url":                       "https://issuer.com",
				"logout_url":                       "https://example.com/logout",
				"email":                            "email",
				"first_name":                       "given_name",
				"last_name":                        "family_name",
				"spectro_team":                     "groups",
				"scopes":                           schema.NewSet(schema.HashString, []interface{}{"openid", "profile"}),
				"user_info_endpoint": []interface{}{
					map[string]interface{}{
						"email":        "email",
						"first_name":   "given_name",
						"last_name":    "family_name",
						"spectro_team": "groups",
					},
				},
			},
		},
	}

	resourceData := resourceSSO().TestResourceData()
	err := resourceData.Set("oidc", input["oidc"])
	if err != nil {
		return
	}
	result := toOIDC(resourceData)

	assert.Equal(t, "https://example.com/callback", result.CallbackURL)
	assert.Equal(t, "client-id", result.ClientID)
	assert.Equal(t, "client-secret", result.ClientSecret)
	assert.Equal(t, []string{"team1", "team2"}, result.DefaultTeams)
	assert.Equal(t, "cert", result.IssuerTLS.CaCertificateBase64)
	assert.Equal(t, true, *result.IssuerTLS.InsecureSkipVerify)
	assert.Equal(t, "https://issuer.com", result.IssuerURL)
	assert.Equal(t, "https://example.com/logout", result.LogoutURL)
	assert.Equal(t, "email", result.RequiredClaims.Email)
	assert.Equal(t, "given_name", result.RequiredClaims.FirstName)
	assert.Equal(t, "family_name", result.RequiredClaims.LastName)
	assert.Equal(t, "groups", result.RequiredClaims.SpectroTeam)
	assert.Equal(t, "email", result.UserInfo.Claims.Email)
	assert.Equal(t, "given_name", result.UserInfo.Claims.FirstName)
	assert.Equal(t, "family_name", result.UserInfo.Claims.LastName)
	assert.Equal(t, "groups", result.UserInfo.Claims.SpectroTeam)
	assert.Equal(t, true, *result.UserInfo.UseUserInfo)
}

func TestFlattenOidc(t *testing.T) {
	encodedCA := base64.StdEncoding.EncodeToString([]byte("test-ca-cert"))
	resourceSchema := map[string]*schema.Schema{
		"oidc": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"callback_url":                     {Type: schema.TypeString, Required: true},
					"client_id":                        {Type: schema.TypeString, Required: true},
					"client_secret":                    {Type: schema.TypeString, Required: true},
					"default_team_ids":                 {Type: schema.TypeList, Elem: &schema.Schema{Type: schema.TypeString}, Required: true},
					"identity_provider_ca_certificate": {Type: schema.TypeString, Required: true},
					"insecure_skip_tls_verify":         {Type: schema.TypeBool, Required: true},
					"issuer_url":                       {Type: schema.TypeString, Required: true},
					"logout_url":                       {Type: schema.TypeString, Required: true},
					"email":                            {Type: schema.TypeString, Required: true},
					"first_name":                       {Type: schema.TypeString, Required: true},
					"last_name":                        {Type: schema.TypeString, Required: true},
					"spectro_team":                     {Type: schema.TypeString, Required: true},
					"scopes":                           {Type: schema.TypeList, Elem: &schema.Schema{Type: schema.TypeString}, Required: true},
					"user_info_endpoint": {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"email":        {Type: schema.TypeString, Required: true},
								"first_name":   {Type: schema.TypeString, Required: true},
								"last_name":    {Type: schema.TypeString, Required: true},
								"spectro_team": {Type: schema.TypeString, Required: true},
							},
						},
					},
				},
			},
		},
	}

	resourceData := schema.TestResourceDataRaw(t, resourceSchema, map[string]interface{}{})

	oidcSpec := &models.V1TenantOidcClientSpec{
		CallbackURL:  "https://example.com/callback",
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		DefaultTeams: []string{"team1", "team2"},
		IssuerTLS: &models.V1OidcIssuerTLS{
			CaCertificateBase64: encodedCA,
			InsecureSkipVerify:  BoolPtr(true),
		},
		IssuerURL: "https://issuer.com",
		LogoutURL: "https://example.com/logout",
		RequiredClaims: &models.V1TenantOidcClaims{
			Email:       "email",
			FirstName:   "given_name",
			LastName:    "family_name",
			SpectroTeam: "groups",
		},
		Scopes: []string{"openid", "profile"},
		UserInfo: &models.V1OidcUserInfo{
			Claims: &models.V1TenantOidcClaims{
				Email:       "email",
				FirstName:   "given_name",
				LastName:    "family_name",
				SpectroTeam: "groups",
			},
		},
	}

	err := flattenOidc(oidcSpec, resourceData)
	assert.NoError(t, err)

	flattened := resourceData.Get("oidc").([]interface{})[0].(map[string]interface{})
	assert.Equal(t, "https://example.com/callback", flattened["callback_url"])
	assert.Equal(t, "client-id", flattened["client_id"])
	assert.Equal(t, "client-secret", flattened["client_secret"])
	assert.Equal(t, []interface{}{"team1", "team2"}, flattened["default_team_ids"])
	assert.Equal(t, "test-ca-cert", flattened["identity_provider_ca_certificate"])
	assert.Equal(t, true, flattened["insecure_skip_tls_verify"])
	assert.Equal(t, "https://issuer.com", flattened["issuer_url"])
	assert.Equal(t, "https://example.com/logout", flattened["logout_url"])
	assert.Equal(t, "email", flattened["email"])
	assert.Equal(t, "given_name", flattened["first_name"])
	assert.Equal(t, "family_name", flattened["last_name"])
	assert.Equal(t, "groups", flattened["spectro_team"])
	assert.Equal(t, []interface{}{"openid", "profile"}, flattened["scopes"])

	userInfo := flattened["user_info_endpoint"].([]interface{})[0].(map[string]interface{})
	assert.Equal(t, "email", userInfo["email"])
	assert.Equal(t, "given_name", userInfo["first_name"])
	assert.Equal(t, "family_name", userInfo["last_name"])
	assert.Equal(t, "groups", userInfo["spectro_team"])
}

func TestToSAML(t *testing.T) {
	saml := map[string]interface{}{
		"saml": []interface{}{
			map[string]interface{}{
				"first_name":                 "John",
				"last_name":                  "Doe",
				"email":                      "user@example.com",
				"spectro_team":               "devops",
				"default_team_ids":           schema.NewSet(schema.HashString, []interface{}{"team1", "team2"}),
				"identity_provider_metadata": "metadata-xml",
				"service_provider":           "https://sso.example.com",
				"enable_single_logout":       true,
				"name_id_format":             "urn:oasis:names:tc:SAML:2.0:nameid-format:emailAddress",
			},
		},
	}

	d := resourceSSO().TestResourceData()
	err := d.Set("saml", saml["saml"])
	if err != nil {
		return
	}

	samlSpec := toSAML(d)
	assert.Equal(t, "John", samlSpec.Attributes[0].MappedAttribute)
	assert.Equal(t, "Doe", samlSpec.Attributes[1].MappedAttribute)
	assert.Equal(t, "user@example.com", samlSpec.Attributes[2].MappedAttribute)
	assert.Equal(t, "devops", samlSpec.Attributes[3].MappedAttribute)
	assert.Equal(t, []string{"team1", "team2"}, samlSpec.DefaultTeams)
	assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("metadata-xml")), samlSpec.FederationMetadata)
	assert.Equal(t, "https://sso.example.com", samlSpec.IdentityProvider)
	assert.Equal(t, true, samlSpec.IsSingleLogoutEnabled)
	assert.Equal(t, "urn:oasis:names:tc:SAML:2.0:nameid-format:emailAddress", samlSpec.NameIDFormat)
}

func TestFlattenSAML(t *testing.T) {
	resourceData := resourceSSO().TestResourceData()

	samlSpec := &models.V1TenantSamlSpec{
		Issuer:                  "https://sso.example.com",
		Certificate:             "cert-data",
		IdentityProvider:        "idp-provider",
		FederationMetadata:      base64.StdEncoding.EncodeToString([]byte("metadata-xml")),
		DefaultTeams:            []string{"team1", "team2"},
		IsSingleLogoutEnabled:   true,
		SingleLogoutURL:         "https://logout.example.com",
		EntityID:                "entity-id",
		NameIDFormat:            "urn:oasis:names:tc:SAML:2.0:nameid-format:emailAddress",
		AudienceURL:             "https://login.example.com",
		ServiceProviderMetadata: "sp-metadata",
		Attributes: []*models.V1TenantSamlSpecAttribute{
			{Name: "FirstName", MappedAttribute: "John"},
			{Name: "LastName", MappedAttribute: "Doe"},
			{Name: "Email", MappedAttribute: "user@example.com"},
			{Name: "SpectroTeam", MappedAttribute: "devops"},
		},
	}

	err := flattenSAML(samlSpec, resourceData)
	assert.NoError(t, err)
	assert.Equal(t, "https://sso.example.com", resourceData.Get("saml.0.issuer"))
	assert.Equal(t, "cert-data", resourceData.Get("saml.0.certificate"))
	assert.Equal(t, "idp-provider", resourceData.Get("saml.0.service_provider"))
	assert.Equal(t, "metadata-xml", resourceData.Get("saml.0.identity_provider_metadata"))
	assert.Equal(t, true, resourceData.Get("saml.0.enable_single_logout"))
	assert.Equal(t, "https://logout.example.com", resourceData.Get("saml.0.single_logout_url"))
	assert.Equal(t, "entity-id", resourceData.Get("saml.0.entity_id"))
	assert.Equal(t, "urn:oasis:names:tc:SAML:2.0:nameid-format:emailAddress", resourceData.Get("saml.0.name_id_format"))
	assert.Equal(t, "https://login.example.com", resourceData.Get("saml.0.login_url"))
	assert.Equal(t, "sp-metadata", resourceData.Get("saml.0.service_provider_metadata"))
	assert.Equal(t, "John", resourceData.Get("saml.0.first_name"))
	assert.Equal(t, "Doe", resourceData.Get("saml.0.last_name"))
	assert.Equal(t, "user@example.com", resourceData.Get("saml.0.email"))
	assert.Equal(t, "devops", resourceData.Get("saml.0.spectro_team"))
}

func TestToDomains(t *testing.T) {
	input := []interface{}{"example.com", "test.com"}
	result := toDomains(input)
	assert.Equal(t, []string{"example.com", "test.com"}, result.Domains)
}

func TestFlattenDomains(t *testing.T) {
	resourceData := resourceSSO().TestResourceData()
	err := resourceData.Set("domains", []interface{}{})
	if err != nil {
		return
	}

	domainSpec := &models.V1TenantDomains{
		Domains: []string{"example.com", "test.com"},
	}
	err = flattenDomains(domainSpec, resourceData)
	assert.NoError(t, err)
}

func TestToAuthProviders(t *testing.T) {
	emptyProviders := toAuthProviders([]interface{}{})
	assert.False(t, emptyProviders.IsEnabled)
	assert.Equal(t, []string{""}, emptyProviders.SsoLogins)

	providers := []interface{}{"sso1", "sso2"}
	result := toAuthProviders(providers)
	assert.True(t, result.IsEnabled)
	assert.Equal(t, []string{"sso1", "sso2"}, result.SsoLogins)
}

func TestFlattenAuthProviders(t *testing.T) {
	resourceData := resourceSSO().TestResourceData()
	err := resourceData.Set("auth_providers", []interface{}{})
	authProviderSpec := &models.V1TenantSsoAuthProvidersEntity{
		SsoLogins: []string{"sso1", "sso2"},
	}
	err = flattenAuthProviders(authProviderSpec, resourceData)
	assert.NoError(t, err)
}

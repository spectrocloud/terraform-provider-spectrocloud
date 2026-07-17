package routes

import (
	"net/http"

	"github.com/spectrocloud/palette-sdk-go/api/models"
)

// SSORoutes wires the four SSO endpoint pairs (SAML/OIDC/domains/providers).
func SSORoutes() []Route {
	return []Route{
		// SAML
		{
			Method: "GET",
			Path:   "/v1/tenants/{tenantUid}/saml/config",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    &models.V1TenantSamlSpec{},
			},
		},
		{
			Method: "POST",
			Path:   "/v1/tenants/{tenantUid}/saml/config",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},

		// OIDC
		{
			Method: "GET",
			Path:   "/v1/tenants/{tenantUid}/oidc/config",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    &models.V1TenantOidcClientSpec{},
			},
		},
		{
			Method: "POST",
			Path:   "/v1/tenants/{tenantUid}/oidc/config",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},

		// Domains
		{
			Method: "GET",
			Path:   "/v1/tenants/{tenantUid}/domains",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload: &models.V1TenantDomains{
					Domains: []string{"example.com"},
				},
			},
		},
		{
			Method: "POST",
			Path:   "/v1/tenants/{tenantUid}/domains",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},

		// SSO auth providers
		{
			Method: "GET",
			Path:   "/v1/tenants/{tenantUid}/sso/auth/providers",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    &models.V1TenantSsoAuthProvidersEntity{},
			},
		},
		{
			Method: "POST",
			Path:   "/v1/tenants/{tenantUid}/sso/auth/providers",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
	}
}

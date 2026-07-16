package routes

import (
	"net/http"
	"strconv"

	"github.com/spectrocloud/palette-sdk-go/api/models"
)

// PlatformSettingRoutes wires the ~10 SDK endpoints that
// resource_platform_setting.go touches:
//   - /v1/tenants/{tenantUid}/authTokenSettings         (GET/PUT)
//   - /v1/tenants/{tenantUid}/preferences/clusterRbacSettings (GET/PUT)
//   - /v1/tenants/{tenantUid}/loginBanner               (GET/PUT)
//   - /v1/tenants/{tenantUid}/preferences/clusterSettings   (GET)
//   - /v1/tenants/{tenantUid}/preferences/clusterSettings/nodesAutoRemediationSetting (PUT)
//   - /v1/tenants/{tenantUid}/preferences/fips          (GET/PUT)
//   - /v1/projects/{uid}/preferences/clusterSettings    (GET)
//   - /v1/projects/{uid}/preferences/clusterSettings/nodesAutoRemediationSetting (PUT)
//   - /v1/spectroclusters/upgrade/settings              (GET/POST)
func PlatformSettingRoutes() []Route {
	enabled := "enabled"
	nonFipsDisabled := "nonFipsDisabled"

	return []Route{
		// Session timeout
		{
			Method: "GET",
			Path:   "/v1/tenants/{tenantUid}/authTokenSettings",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload: &models.V1AuthTokenSettings{
					ExpiryTimeMinutes: 240,
				},
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/tenants/{tenantUid}/authTokenSettings",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},

		// Cluster RBAC settings (automatic cluster role binding)
		{
			Method: "GET",
			Path:   "/v1/tenants/{tenantUid}/preferences/clusterRbacSettings",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload: &models.V1TenantClusterRbacSettings{
					AutomaticClusterRoleBinding: &enabled,
				},
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/tenants/{tenantUid}/preferences/clusterRbacSettings",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},

		// Login banner
		{
			Method: "GET",
			Path:   "/v1/tenants/{tenantUid}/loginBanner",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload: &models.V1LoginBannerSettings{
					Title:     "Welcome",
					Message:   "Login banner message",
					IsEnabled: true,
				},
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/tenants/{tenantUid}/loginBanner",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},

		// Tenant cluster auto-remediation (GET reads the whole cluster
		// settings block, which embeds nodesAutoRemediationSetting)
		{
			Method: "GET",
			Path:   "/v1/tenants/{tenantUid}/preferences/clusterSettings",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload: &models.V1TenantClusterSettings{
					NodesAutoRemediationSetting: &models.V1NodesAutoRemediationSettings{
						IsEnabled:                   true,
						DisableNodesAutoRemediation: false,
					},
				},
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/tenants/{tenantUid}/preferences/clusterSettings/nodesAutoRemediationSetting",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},

		// Project cluster auto-remediation
		{
			Method: "GET",
			Path:   "/v1/projects/{uid}/preferences/clusterSettings",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload: &models.V1ProjectClusterSettings{
					NodesAutoRemediationSetting: &models.V1NodesAutoRemediationSettings{
						IsEnabled:                   true,
						DisableNodesAutoRemediation: false,
					},
				},
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/projects/{uid}/preferences/clusterSettings/nodesAutoRemediationSetting",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},

		// FIPS
		{
			Method: "GET",
			Path:   "/v1/tenants/{tenantUid}/preferences/fips",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload: &models.V1FipsSettings{
					FipsPackConfig:           &models.V1NonFipsConfig{Mode: &nonFipsDisabled},
					FipsClusterFeatureConfig: &models.V1NonFipsConfig{Mode: &nonFipsDisabled},
					FipsClusterImportConfig:  &models.V1NonFipsConfig{Mode: &nonFipsDisabled},
				},
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/tenants/{tenantUid}/preferences/fips",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},

		// Platform upgrade settings — note the SDK uses POST for update.
		{
			Method: "GET",
			Path:   "/v1/spectroclusters/upgrade/settings",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload: &models.V1ClusterUpgradeSettingsEntity{
					SpectroComponents: "unlock",
				},
			},
		},
		{
			Method: "POST",
			Path:   "/v1/spectroclusters/upgrade/settings",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
	}
}

// PlatformSettingNegativeRoutes returns errors on the PUT paths. Reads
// stay success-shaped because the SDK dereferences resp.Payload before
// checking err (same nil-payload class of bug seen with
// GetPasswordPolicy / GetDeveloperSetting / GetResourceLimits).
func PlatformSettingNegativeRoutes() []Route {
	return []Route{
		{
			Method: "PUT",
			Path:   "/v1/tenants/{tenantUid}/authTokenSettings",
			Response: ResponseData{
				StatusCode: http.StatusBadRequest,
				Payload:    getError(strconv.Itoa(http.StatusBadRequest), "Invalid session timeout"),
			},
		},
	}
}

package routes

import (
	"net/http"
	"strconv"

	"github.com/spectrocloud/palette-sdk-go/api/models"
)

// Fixture values echoed by the positive developer-setting routes. The
// numbers match what prepareDeveloperSettingResourceData sets in
// resource_developer_setting_test.go so a full CRUD cycle Read matches
// its own writes.
func mockDeveloperSettingPayload() *models.V1DeveloperCredit {
	return &models.V1DeveloperCredit{
		CPU:                  12,
		MemoryGiB:            16,
		StorageGiB:           20,
		VirtualClustersLimit: 2,
	}
}

func mockClusterGroupPrefPayload() *models.V1TenantEnableClusterGroup {
	return &models.V1TenantEnableClusterGroup{
		HideSystemClusterGroups: false,
	}
}

// DeveloperSettingRoutes wires the developer-credit + cluster-group
// preference endpoints, both tenant-scoped and both singletons (Create/
// Update/Delete on the Terraform side all map to a PUT on the API).
func DeveloperSettingRoutes() []Route {
	return []Route{
		{
			Method: "PUT",
			Path:   "/v1/tenants/{tenantUid}/preferences/developerCredit",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
		{
			Method: "GET",
			Path:   "/v1/tenants/{tenantUid}/preferences/developerCredit",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    mockDeveloperSettingPayload(),
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/tenants/{tenantUid}/preferences/clusterGroup",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
		{
			Method: "GET",
			Path:   "/v1/tenants/{tenantUid}/preferences/clusterGroup",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    mockClusterGroupPrefPayload(),
			},
		},
	}
}

// DeveloperSettingNegativeRoutes returns error responses on the write
// paths. Read-path negatives are NOT included: palette-sdk-go's
// GetDeveloperSetting and GetSystemClusterGroupPreference both
// dereference resp.Payload before checking err, so a non-2xx GET
// panics with SIGSEGV — same pattern as GetPasswordPolicy. Rather than
// silently panic, we let those return their positive payload on the
// negative server and cover only the Update failures here.
func DeveloperSettingNegativeRoutes() []Route {
	return []Route{
		{
			Method: "PUT",
			Path:   "/v1/tenants/{tenantUid}/preferences/developerCredit",
			Response: ResponseData{
				StatusCode: http.StatusBadRequest,
				Payload:    getError(strconv.Itoa(http.StatusBadRequest), "Invalid developer credit"),
			},
		},
		{
			// Reads on the negative server: return success anyway. This is
			// deliberate — see comment above.
			Method: "GET",
			Path:   "/v1/tenants/{tenantUid}/preferences/developerCredit",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    mockDeveloperSettingPayload(),
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/tenants/{tenantUid}/preferences/clusterGroup",
			Response: ResponseData{
				StatusCode: http.StatusBadRequest,
				Payload:    getError(strconv.Itoa(http.StatusBadRequest), "Invalid cluster group preference"),
			},
		},
		{
			Method: "GET",
			Path:   "/v1/tenants/{tenantUid}/preferences/clusterGroup",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    mockClusterGroupPrefPayload(),
			},
		},
	}
}

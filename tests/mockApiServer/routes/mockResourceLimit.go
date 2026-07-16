package routes

import (
	"net/http"
	"strconv"

	"github.com/spectrocloud/palette-sdk-go/api/models"
)

// Small fixture — the resource itself defaults to ~22 kinds, but Read only
// cares about matching Kind → schema-field mapping, so a partial payload
// works. Anything not in this list falls back to the schema default.
func mockResourceLimitPayload() *models.V1TenantResourceLimits {
	return &models.V1TenantResourceLimits{
		Resources: []*models.V1TenantResourceLimit{
			{Kind: models.V1ResourceLimitTypeAlert, Limit: 100},
			{Kind: models.V1ResourceLimitTypeAPIKey, Limit: 20},
			{Kind: models.V1ResourceLimitTypeSpectrocluster, Limit: 10000},
		},
	}
}

// ResourceLimitRoutes wires the tenant resource-limit endpoints. Update
// uses PATCH (not PUT) — the API's contract, mirrored here so a client
// that hits POST/PUT gets the standard "no route" 404 rather than a
// spurious success.
func ResourceLimitRoutes() []Route {
	return []Route{
		{
			Method: "PATCH",
			Path:   "/v1/tenants/{tenantUid}/resourceLimits",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
		{
			Method: "GET",
			Path:   "/v1/tenants/{tenantUid}/resourceLimits",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    mockResourceLimitPayload(),
			},
		},
	}
}

// ResourceLimitNegativeRoutes returns errors on the PATCH path only.
// GET returns success on the negative server because the SDK's
// GetResourceLimits dereferences resp.Payload before checking err
// (same nil-payload bug pattern as GetPasswordPolicy /
// GetDeveloperSetting) — a real-error response would SIGSEGV inside
// the SDK, not the provider.
func ResourceLimitNegativeRoutes() []Route {
	return []Route{
		{
			Method: "PATCH",
			Path:   "/v1/tenants/{tenantUid}/resourceLimits",
			Response: ResponseData{
				StatusCode: http.StatusBadRequest,
				Payload:    getError(strconv.Itoa(http.StatusBadRequest), "Invalid resource limits"),
			},
		},
		{
			Method: "GET",
			Path:   "/v1/tenants/{tenantUid}/resourceLimits",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    mockResourceLimitPayload(),
			},
		},
	}
}

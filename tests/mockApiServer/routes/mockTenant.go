package routes

import (
	"net/http"
)

// Stable tenant UID for tenant-scoped macro and data source tests.
const mockTenantUID = "test-tenant-uid"

func getMockUserInfoPayload() map[string]interface{} {
	return map[string]interface{}{
		"orgName":   "Default",
		"tenantUid": mockTenantUID,
		"userUid":   "test-user-uid",
	}
}

func TenantRoutes() []Route {
	return []Route{
		{
			Method: "GET",
			Path:   "/v1/users/info",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    getMockUserInfoPayload(),
			},
		}}
}

func TenantNegativeRoutes() []Route {
	return []Route{
		{
			Method: "GET",
			Path:   "/v1/users/info",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    getMockUserInfoPayload(),
			},
		}}
}

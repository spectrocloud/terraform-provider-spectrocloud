package routes

import (
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"net/http"
)

func DeveloperSettingRoutes() []Route {
	return []Route{
		{
			Method: "GET",
			Path:   "/v1/tenants/{tenantUid}/preferences/developerCredit",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload: models.V1DeveloperCredit{
					CPU:                  12,
					MemoryGiB:            16,
					StorageGiB:           20,
					VirtualClustersLimit: 2,
				},
			},
		},
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
			Path:   "/v1/tenants/{tenantUid}/preferences/clusterGroup",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload: models.V1TenantEnableClusterGroup{
					HideSystemClusterGroups: false,
				},
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
	}
}

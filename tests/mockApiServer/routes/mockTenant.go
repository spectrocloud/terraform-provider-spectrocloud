package routes

import (
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"net/http"
)

func TenantRoutes() []Route {
	return []Route{
		{
			Method: "GET",
			Path:   "/v1/users/info",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload: models.V1UserInfo{
					OrgName:   "Default",
					TenantUID: generateRandomStringUID(),
					UserUID:   generateRandomStringUID(),
				},
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
				Payload: models.V1UserInfo{
					OrgName:   "Default",
					TenantUID: generateRandomStringUID(),
					UserUID:   generateRandomStringUID(),
				},
			},
		}}
}

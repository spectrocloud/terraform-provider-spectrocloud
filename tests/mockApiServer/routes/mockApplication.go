package routes

import "github.com/spectrocloud/palette-sdk-go/api/models"

func ApplicationRoutes() []Route {
	return []Route{
		{
			Method: "POST",
			Path:   "/v1/dashboard/appDeployments",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1AppDeploymentsSummary{
					AppDeployments: []*models.V1AppDeploymentSummary{
						{
							Metadata: &models.V1ObjectMeta{
								Name: "test-app-deployment",
								UID:  "test-app-id",
							},
						},
					},
				},
			},
		},
	}
}

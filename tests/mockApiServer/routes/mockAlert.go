package routes

import "github.com/spectrocloud/palette-sdk-go/api/models"

func AlertRoutes() []Route {
	return []Route{
		{
			Method: "PUT",
			Path:   "/v1/projects/{uid}/alerts/{component}/{alertUid}",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/projects/{uid}/alerts/{component}/{alertUid}",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "POST",
			Path:   "/v1/projects/{uid}/alerts/{component}",
			Response: ResponseData{
				StatusCode: 201,
				Payload:    map[string]string{"UID": "test-alert-1"},
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/projects/{uid}/alerts/{component}",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "GET",
			Path:   "/v1/projects/{uid}/alerts/{component}/{alertUid}",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1Channel{
					AlertAllUsers: false,
					CreatedBy:     "test-user",
					HTTP: &models.V1ChannelHTTP{
						Body: "test body",
						Headers: map[string]string{
							"test": "test",
						},
						Method: "PUT",
						URL:    "test.com",
					},
					Identifiers: []string{"test1"},
					IsActive:    false,
					Status: &models.V1AlertNotificationStatus{
						IsSucceeded: false,
						Message:     "test message",
						Time:        models.V1Time{},
					},
					Type: "test-type",
					UID:  "test-uid",
				},
			},
		},
	}
}

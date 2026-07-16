package routes

import (
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"net/http"
	"strconv"
)

func getUsersResponse() models.V1Users {
	return models.V1Users{
		Items: []*models.V1User{
			{
				Metadata: &models.V1ObjectMeta{
					Annotations:           nil,
					CreationTimestamp:     models.V1Time{},
					DeletionTimestamp:     models.V1Time{},
					Labels:                nil,
					LastModifiedTimestamp: models.V1Time{},
					Name:                  "test",
					UID:                   "12345",
				},
				Spec: &models.V1UserSpec{
					EmailID:   "test@spectrocloud.com",
					FirstName: "test",
					LastName:  "spectro",
					Roles:     nil,
				},
				Status: &models.V1UserStatus{
					ActivationLink:      "",
					IsActive:            true,
					IsPasswordResetting: false,
					LastSignIn:          models.V1Time{},
				},
			},
			{
				Metadata: &models.V1ObjectMeta{
					Annotations:           nil,
					CreationTimestamp:     models.V1Time{},
					DeletionTimestamp:     models.V1Time{},
					Labels:                nil,
					LastModifiedTimestamp: models.V1Time{},
					Name:                  "test-user2",
					UID:                   "test-user-12345",
				},
				Spec: &models.V1UserSpec{
					EmailID:   "test-user2@spectrocloud.com",
					FirstName: "test-user2",
					LastName:  "spectro",
					Roles:     nil,
				},
				Status: &models.V1UserStatus{
					ActivationLink:      "",
					IsActive:            true,
					IsPasswordResetting: false,
					LastSignIn:          models.V1Time{},
				},
			},
		},
		Listmeta: &models.V1ListMetaData{
			Continue: "",
			Count:    2,
			Limit:    10,
			Offset:   0,
		},
	}
}

func UserRoutes() []Route {
	return []Route{
		// Must be registered before /v1/users/{uid} so "info" is not captured as a UID.
		{
			Method: "GET",
			Path:   "/v1/users/info",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    getMockUserInfoPayload(),
			},
		},
		{
			Method: "GET",
			Path:   "/v1/users",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    getUsersResponse(),
			},
		},
		{
			Method: "GET",
			Path:   "/v1/users/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload: &models.V1User{
					Metadata: &models.V1ObjectMeta{
						Name: "test",
						UID:  "12345",
					},
					Spec: &models.V1UserSpec{
						EmailID:   "test@spectrocloud.com",
						FirstName: "test",
						LastName:  "spectro",
					},
				},
			},
		},
	}
}

func UserNegativeRoutes() []Route {
	return []Route{
		{
			Method: "GET",
			Path:   "/v1/users",
			Response: ResponseData{
				StatusCode: http.StatusNotFound,
				Payload:    getError(strconv.Itoa(http.StatusOK), "User not found"),
			},
		},
	}
}

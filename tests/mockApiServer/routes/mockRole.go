package routes

import (
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"net/http"
	"strconv"
)

func getRolesList() *models.V1Roles {
	return &models.V1Roles{
		Items: []*models.V1Role{
			{
				Metadata: &models.V1ObjectMeta{
					Annotations:           nil,
					CreationTimestamp:     models.V1Time{},
					DeletionTimestamp:     models.V1Time{},
					Labels:                nil,
					LastModifiedTimestamp: models.V1Time{},
					Name:                  "test-role",
					UID:                   generateRandomStringUID(),
				},
				Spec: &models.V1RoleSpec{
					Permissions: []string{"perm1", "perm2"},
					Scope:       "project",
					Type:        "",
				},
				Status: &models.V1RoleStatus{
					IsEnabled: true,
				},
			},
		},
		Listmeta: &models.V1ListMetaData{
			Continue: "",
			Count:    0,
			Limit:    0,
			Offset:   0,
		},
	}
}

func RolesRoutes() []Route {
	return []Route{
		{
			Method: "GET",
			Path:   "/v1/roles",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    getRolesList(),
			},
		},
	}
}

func RolesNegativeRoutes() []Route {
	return []Route{
		{
			Method: "GET",
			Path:   "/v1/roles",
			Response: ResponseData{
				StatusCode: http.StatusNotFound,
				Payload:    getError(strconv.Itoa(http.StatusNotFound), "No roles are found"),
			},
		},
	}
}

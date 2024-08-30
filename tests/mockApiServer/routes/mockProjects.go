package routes

import (
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"net/http"
	"strconv"
)

func getMockProjectPayload() models.V1Project {
	return models.V1Project{
		Metadata: &models.V1ObjectMeta{
			Annotations:       nil,
			CreationTimestamp: models.V1Time{},
			DeletionTimestamp: models.V1Time{},
			Labels: map[string]string{
				"description": "default project",
			},
			LastModifiedTimestamp: models.V1Time{},
			Name:                  "Default",
			UID:                   generateRandomStringUID(),
		},
		Spec: &models.V1ProjectSpec{
			Alerts:  nil,
			LogoURL: "",
			Teams:   nil,
			Users:   nil,
		},
		Status: &models.V1ProjectStatus{
			CleanUpStatus: nil,
			IsDisabled:    false,
		},
	}

}

func ProjectRoutes() []Route {
	return []Route{
		{
			Method: "POST",
			Path:   "/v1/projects",
			Response: ResponseData{
				StatusCode: http.StatusCreated,
				Payload:    map[string]interface{}{"UID": generateRandomStringUID()},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/projects/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    getMockProjectPayload(),
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/projects/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    map[string]interface{}{"UID": generateRandomStringUID()},
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/projects/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
		{
			Method: "GET",
			Path:   "/v1/dashboard/projects/metadata",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload: &models.V1ProjectsMetadata{
					Items: []*models.V1ProjectMetadata{
						{
							Metadata: &models.V1ObjectEntity{
								Name: "Default",
								UID:  generateRandomStringUID(),
							},
						},
					},
				},
			},
		},
	}
}

func ProjectNegativeRoutes() []Route {
	return []Route{
		{
			Method: "POST",
			Path:   "/v1/projects",
			Response: ResponseData{
				StatusCode: http.StatusConflict,
				Payload:    getError(strconv.Itoa(http.StatusConflict), "Project already exist"),
			},
		},
		{
			Method: "GET",
			Path:   "/v1/projects/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusNotFound,
				Payload:    getError(strconv.Itoa(http.StatusNotFound), "Project not found"),
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/projects/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusMethodNotAllowed,
				Payload:    getError(strconv.Itoa(http.StatusMethodNotAllowed), "Operation not allowed"),
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/projects/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusNotFound,
				Payload:    getError(strconv.Itoa(http.StatusNotFound), "Project not found"),
			},
		},
		{
			Method: "GET",
			Path:   "/v1/dashboard/projects/metadata",
			Response: ResponseData{
				StatusCode: http.StatusNotFound,
				Payload:    getError(strconv.Itoa(http.StatusNotFound), "Project metadata not found"),
			},
		},
	}
}

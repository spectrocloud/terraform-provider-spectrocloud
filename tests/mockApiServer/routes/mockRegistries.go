package routes

import (
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"net/http"
)

func getHelmRegistryPayload() *models.V1HelmRegistry {
	return &models.V1HelmRegistry{
		APIVersion: "",
		Kind:       "",
		Metadata: &models.V1ObjectMeta{
			Annotations:           nil,
			CreationTimestamp:     models.V1Time{},
			DeletionTimestamp:     models.V1Time{},
			Labels:                nil,
			LastModifiedTimestamp: models.V1Time{},
			Name:                  "Public",
			UID:                   generateRandomStringUID(),
		},
		Spec: &models.V1HelmRegistrySpec{
			Auth:        nil,
			Endpoint:    nil,
			IsPrivate:   false,
			Name:        "Public",
			RegistryUID: generateRandomStringUID(),
			Scope:       "project",
		},
		Status: &models.V1HelmRegistryStatus{
			HelmSyncStatus: &models.V1RegistrySyncStatus{
				LastRunTime:    models.V1Time{},
				LastSyncedTime: models.V1Time{},
				Message:        "",
				Status:         "Active",
			},
		},
	}
}

func RegistriesRoutes() []Route {
	return []Route{
		{
			Method: "GET",
			Path:   "/v1/registries/oci/summary",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload: &models.V1OciRegistries{
					Items: []*models.V1OciRegistry{
						{
							Metadata: &models.V1ObjectMeta{
								Name: "test-registry-oci",
								UID:  generateRandomStringUID(),
							},
							Spec:   nil,
							Status: nil,
						},
					},
				},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/registries/helm",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload: &models.V1HelmRegistries{
					Items: []*models.V1HelmRegistry{getHelmRegistryPayload()},
				},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/registries/helm/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    getHelmRegistryPayload(),
			},
		},
		{
			Method: "GET",
			Path:   "/v1/registries/metadata",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1RegistriesMetadata{
					Items: []*models.V1RegistryMetadata{
						{
							IsDefault: false,
							IsPrivate: false,
							Kind:      "",
							Name:      "test-registry-name",
							Scope:     "project",
							UID:       "test-registry-uid",
						},
					},
				},
			},
		},
	}
}

func RegistriesNegativeRoutes() []Route {
	return []Route{
		{
			Method: "GET",
			Path:   "/v1/registries/helm/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    getHelmRegistryPayload(),
				//StatusCode: http.StatusNotFound,
				//Payload:    getError(strconv.Itoa(http.StatusConflict), "Registry not found"),
			},
		},
	}
}

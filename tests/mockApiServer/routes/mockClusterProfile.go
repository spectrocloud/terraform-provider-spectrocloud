package routes

import (
	"net/http"

	"github.com/spectrocloud/palette-sdk-go/api/models"

	"github.com/spectrocloud/terraform-provider-spectrocloud/util/ptr"
)

func getClusterProfilesMetadataResponse() *models.V1ClusterProfilesMetadata {
	return &models.V1ClusterProfilesMetadata{
		Items: []*models.V1ClusterProfileMetadata{
			{
				Metadata: &models.V1ObjectEntity{
					Name: "test-cluster-profile-1",
					UID:  generateRandomStringUID(),
				},
				Spec: &models.V1ClusterProfileMetadataSpec{
					CloudType: "aws",
					Version:   "1.0.0",
				},
			},
			{
				Metadata: &models.V1ObjectEntity{
					Name: "test-cluster-profile-2",
					UID:  generateRandomStringUID(),
				},
				Spec: &models.V1ClusterProfileMetadataSpec{
					CloudType: "gcp",
					Version:   "1.0.0",
				},
			},
		},
	}
}

func getClusterProfileResponse() *models.V1ClusterProfile {
	return &models.V1ClusterProfile{
		APIVersion: "",
		Kind:       "",
		Metadata: &models.V1ObjectMeta{
			Annotations: map[string]string{
				"scope": "project",
			},
			CreationTimestamp:     models.V1Time{},
			DeletionTimestamp:     models.V1Time{},
			Labels:                nil,
			LastModifiedTimestamp: models.V1Time{},
			Name:                  "test-cluster-profile-1",
			UID:                   generateRandomStringUID(),
		},
		Spec: &models.V1ClusterProfileSpec{
			Draft: nil,
			Published: &models.V1ClusterProfileTemplate{
				CloudType:        "aws",
				Name:             "test-cluster-profile-1",
				PackServerRefs:   nil,
				PackServerSecret: "",
				Packs: []*models.V1PackRef{
					{
						Name:        ptr.To("k8"),
						PackUID:     generateRandomStringUID(),
						RegistryUID: generateRandomStringUID(),
						Schema:      nil,
						Values:      "{test-json:test}",
						Version:     "1.0.0",
					},
				},
				ProfileVersion: "1.0.0",
				RelatedObject:  nil,
				Type:           "cluster",
				UID:            generateRandomStringUID(),
				Version:        0,
			},
			Version:  "1.0.0",
			Versions: nil,
		},
		Status: &models.V1ClusterProfileStatus{
			HasUserMacros: false,
			InUseClusters: nil,
			IsPublished:   true,
		},
	}
}

func getClusterProfilePackManifestResponse() *models.V1ManifestEntities {
	return &models.V1ManifestEntities{
		Items: []*models.V1ManifestEntity{
			{
				Metadata: &models.V1ObjectMeta{
					Annotations:           nil,
					CreationTimestamp:     models.V1Time{},
					DeletionTimestamp:     models.V1Time{},
					Labels:                nil,
					LastModifiedTimestamp: models.V1Time{},
					Name:                  "test-manifest-1",
					UID:                   generateRandomStringUID(),
				},
				Spec: &models.V1ManifestSpec{
					Draft: &models.V1ManifestData{
						Content: "test-content",
						Digest:  "test-digest",
					},
					Published: &models.V1ManifestData{
						Content: "test-content",
						Digest:  "test-digest",
					},
				},
			},
		},
	}
}

func ClusterProfileRoutes() []Route {
	return []Route{
		{
			Method: "POST",
			Path:   "/v1/clusterprofiles/import/file",
			Response: ResponseData{
				StatusCode: 201,
				Payload:    map[string]string{"UID": "cluster-profile-import-1"},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/clusterprofiles/{uid}/variables",
			Response: ResponseData{
				StatusCode: 200,
				Payload:    &models.V1Variables{},
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/clusterprofiles/{uid}",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "POST",
			Path:   "/v1/clusterprofiles",
			Response: ResponseData{
				StatusCode: 201,
				Payload:    map[string]string{"UID": "cluster-profile-1"},
			},
		},
		{
			Method: "PATCH",
			Path:   "/v1/clusterprofiles/{uid}/publish",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "GET",
			Path:   "/v1/dashboard/clusterprofiles/metadata",
			Response: ResponseData{
				StatusCode: 200,
				Payload:    getClusterProfilesMetadataResponse(),
			},
		},
		{
			Method: "GET",
			Path:   "/v1/clusterprofiles/{uid}",
			Response: ResponseData{
				StatusCode: 200,
				Payload:    getClusterProfileResponse(),
			},
		},
		{
			Method: "GET",
			Path:   "/v1/clusterprofiles/{uid}/packs/{packName}/manifests",
			Response: ResponseData{
				StatusCode: 200,
				Payload:    getClusterProfilePackManifestResponse(),
			},
		},
	}
}

func ClusterProfileNegativeRoutes() []Route {
	return []Route{
		{
			Method: "GET",
			Path:   "/v1/dashboard/clusterprofiles/metadata",
			Response: ResponseData{
				StatusCode: 200,
				Payload:    &models.V1ClusterProfilesMetadata{},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/clusterprofiles/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusLocked,
				Payload:    nil,
			},
		},
	}
}

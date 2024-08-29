package routes

import (
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"net/http"
)

func getPackSummaryPayload() *models.V1PackSummaries {
	return &models.V1PackSummaries{
		Items: []*models.V1PackSummary{
			{
				APIVersion: "",
				Kind:       "",
				Metadata: &models.V1ObjectMeta{
					Annotations:           nil,
					CreationTimestamp:     models.V1Time{},
					DeletionTimestamp:     models.V1Time{},
					Labels:                nil,
					LastModifiedTimestamp: models.V1Time{},
					Name:                  "k8",
					UID:                   generateRandomStringUID(),
				},
				Spec: &models.V1PackSummarySpec{
					CloudTypes:  []string{"aws"},
					AddonType:   "infra",
					Name:        "k8",
					RegistryUID: "test-reg-uid",
					Type:        "helm",
					Values:      "test-test",
					Version:     "1.0",
				},
				Status: nil,
			},
			//{
			//	APIVersion: "",
			//	Kind:       "",
			//	Metadata: &models.V1ObjectMeta{
			//		Annotations:           nil,
			//		CreationTimestamp:     models.V1Time{},
			//		DeletionTimestamp:     models.V1Time{},
			//		Labels:                nil,
			//		LastModifiedTimestamp: models.V1Time{},
			//		Name:                  "cni",
			//		UID:                   generateRandomStringUID(),
			//	},
			//	Spec: &models.V1PackSummarySpec{
			//		CloudTypes:  []string{"aws"},
			//		AddonType:   "infra",
			//		Name:        "cni",
			//		RegistryUID: "test-reg-uid",
			//		Type:        "helm",
			//		Values:      "test-test",
			//		Version:     "1.0",
			//	},
			//	Status: nil,
			//},
		},
		Listmeta: nil,
	}
}

func getPackSummaryPayloadWithMultiPacks() *models.V1PackSummaries {
	return &models.V1PackSummaries{
		Items: []*models.V1PackSummary{
			{
				APIVersion: "",
				Kind:       "",
				Metadata: &models.V1ObjectMeta{
					Annotations:           nil,
					CreationTimestamp:     models.V1Time{},
					DeletionTimestamp:     models.V1Time{},
					Labels:                nil,
					LastModifiedTimestamp: models.V1Time{},
					Name:                  "k8",
					UID:                   generateRandomStringUID(),
				},
				Spec: &models.V1PackSummarySpec{
					CloudTypes:  []string{"aws"},
					AddonType:   "infra",
					Name:        "k8",
					RegistryUID: "test-reg-uid",
					Type:        "helm",
					Values:      "test-test",
					Version:     "1.0",
				},
				Status: nil,
			},
			{
				APIVersion: "",
				Kind:       "",
				Metadata: &models.V1ObjectMeta{
					Annotations:           nil,
					CreationTimestamp:     models.V1Time{},
					DeletionTimestamp:     models.V1Time{},
					Labels:                nil,
					LastModifiedTimestamp: models.V1Time{},
					Name:                  "cni",
					UID:                   generateRandomStringUID(),
				},
				Spec: &models.V1PackSummarySpec{
					CloudTypes:  []string{"aws"},
					AddonType:   "infra",
					Name:        "cni",
					RegistryUID: "test-reg-uid",
					Type:        "helm",
					Values:      "test-test",
					Version:     "1.0",
				},
				Status: nil,
			},
		},
		Listmeta: nil,
	}
}

func getPacksNameRegistryUIDNegative() *models.V1PackTagEntity {
	return &models.V1PackTagEntity{
		AddonSubType: "",
		AddonType:    "infra",
		CloudTypes:   []string{"aws", "eks"},
		DisplayName:  "k8",
		Layer:        "",
		LogoURL:      "",
		Name:         "k8",
		PackValues: []*models.V1PackUIDValues{
			{
				Annotations:  nil,
				Dependencies: nil,
				PackUID:      generateRandomStringUID(),
				Presets:      nil,
				Readme:       "",
				Schema:       nil,
				Template:     nil,
				Values:       "test-test",
			},
		},
		RegistryUID: generateRandomStringUID(),
		Tags: []*models.V1PackTags{
			{
				Group:      "dev",
				PackUID:    generateRandomStringUID(),
				ParentTags: nil,
				Tag:        "unit-test",
				Version:    "1.0",
			},
		},
	}
}

func getPacksNameRegistryUID() *models.V1PackTagEntity {
	return &models.V1PackTagEntity{
		AddonSubType: "",
		AddonType:    "infra",
		CloudTypes:   []string{"aws", "eks"},
		DisplayName:  "k8",
		Layer:        "",
		LogoURL:      "",
		Name:         "k8",
		PackValues: []*models.V1PackUIDValues{
			{
				Annotations:  nil,
				Dependencies: nil,
				PackUID:      "test-pack-uid",
				Presets:      nil,
				Readme:       "",
				Schema:       nil,
				Template:     nil,
				Values:       "test-test",
			},
		},
		RegistryUID: generateRandomStringUID(),
		Tags: []*models.V1PackTags{
			{
				Group:      "dev",
				PackUID:    "test-pack-uid",
				ParentTags: nil,
				Tag:        "unit-test",
				Version:    "1.0",
			},
		},
	}
}

func PacksRoutes() []Route {
	return []Route{
		{
			Method: "GET",
			Path:   "/v1/packs",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    getPackSummaryPayload(),
			},
		},
		{
			Method: "GET",
			Path:   "/v1/packs/{packName}/registries/{registryUid}",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    getPacksNameRegistryUID(),
			},
		},
	}
}

func PacksNegativeRoutes() []Route {
	return []Route{
		{
			Method: "GET",
			Path:   "/v1/packs",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    getPackSummaryPayloadWithMultiPacks(),
			},
		},
		{
			Method: "GET",
			Path:   "/v1/packs/{packName}/registries/{registryUid}",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    getPacksNameRegistryUIDNegative(),
			},
		},
	}
}

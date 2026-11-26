package routes

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/spectrocloud/palette-sdk-go/api/models"
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
					// Palette reports the pack tier as the "source" annotation (PLT-2544).
					Annotations: map[string]string{"source": "partner"},
					CloudTypes:  []string{"aws"},
					AddonType:   "infra",
					Name:        "k8",
					RegistryUID: "test-reg-uid",
					Type:        models.V1PackTypeHelm.Pointer(),
					Values:      "test-test",
					Version:     "1.0",
				},
				Status: nil,
			},
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
					Type:        models.V1PackTypeHelm.Pointer(),
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
					Type:        models.V1PackTypeHelm.Pointer(),
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
			Method:  "GET",
			Path:    "/v1/packs/{packName}/registries/{registryUid}",
			Handler: packsNameRegistryUIDHandler,
		},
	}
}

// Day-2 pack tag change fixture: a pack with two versions whose UIDs differ, so tests can check that an update
// sends the UID of the tag being applied.
const (
	TwoVersionPackName  = "two-version-pack"
	TwoVersionPackOld   = "0.8.1"
	TwoVersionPackNew   = "0.16.2"
	TwoVersionPackOldID = "two-version-pack-uid-0.8.1"
	TwoVersionPackNewID = "two-version-pack-uid-0.16.2"
)

// packsNameRegistryUIDHandler serves GET /v1/packs/{packName}/registries/{registryUid}: the two-version fixture for
// TwoVersionPackName, and the original static payload for every other pack.
func packsNameRegistryUIDHandler(w http.ResponseWriter, r *http.Request) {
	payload := getPacksNameRegistryUID()
	if mux.Vars(r)["packName"] == TwoVersionPackName {
		payload = &models.V1PackTagEntity{
			Name: TwoVersionPackName,
			Tags: []*models.V1PackTags{
				{PackUID: TwoVersionPackOldID, Tag: TwoVersionPackOld, Version: TwoVersionPackOld},
				{PackUID: TwoVersionPackNewID, Tag: TwoVersionPackNew, Version: TwoVersionPackNew},
			},
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(payload)
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

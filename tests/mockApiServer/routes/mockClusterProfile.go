package routes

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

const (
	clusterProfileUID1 = "cluster-profile-import-1"
	clusterProfileUID2 = "cluster-profile-import-2"

	// clusterProfileGetErrorUID drives the GetClusterProfile error branch inside
	// setReplaceWithProfileForExisting (cluster_common_profiles.go), which otherwise never
	// errors against the always-200 static fixture.
	clusterProfileGetErrorUID = "cluster-profile-get-error-uid"
)

func getClusterProfilesMetadataResponse() *models.V1ClusterProfilesMetadata {
	return &models.V1ClusterProfilesMetadata{
		Items: []*models.V1ClusterProfileMetadata{
			{
				Metadata: &models.V1ObjectEntity{
					Name: "test-cluster-profile-1",
					UID:  clusterProfileUID1,
				},
				Spec: &models.V1ClusterProfileMetadataSpec{
					CloudType: "aws",
					Version:   "1.0.0",
				},
			},
			{
				Metadata: &models.V1ObjectEntity{
					Name: "test-cluster-profile-2",
					UID:  clusterProfileUID2,
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
			UID:                   clusterProfileUID1,
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
						Name:        strPtr("k8"),
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

// clusterProfileGetHandler serves GET /v1/clusterprofiles/{uid} by dispatching on UID.
// The default branch preserves the original static payload (name
// "test-cluster-profile-1", Published.Type "cluster") so every pre-existing test keeps
// passing. clusterProfileUID2 gets its own addon-typed payload matching the name/type it
// already has as an attached ClusterProfileTemplate in mockCluster.go's default cluster
// fixture (name "test-cluster-profile-2", type "addon") — needed so computeProfilesToDelete's
// final defensive GetClusterProfile(oldUID) check (cluster_common_profiles.go) can see a
// non-infra type and actually mark the profile for deletion instead of always treating
// every UID as infra (which the single static payload did for every UID before this change).
func clusterProfileGetHandler(w http.ResponseWriter, r *http.Request) {
	uid := mux.Vars(r)["uid"]
	w.Header().Set("Content-Type", "application/json")
	if uid == clusterProfileGetErrorUID {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(getError("500", "failed to get cluster profile"))
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(clusterProfileFixtureFor(uid))
}

func clusterProfileFixtureFor(uid string) *models.V1ClusterProfile {
	switch uid {
	case clusterProfileUID2:
		p := getClusterProfileResponse()
		p.Metadata.Name = "test-cluster-profile-2"
		p.Metadata.UID = clusterProfileUID2
		p.Spec.Published.Name = "test-cluster-profile-2"
		p.Spec.Published.Type = "addon"
		return p
	default:
		return getClusterProfileResponse()
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
			Method: "PATCH",
			Path:   "/v1/clusterprofiles/{uid}/variables",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/clusterprofiles/{uid}/variables",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
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
			Method: "POST",
			Path:   "/v1/clusterprofiles/{uid}/clone",
			Response: ResponseData{
				StatusCode: 201,
				Payload:    map[string]string{"UID": "cloned-profile-uid"},
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/clusterprofiles/{uid}",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "PATCH",
			Path:   "/v1/clusterprofiles/{uid}/metadata",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
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
			Method:  "GET",
			Path:    "/v1/clusterprofiles/{uid}",
			Handler: clusterProfileGetHandler,
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
			Method: "POST",
			Path:   "/v1/clusterprofiles",
			Response: ResponseData{
				StatusCode: http.StatusConflict,
				Payload:    getError("ClusterProfileAlreadyExists", "Cluster Profile already exists"),
			},
		},
		{
			Method: "GET",
			Path:   "/v1/dashboard/clusterprofiles/metadata",
			Response: ResponseData{
				StatusCode: 200,
				// Same metadata as positive server so Create-on-conflict can adopt via GetClusterProfileUID.
				// Data source negative tests must use a name not present in this list.
				Payload: getClusterProfilesMetadataResponse(),
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
			Path:   "/v1/clusterprofiles/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusLocked,
				Payload:    nil,
			},
		},
	}
}

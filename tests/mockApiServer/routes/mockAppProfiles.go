package routes

import (
	"github.com/spectrocloud/palette-sdk-go/api/client/version1"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

func AppProfilesRoutes() []Route {
	return []Route{
		{
			Method: "POST",
			Path:   "/v1/appProfiles",
			Response: ResponseData{
				StatusCode: 201,
				Payload:    map[string]string{"UID": "test-app-profile-test"},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/appProfiles/{uid}/tiers",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1AppProfileTiers{
					Metadata: &models.V1ObjectMeta{
						Name: "test-tier-1",
						UID:  "test-uid",
					},
					Spec: &models.V1AppProfileTiersSpec{
						AppTiers: []*models.V1AppTier{
							{
								Metadata: &models.V1ObjectMeta{
									Name: "test-tier-0",
									UID:  "test-0-uid",
								},
								Spec: &models.V1AppTierSpec{
									ContainerRegistryUID: "test",
									InstallOrder:         0,
									Manifests: []*models.V1ObjectReference{
										{
											Kind: "cluster",
											Name: "test-manifest",
											UID:  "test-manifest-uid",
										},
									},
									Properties:       nil,
									RegistryUID:      "test-reg-uid",
									SourceAppTierUID: "test-source",
									Type:             models.NewV1AppTierType("manifest"),
									Values:           "test-values",
									Version:          "1.0.0",
								},
							},
						},
					},
				},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/appProfiles/{uid}/tiers/{tierUid}/manifests/{manifestUid}",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1Manifest{
					Metadata: &models.V1ObjectMeta{
						Name: "test-manifest",
						UID:  "test-manifest-uid",
					},
					Spec: &models.V1ManifestPublishedSpec{
						Published: &models.V1ManifestData{
							Content: "test-manifest-content",
							Digest:  "test-digest",
						},
					},
				},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/appProfiles/{uid}",
			Response: ResponseData{
				StatusCode: 200,
				Payload: models.V1AppProfile{
					Metadata: &models.V1ObjectMeta{
						Name: "test-app-profile",
						UID:  "test-app-profile-id",
					},
					Spec: &models.V1AppProfileSpec{
						ParentUID: "test-parent-id",
						Template: &models.V1AppProfileTemplate{
							AppTiers: []*models.V1AppTierRef{
								{
									Name:    "test-tier-1",
									Type:    models.NewV1AppTierType("manifest"),
									UID:     "tes-uid",
									Version: "1.0.0",
								},
							},
							RegistryRefs: nil,
						},
						Version: "1.0.0",
						Versions: []*models.V1AppProfileVersion{
							{
								UID:     "v1-id",
								Version: "1.0.0",
							},
						},
					},
					Status: nil,
				},
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/appProfiles/{uid}",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    &version1.V1AppProfilesUIDDeleteNoContent{},
			},
		},
	}
}

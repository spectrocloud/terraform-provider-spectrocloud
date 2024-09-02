package routes

import (
	"github.com/spectrocloud/gomi/pkg/ptr"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)
import "github.com/spectrocloud/palette-sdk-go/api/client/v1"

func getBSLListLocation() *models.V1UserAssetsLocations {
	return &models.V1UserAssetsLocations{
		Items: []*models.V1UserAssetsLocation{
			{
				Metadata: &models.V1ObjectMeta{
					Annotations:           nil,
					CreationTimestamp:     models.V1Time{},
					DeletionTimestamp:     models.V1Time{},
					Labels:                nil,
					LastModifiedTimestamp: models.V1Time{},
					Name:                  "test-bsl-location",
					UID:                   "test-bsl-location-id",
				},
				Spec: &models.V1UserAssetsLocationSpec{},
			},
		},
	}
}

func ClusterCommonRoutes() []Route {
	s := "test-dep-1"
	return []Route{
		{
			Method: "POST",
			Path:   "/v1/appDeployments",
			Response: ResponseData{
				StatusCode: 201,
				Payload: &v1.V1AppDeploymentsVirtualClusterCreateCreated{
					AuditUID: "test-audit-id-1",
					Payload: &models.V1UID{
						UID: ptr.StringPtr(s),
					},
				},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/clustergroups/hostCluster",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1ClusterGroupsHostClusterSummary{
					Summaries: []*models.V1ClusterGroupSummary{
						{
							Metadata: &models.V1ObjectMeta{
								Name: "test-cluster-group",
								UID:  generateRandomStringUID(),
							},
							Spec: &models.V1ClusterGroupSummarySpec{
								Scope: "project",
							},
						},
					},
				},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/clustergroups/hostCluster/metadata",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1ClusterGroupsHostClusterMetadata{
					Items: []*models.V1ObjectScopeEntity{
						{
							Name:  "test-cluster-group",
							Scope: "system",
							UID:   generateRandomStringUID(),
						},
					},
				},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/users/assets/locations",
			Response: ResponseData{
				StatusCode: 200,
				Payload:    getBSLListLocation(),
			},
		},
		{
			Method: "GET",
			Path:   "/v1/overlords/vsphere/{uid}/pools",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1IPPools{
					Items: []*models.V1IPPoolEntity{
						{
							Metadata: &models.V1ObjectMeta{
								Name: "test-name",
								UID:  "test-pcg-id",
							},
						},
					},
				},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/overlords",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1Overlords{
					Items: []*models.V1Overlord{
						{
							Metadata: &models.V1ObjectMeta{
								Name: "test-pcg-name",
								UID:  "test-pcg-id",
							},
						},
					},
				},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/dashboard/workspaces",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1DashboardWorkspaces{
					Items: []*models.V1DashboardWorkspace{
						{
							Metadata: &models.V1ObjectMeta{
								Name: "test-workspace",
								UID:  "test-workspace-uid",
							},
						},
					},
				},
			},
		},
		{
			Method: "POST",
			Path:   "/v1/dashboard/appProfiles",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1AppProfilesSummary{
					AppProfiles: []*models.V1AppProfileSummary{
						{
							Metadata: &models.V1ObjectMeta{
								Name: "test-application-profile",
								UID:  "1.0.0",
							},
							Spec: &models.V1AppProfileSummarySpec{
								Version: "1.0.0",
								Versions: []*models.V1AppProfileVersion{
									{
										UID:     generateRandomStringUID(),
										Version: "1.0.0",
									},
								},
							},
						},
					},
					Listmeta: nil,
				},
			},
		},
	}
}

func ClusterCommonNegativeRoutes() []Route {
	return []Route{}
}

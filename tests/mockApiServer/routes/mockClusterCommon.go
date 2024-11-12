package routes

import (
	"github.com/spectrocloud/palette-sdk-go/api/models"

	"github.com/spectrocloud/terraform-provider-spectrocloud/util/ptr"
)

func ClusterCommonRoutes() []Route {

	return []Route{
		{
			Method: "POST",
			Path:   "/v1/spectroclusters/{uid}/upgrade/settings",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    map[string]string{"AuditUID": generateRandomStringUID()},
			},
		},
		{
			Method: "POST",
			Path:   "/v1/appDeployments",
			Response: ResponseData{
				StatusCode: 201,
				Payload:    map[string]string{"UID": "test-application-id"},
			},
		},
		{
			Method: "POST",
			Path:   "/v1/appDeployments/clusterGroup",
			Response: ResponseData{
				StatusCode: 201,
				Payload:    map[string]string{"UID": "test-application-id"},
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/appDeployments/{uid}",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "GET",
			Path:   "/v1/appDeployments/{uid}",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1AppDeployment{
					Metadata: &models.V1ObjectMeta{
						Name:        "test-app-deployment",
						UID:         "test-app-id",
						Annotations: map[string]string{"skip_apps": "skip_apps"},
						Labels:      map[string]string{"skip_apps": "skip_apps"},
					},
					Spec: &models.V1AppDeploymentSpec{
						Config: &models.V1AppDeploymentConfig{
							Target: &models.V1AppDeploymentTargetConfig{
								ClusterRef: &models.V1AppDeploymentClusterRef{
									DeploymentClusterType: "test",
									Name:                  "test-cluster-ref",
									UID:                   "test-clsuterref-uid",
								},
								EnvRef: &models.V1AppDeploymentTargetEnvironmentRef{
									Name: "test-clsuterref-name",
									Type: "test",
									UID:  "test-envref-id",
								},
							},
						},
						Profile: &models.V1AppDeploymentProfile{
							Metadata: &models.V1AppDeploymentProfileMeta{
								Name:    "test-app-profile",
								UID:     "test-app-profile-id",
								Version: "1.0.0",
							},
							Template: &models.V1AppProfileTemplate{
								AppTiers: []*models.V1AppTierRef{
									{
										Name:    "test-app-tier-name",
										Type:    "test",
										UID:     "test-app-id",
										Version: "1.0.0",
									},
								},
								RegistryRefs: []*models.V1ObjectReference{
									{
										Kind: "test-template",
										Name: "test-reg-ref-name",
										UID:  "test-reg-ref-id",
									},
								},
							},
						},
					},
					Status: &models.V1AppDeploymentStatus{
						AppTiers: []*models.V1ClusterPackStatus{
							{
								Condition: &models.V1ClusterCondition{
									LastProbeTime:      models.V1Time{},
									LastTransitionTime: models.V1Time{},
									Message:            "",
									Reason:             "",
									Status:             ptr.To("Ready"),
									Type:               nil,
								},
								EndTime:    models.V1Time{},
								Manifests:  nil,
								Name:       "test-pack-a",
								ProfileUID: "test-profile-uid",
								Services:   nil,
								StartTime:  models.V1Time{},
								Type:       "test",
								Version:    "1.0.0",
							},
						},
						LifecycleStatus: &models.V1LifecycleStatus{
							Msg:    "test msg",
							Status: "Deployed",
						},
						State: "Deployed",
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

package routes

import (
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

func getClusterConfigTemplateResponse() *models.V1ClusterTemplate {
	return &models.V1ClusterTemplate{
		Metadata: &models.V1ObjectMeta{
			Name: "test-cluster-config-template",
			UID:  "test-cluster-config-template-id",
			Labels: map[string]string{
				"env":  "test",
				"team": "platform",
			},
			Annotations: map[string]string{
				"description": "Test cluster config template",
			},
		},
		Spec: &models.V1ClusterTemplateSpec{
			CloudType: "aws",
			Profiles: []*models.V1ClusterTemplateProfile{
				{
					UID: "test-profile-uid-1",
					Variables: []*models.V1ClusterTemplateVariable{
						{
							Name:           "region",
							Value:          "us-west-2",
							AssignStrategy: "all",
						},
						{
							Name:           "instance_type",
							Value:          "t3.medium",
							AssignStrategy: "all",
						},
					},
				},
			},
			Policies: []*models.V1PolicyRef{
				{
					UID:  "test-policy-uid-1",
					Kind: "maintenance",
				},
			},
			Clusters: map[string]models.V1ClusterTemplateSpcRef{
				"cluster-uid-1": {
					ClusterUID: "cluster-uid-1",
					Name:       "test-cluster-1",
				},
				"cluster-uid-2": {
					ClusterUID: "cluster-uid-2",
					Name:       "test-cluster-2",
				},
			},
		},
		Status: &models.V1ClusterTemplateStatus{
			State: "Applied",
			ClusterStatusCounts: &models.V1ClusterReconcileStatusCounts{
				Clusters: &models.V1ClusterReconcileStatusCountsClusters{
					Applied: []string{"cluster-uid-1", "cluster-uid-2"},
					Failed:  []string{},
					Pending: []string{},
				},
			},
		},
	}
}

func getClusterConfigTemplateCreateResponse() *models.V1UID {
	uid := "test-cluster-config-template-id"
	return &models.V1UID{
		UID: &uid,
	}
}

func getClusterConfigTemplatesSummaryResponse() *models.V1ClusterTemplatesSummary {
	return &models.V1ClusterTemplatesSummary{
		Items: []*models.V1ClusterTemplateSummary{
			{
				Metadata: &models.V1ObjectMeta{
					Name: "test-cluster-config-template",
					UID:  "test-cluster-config-template-id",
					Labels: map[string]string{
						"env":  "test",
						"team": "platform",
					},
					Annotations: map[string]string{
						"description": "Test cluster config template",
					},
				},
			},
		},
	}
}

// ClusterConfigTemplateRoutes defines routes for cluster config template operations
func ClusterConfigTemplateRoutes() []Route {
	return []Route{
		{
			Method: "POST",
			Path:   "/v1/clusterTemplates",
			Response: ResponseData{
				StatusCode: 201,
				Payload:    getClusterConfigTemplateCreateResponse(),
			},
		},
		{
			Method: "POST",
			Path:   "/v1/dashboard/clusterTemplates",
			Response: ResponseData{
				StatusCode: 200,
				Payload:    getClusterConfigTemplatesSummaryResponse(),
			},
		},
		{
			Method: "GET",
			Path:   "/v1/clusterTemplates/{uid}",
			Response: ResponseData{
				StatusCode: 200,
				Payload:    getClusterConfigTemplateResponse(),
			},
		},
		{
			Method: "PATCH",
			Path:   "/v1/clusterTemplates/{uid}/metadata",
			Response: ResponseData{
				StatusCode: 204,
			},
		},
		{
			Method: "PATCH",
			Path:   "/v1/clusterTemplates/{uid}/policies",
			Response: ResponseData{
				StatusCode: 204,
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/clusterTemplates/{uid}/profiles",
			Response: ResponseData{
				StatusCode: 204,
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/clusterTemplates/{uid}",
			Response: ResponseData{
				StatusCode: 204,
			},
		},
		{
			Method: "PATCH",
			Path:   "/v1/clusterTemplates/{uid}/profiles/variables",
			Response: ResponseData{
				StatusCode: 204,
			},
		},
		{
			Method: "PATCH",
			Path:   "/v1/spectroclusters/clusterTemplates/{uid}/clusters/upgrade",
			Response: ResponseData{
				StatusCode: 204,
			},
		},
	}
}

package routes

import (
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

func getClusterConfigPolicyResponse() *models.V1SpcPolicyEntity {
	scheduleName := "weekly-maintenance"
	startCron := "0 2 * * SUN"
	durationHrs := int64(4)

	return &models.V1SpcPolicyEntity{
		Metadata: &models.V1ObjectMeta{
			Name: "test-cluster-config-policy",
			UID:  "test-cluster-config-policy-id",
			Labels: map[string]string{
				"env":  "production",
				"team": "devops",
			},
		},
		Spec: &models.V1SpcPolicySpec{
			Schedules: []*models.V1Schedule{
				{
					Name:        &scheduleName,
					StartCron:   &startCron,
					DurationHrs: &durationHrs,
				},
			},
		},
	}
}

func getClusterConfigPolicyCreateResponse() *models.V1UID {
	uid := "test-cluster-config-policy-id"
	return &models.V1UID{
		UID: &uid,
	}
}

func getClusterConfigPoliciesSummaryResponse() *models.V1SpcPoliciesSummary {
	return &models.V1SpcPoliciesSummary{
		Items: []*models.V1SpcPolicySummary{
			{
				Metadata: &models.V1ObjectMeta{
					Name: "test-cluster-config-policy",
					UID:  "test-cluster-config-policy-id",
					Labels: map[string]string{
						"env":  "production",
						"team": "devops",
					},
				},
			},
		},
	}
}

// ClusterConfigPolicyRoutes defines routes for cluster config policy operations
func ClusterConfigPolicyRoutes() []Route {
	return []Route{
		{
			Method: "POST",
			Path:   "/v1/spcPolicies/maintenance",
			Response: ResponseData{
				StatusCode: 201,
				Payload:    getClusterConfigPolicyCreateResponse(),
			},
		},
		{
			Method: "POST",
			Path:   "/v1/dashboard/spcPolicies/maintenance",
			Response: ResponseData{
				StatusCode: 200,
				Payload:    getClusterConfigPoliciesSummaryResponse(),
			},
		},
		{
			Method: "GET",
			Path:   "/v1/spcPolicies/maintenance/{uid}",
			Response: ResponseData{
				StatusCode: 200,
				Payload:    getClusterConfigPolicyResponse(),
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/spcPolicies/maintenance/{uid}",
			Response: ResponseData{
				StatusCode: 204,
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/spcPolicies/maintenance/{uid}",
			Response: ResponseData{
				StatusCode: 204,
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/spcPolicies/{uid}",
			Response: ResponseData{
				StatusCode: 204,
			},
		},
	}
}

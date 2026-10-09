package routes

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

// ClusterTemplateAttachRejectUID drives the AttachClusterTemplate API-error
// branch, simulating Hubble's real ClusterNotEligibleForAttach rejection
// (HTTP 400) for a cluster that's already attached to a different template -
// e.g. when a user changes cluster_template.id while already attached.
const ClusterTemplateAttachRejectUID = "cluster-template-attach-reject"

// getClusterTemplateProfileVariablesResponse — canned payload for GET
// /v1/clusterTemplates/{uid}/profiles/{profileUid}/variables. Returns one
// variable ("region") assigned to cluster "test-cluster-id", which is the
// cluster UID every flattenClusterTemplateVariables test in this suite uses.
func getClusterTemplateProfileVariablesResponse() *models.V1ClusterTemplateProfileVariablesResponse {
	varName := "region"
	assignmentState := "Assigned"
	clusterUID := "test-cluster-id"
	return &models.V1ClusterTemplateProfileVariablesResponse{
		Variables: []*models.V1ClusterTemplateProfileVariableWithClusters{
			{
				Variable: &models.V1Variable{
					Name:         &varName,
					DefaultValue: "us-east-1",
				},
				Clusters: []*models.V1ClusterTemplateVariableClusterAssignment{
					{
						UID:             &clusterUID,
						AssignedBy:      "spectrocluster",
						AssignedValue:   "us-east-1",
						AssignmentState: &assignmentState,
					},
				},
			},
		},
	}
}

func clusterTemplateProfileVariablesGetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(getClusterTemplateProfileVariablesResponse())
}

// clusterTemplateAttachHandler serves POST
// /v1/spectroclusters/{uid}/clusterTemplates/{templateUid}/attach,
// dispatching on the target templateUid so the Hubble rejection can be
// simulated.
func clusterTemplateAttachHandler(w http.ResponseWriter, r *http.Request) {
	templateUID := mux.Vars(r)["templateUid"]
	if templateUID == ClusterTemplateAttachRejectUID {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(getError("ClusterNotEligibleForAttach",
			"Cluster is not eligible for attach: Cluster is already attached to cluster template 'test-cluster-config-template'"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

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
			// PLT-2410 follow-up: read side of the per-profile cluster_template
			// variable assignments, used by flattenClusterTemplateVariables.
			Method:  "GET",
			Path:    "/v1/clusterTemplates/{uid}/profiles/{profileUid}/variables",
			Handler: clusterTemplateProfileVariablesGetHandler,
		},
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
		{
			// PLT-2410: Day 2 attach — binds an existing cluster to a cluster
			// template. templateUid-dispatched — see clusterTemplateAttachHandler
			// for the ClusterTemplateAttachRejectUID branch.
			Method:  "POST",
			Path:    "/v1/spectroclusters/{uid}/clusterTemplates/{templateUid}/attach",
			Handler: clusterTemplateAttachHandler,
		},
	}
}

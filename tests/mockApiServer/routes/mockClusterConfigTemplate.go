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

// ClusterTemplateAttachPaginatedClusterUID is a cluster whose attachable-template list reports more
// pages, so a template missing from the first page must not block the plan.
const ClusterTemplateAttachPaginatedClusterUID = "cluster-attach-list-paginated"

// clusterTemplatesAttachSummaryHandler serves POST /v1/dashboard/clusterTemplates/attach. Only
// "test-cluster-config-template-id" is attachable; any other template is treated as ineligible.
func clusterTemplatesAttachSummaryHandler(w http.ResponseWriter, r *http.Request) {
	var spec models.V1ClusterTemplatesAttachFilterSpec
	_ = json.NewDecoder(r.Body).Decode(&spec)

	summary := &models.V1ClusterTemplatesAttachSummary{
		Items: []*models.V1ClusterTemplateAttachItem{{
			Metadata: &models.V1ObjectMeta{UID: "test-cluster-config-template-id", Name: "test-cluster-config-template"},
			Status:   &models.V1ClusterTemplateAttachStatus{IsEligible: true, Reasons: []string{}},
		}},
		Listmeta: &models.V1ListMetaData{Count: 1},
	}
	if spec.Filter != nil && spec.Filter.ClusterUID != nil && *spec.Filter.ClusterUID == ClusterTemplateAttachPaginatedClusterUID {
		summary.Listmeta.Continue = "next-page-token"
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(summary)
}

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

// ClusterTemplateReconcileAppliedClusterUID is the only cluster UID for which
// GET /v1/clusterTemplates/{uid}/spectroclusters/{clusterUid}/reconcile
// returns a reconcile document (state Applied: "region" resolved on
// cluster-profile-import-1, "newVar" unresolved on cluster-profile-v2). Every
// other cluster gets 404, so updateClusterTemplateVariables falls back to the
// cluster template document as before.
const ClusterTemplateReconcileAppliedClusterUID = "cluster-template-reconcile-applied"

func clusterTemplateReconcileGetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if mux.Vars(r)["clusterUid"] != ClusterTemplateReconcileAppliedClusterUID {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(getError("ResourceNotFound", "cluster reconcile not found"))
		return
	}
	region, newVar := "region", "newVar"
	profile1, profile2 := "cluster-profile-import-1", "cluster-profile-v2"
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(&models.V1ClusterReconcile{
		Spec: &models.V1ClusterReconcileSpec{
			Profiles: []*models.V1ClusterTemplateProfileVariables{
				{
					UID: &profile1,
					Variables: &models.V1ClusterTemplateProfileVariablesVariables{
						ResolvedVariables: []*models.V1SpectroClusterVariable{{Name: &region, Value: "us-east-1"}},
					},
				},
				{
					UID: &profile2,
					Variables: &models.V1ClusterTemplateProfileVariablesVariables{
						UnResolvedVariables: []*models.V1SpectroClusterVariable{{Name: &newVar}},
					},
				},
			},
		},
		Status: &models.V1ClusterReconcileStatus{State: "Applied"},
	})
}

// ClusterConfigTemplateRoutes defines routes for cluster config template operations
func ClusterConfigTemplateRoutes() []Route {
	return []Route{
		{
			Method:  "GET",
			Path:    "/v1/clusterTemplates/{uid}/spectroclusters/{clusterUid}/reconcile",
			Handler: clusterTemplateReconcileGetHandler,
		},
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
		{
			// Templates attachable to a cluster, checked at plan time before a Day 2 attach.
			Method:  "POST",
			Path:    "/v1/dashboard/clusterTemplates/attach",
			Handler: clusterTemplatesAttachSummaryHandler,
		},
	}
}

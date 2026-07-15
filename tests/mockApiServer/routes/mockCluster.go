package routes

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	v1 "github.com/spectrocloud/palette-sdk-go/api/client/version1"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

// -----------------------------------------------------------------------------
// State-driven cluster fixtures (used by clusterGetHandler + overviewHandler).
//
// Batch 4 introduced state-refresh unit tests for cluster_common_crud.go and
// cluster_profile_common_crud.go. Those tests call the refresh funcs directly
// against a real V1Client backed by the mock server, but each refresh func
// needs to see the cluster in a specific status (Ready / Running / Deleted /
// Paused / SpcApply-false / etc). Rather than juggling multiple routes at the
// same path, we route by UID: the test picks its scenario by choosing the
// cluster UID it queries. Any UID not listed here falls through to the
// default fixture (test-cluster-id → the existing well-populated cluster
// that all pre-batch-4 tests rely on).
// -----------------------------------------------------------------------------

// clusterGetHandler serves GET /v1/spectroclusters/{uid} by dispatching on
// the incoming UID. The DEFAULT branch preserves the original static
// payload, so every existing cluster CRUD test keeps working; new UIDs
// added below exist purely for state-refresh coverage.
func clusterGetHandler(w http.ResponseWriter, r *http.Request) {
	uid := mux.Vars(r)["uid"]
	payload, status := clusterFixtureFor(uid)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload != nil {
		_ = json.NewEncoder(w).Encode(payload)
	}
}

// clusterFixtureFor is a pure lookup — no side effects, no state — so tests
// can rely on repeated GETs against the same UID returning the same payload.
// Add a new case here (and a corresponding constant) when a new
// resourceClusterStateRefreshFunc branch needs to be exercised.
func clusterFixtureFor(uid string) (interface{}, int) {
	switch uid {
	case "cluster-uid-not-found":
		// GetClusterWithoutStatus swallows 404 (apiutil.Is404). Used to
		// drive the "cluster == nil" branch in refresh funcs.
		return getError("ResourceNotFound", "Cluster not found"), http.StatusNotFound

	case "cluster-uid-server-error":
		// Non-404 server error — refresh funcs must propagate as err.
		return getError("500", "internal server error"), http.StatusInternalServerError

	case "cluster-uid-nil-status":
		// GetCluster treats Status==nil as "gone" and returns (nil, nil).
		// GetClusterWithoutStatus returns the cluster object with nil Status;
		// resourceClusterReadyRefreshFunc reads that and reports NotReady.
		c := getMockSpectroCluster()
		c.Status = nil
		return c, http.StatusOK

	case "cluster-uid-provisioning":
		c := getMockSpectroCluster()
		c.Status.State = "Provisioning"
		return c, http.StatusOK

	case "cluster-uid-running":
		// Running WITHOUT the "-Healthy" suffix. The refresh func's
		// second GetClusterOverview call decides whether -Healthy gets
		// appended — see overviewHandler for the paired route.
		c := getMockSpectroCluster()
		c.Status.State = "Running"
		return c, http.StatusOK

	case "cluster-uid-running-unhealthy":
		c := getMockSpectroCluster()
		c.Status.State = "Running"
		return c, http.StatusOK

	case "cluster-uid-deleted-state":
		// GetCluster returns nil when State=="Deleted"; drives the
		// "Deleted" branch of resourceClusterStateRefreshFunc.
		c := getMockSpectroCluster()
		c.Status.State = "Deleted"
		return c, http.StatusOK

	case "cluster-uid-spcapply-false":
		// resourceClusterProfileStateRefreshFunc reads Status.SpcApply.CanBeApplied.
		c := getMockSpectroCluster()
		c.Status.SpcApply = &models.V1SpcApply{CanBeApplied: false}
		return c, http.StatusOK

	case "cluster-uid-paused":
		c := getMockSpectroCluster()
		c.Status.Virtual = &models.V1Virtual{
			LifecycleStatus: &models.V1LifecycleStatus{Status: "Paused"},
		}
		return c, http.StatusOK

	case "cluster-uid-addon-ready":
		// resourceAddonDeploymentStateRefreshFunc scans Status.Conditions
		// and Status.Packs. This payload passes all checks for the
		// "packs ready" happy path.
		return getMockSpectroClusterWithAddonReady(), http.StatusOK

	case "cluster-uid-addon-node-not-ready":
		return getMockSpectroClusterWithNodeNotReady(), http.StatusOK

	case "cluster-uid-addon-profile-not-attached":
		return getMockSpectroClusterWithoutMatchingProfile(), http.StatusOK

	default:
		// test-cluster-id and every other UID: return the well-populated
		// fixture the pre-batch-4 CRUD tests already depend on.
		return getMockSpectroCluster(), http.StatusOK
	}
}

// overviewHandler serves GET /v1/dashboard/spectroclusters/{uid}/overview.
// GetClusterOverview is called by resourceClusterStateRefreshFunc after
// Status.State == "Running" to determine whether to append "-Healthy" to
// the state string. Dispatch on UID so tests can drive both branches.
func overviewHandler(w http.ResponseWriter, r *http.Request) {
	uid := mux.Vars(r)["uid"]
	summary := &models.V1SpectroClusterUIDSummary{
		Status: &models.V1SpectroClusterUIDStatusSummary{},
	}
	switch uid {
	case "cluster-uid-running-unhealthy":
		summary.Status.Health = &models.V1SpectroClusterHealthStatus{State: "Unhealthy"}
	default:
		summary.Status.Health = &models.V1SpectroClusterHealthStatus{State: "Healthy"}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(summary)
}

// getMockSpectroClusterWithAddonReady returns a cluster whose Conditions
// are all True and whose Packs match the addon profile UID used in tests.
//
// IMPORTANT: Metadata.UID is overwritten to match the fixture's dispatch
// key. resourceAddonDeploymentStateRefreshFunc closes over the cluster's
// OWN Metadata.UID and re-fetches on every poll — if we leave the default
// "test-cluster-id" here, that re-fetch returns the plain fixture (no
// packs, no conditions), the refresh func reports Profile:NotAttached
// forever, and the wait hangs until the 60-minute timeout. Every fixture
// used by state-refresh tests must set Metadata.UID to the same key
// tests use when calling the refresh func.
func getMockSpectroClusterWithAddonReady() *models.V1SpectroCluster {
	c := getMockSpectroCluster()
	c.Metadata.UID = "cluster-uid-addon-ready"
	tru := "True"
	ready := "Ready"
	c.Status.Conditions = []*models.V1ClusterCondition{
		{Status: &tru, Type: &ready},
	}
	c.Status.Packs = []*models.V1ClusterPackStatus{
		{
			ProfileUID: "test-addon-profile-uid",
			Name:       "kubernetes",
			Condition: &models.V1ClusterCondition{
				Status: &tru,
				Type:   &ready,
			},
		},
	}
	return c
}

// getMockSpectroClusterWithNodeNotReady drives the "Node:NotReady" branch
// of resourceAddonDeploymentStateRefreshFunc.
func getMockSpectroClusterWithNodeNotReady() *models.V1SpectroCluster {
	c := getMockSpectroCluster()
	c.Metadata.UID = "cluster-uid-addon-node-not-ready"
	pending := "False"
	ready := "Ready"
	c.Status.Conditions = []*models.V1ClusterCondition{
		{Status: &pending, Type: &ready},
	}
	return c
}

// getMockSpectroClusterWithoutMatchingProfile drives the
// "Profile:NotAttached" branch — nodes are ready but no pack matches the
// profile UID the test asks about.
func getMockSpectroClusterWithoutMatchingProfile() *models.V1SpectroCluster {
	c := getMockSpectroCluster()
	c.Metadata.UID = "cluster-uid-addon-profile-not-attached"
	tru := "True"
	ready := "Ready"
	c.Status.Conditions = []*models.V1ClusterCondition{
		{Status: &tru, Type: &ready},
	}
	c.Status.Packs = []*models.V1ClusterPackStatus{
		{
			ProfileUID: "some-other-profile-uid",
			Name:       "unrelated-pack",
			Condition: &models.V1ClusterCondition{
				Status: &tru,
				Type:   &ready,
			},
		},
	}
	return c
}

func getMockSpectroCluster() *models.V1SpectroCluster {
	return &models.V1SpectroCluster{
		APIVersion: "",
		Kind:       "",
		Metadata: &models.V1ObjectMeta{
			Name: "test-cluster",
			UID:  "test-cluster-id",
			Labels: map[string]string{
				"env": "test",
			},
		},
		Spec: &models.V1SpectroClusterSpec{
			CloudType:   "aws",
			ClusterType: "full",
			CloudConfigRef: &models.V1ObjectReference{
				UID: MockCloudConfigUID,
			},
			ClusterConfig: &models.V1ClusterConfig{
				ClusterMetaAttribute:        "test-cluster-meta-attributes",
				UpdateWorkerPoolsInParallel: true,
				Timezone:                    "UTC",
			},
			ClusterProfileTemplates: []*models.V1ClusterProfileTemplate{
				{
					UID:  clusterProfileUID1,
					Name: "test-cluster-profile-1",
					Type: "cluster",
				},
				{
					UID:  clusterProfileUID2,
					Name: "test-cluster-profile-2",
					Type: "addon",
				},
			},
		},
		Status: &models.V1SpectroClusterStatus{
			State: "Running",
			Repave: &models.V1ClusterRepaveStatus{
				State: models.V1ClusterRepaveStateApproved.Pointer(),
			},
			SpcApply: &models.V1SpcApply{
				CanBeApplied: true,
			},
		},
	}
}

func ClusterRoutes() []Route {
	var buffer bytes.Buffer
	return []Route{
		{
			Method: "POST",
			Path:   "/v1/dashboard/spectroclusters/search",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1SpectroClustersSummary{
					Items: []*models.V1SpectroClusterSummary{
						{
							Metadata: &models.V1ObjectMeta{
								Name: "test-cluster",
								UID:  "test-cluster-id",
							},
							SpecSummary: nil,
							Status:      nil,
						},
					},
					Listmeta: &models.V1ListMetaData{
						Continue: "",
					},
				},
			},
		},
		{
			// UID-dispatched GET — default behavior returns
			// getMockSpectroCluster(), matching the previous static
			// response so pre-batch-4 tests keep passing.
			Method:  "GET",
			Path:    "/v1/spectroclusters/{uid}",
			Handler: clusterGetHandler,
		},
		{
			// DELETE cluster — needed by resourceClusterDelete tests.
			// GET-by-uid always returns 200 in the mock, so
			// h.GetCluster(uid) inside DeleteCluster/ForceDeleteCluster
			// succeeds, then DELETE fires.
			Method: "DELETE",
			Path:   "/v1/spectroclusters/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
		{
			// Overview endpoint — GetClusterOverview is called by
			// resourceClusterStateRefreshFunc when State == "Running".
			// UID-dispatched so tests can toggle Healthy / Unhealthy.
			Method:  "GET",
			Path:    "/v1/dashboard/spectroclusters/{uid}/overview",
			Handler: overviewHandler,
		},
		{
			Method: "GET",
			Path:   "/v1/spectroclusters/{uid}/assets/kubeconfig",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &v1.V1SpectroClustersUIDKubeConfigOK{
					ContentDisposition: "test-content",
					Payload:            &buffer,
				},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/spectroclusters/{uid}/assets/adminKubeconfig",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &v1.V1SpectroClustersUIDKubeConfigOK{
					ContentDisposition: "test-content",
					Payload:            &buffer,
				},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/spectroclusters/{uid}/assets/kubeconfigclient",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &v1.V1SpectroClustersUIDKubeConfigClientGetOK{
					ContentDisposition: "test-content",
					Payload:            &buffer,
				},
			},
		},
		{
			Method: "PATCH",
			Path:   "/v1/spectroclusters/{uid}/profiles",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/spectroclusters/{uid}/profiles",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "GET",
			Path:   "/v1/spectroclusters/{uid}/variables",
			Response: ResponseData{
				StatusCode: 200,
				Payload: []*models.V1SpectroClusterVariables{
					{
						ProfileUID: strPtr(clusterProfileUID1),
						Variables: []*models.V1SpectroClusterVariableResponse{
							{
								Name:  strPtr("region"),
								Value: "us-east-1",
							},
						},
					},
				},
			},
		},
		{
			Method: "PATCH",
			Path:   "/v1/spectroclusters/{uid}/variables",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "GET",
			Path:   "/v1/spectroclusters/{uid}/features/backup",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1ClusterBackup{
					Spec: &models.V1ClusterBackupSpec{},
				},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/spectroclusters/{uid}/features/complianceScan",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1ClusterComplianceScan{
					Spec: &models.V1ClusterComplianceScanSpec{},
				},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/spectroclusters/{uid}/config/rbacs",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1ClusterRbacs{
					Items: []*models.V1ClusterRbac{},
				},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/spectroclusters/{uid}/config/namespaces",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1ClusterNamespaceResources{
					Items: []*models.V1ClusterNamespaceResource{},
				},
			},
		},
	}
}

func ClusterNegativeRoutes() []Route {
	return []Route{
		{
			Method: "POST",
			Path:   "/v1/dashboard/spectroclusters/search",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload: &models.V1SpectroClustersSummary{
					Items:    []*models.V1SpectroClusterSummary{},
					Listmeta: &models.V1ListMetaData{},
				},
			},
		},
	}
}

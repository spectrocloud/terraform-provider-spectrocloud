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
		// isClusterRunningHealthy (resource_cluster_brownfield.go) calls
		// GetClusterOverview(cluster.Metadata.UID), not the requested uid —
		// override Metadata.UID here so that lookup lands back on this same
		// dispatch key instead of falling through to the default fixture's
		// "test-cluster-id".
		c.Metadata.UID = "cluster-uid-running-unhealthy"
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

	case "cluster-uid-vcluster-running":
		// Batch 18 — for waitForVirtualClusterLifecycleResume, whose
		// target state is "Running" on Status.Virtual.LifecycleStatus.
		c := getMockSpectroCluster()
		c.Status.Virtual = &models.V1Virtual{
			LifecycleStatus: &models.V1LifecycleStatus{Status: "Running"},
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

	case "test-gcp-cluster-id":
		// Batch 3c GCP CRUD: the default fixture reports CloudType="aws"
		// which fails ValidateCloudType inside resourceClusterGcpRead.
		// Serve a GCP-typed clone under a dedicated UID.
		c := getMockSpectroCluster()
		c.Metadata.UID = "test-gcp-cluster-id"
		c.Spec.CloudType = "gcp"
		return c, http.StatusOK

	case "test-gke-cluster-id":
		// Batch 3c GKE CRUD counterpart.
		c := getMockSpectroCluster()
		c.Metadata.UID = "test-gke-cluster-id"
		c.Spec.CloudType = "gke"
		return c, http.StatusOK

	case "test-vsphere-cluster-id":
		// Batch 3f vSphere CRUD: separate UID so ValidateCloudType inside
		// resourceClusterVsphereRead sees CloudType="vsphere".
		c := getMockSpectroCluster()
		c.Metadata.UID = "test-vsphere-cluster-id"
		c.Spec.CloudType = "vsphere"
		return c, http.StatusOK

	case "test-vsphere-cluster-cloudconfig-error-id":
		// Drives resourceClusterVsphereRead's GetCloudConfigVsphere error
		// branch: CloudConfigRef points at VsphereCloudConfigErrorUID so the
		// subsequent cloud-config fetch fails.
		c := getMockSpectroCluster()
		c.Metadata.UID = "test-vsphere-cluster-cloudconfig-error-id"
		c.Spec.CloudType = "vsphere"
		c.Spec.CloudConfigRef = &models.V1ObjectReference{UID: VsphereCloudConfigErrorUID}
		return c, http.StatusOK

	case "test-edge-vsphere-cluster-id":
		// Batch 3f edge_vsphere counterpart. Edge vSphere shares the
		// vsphere cloud-config model but the outer cluster reports
		// CloudType="edge-vsphere" (see NameToCloudType).
		c := getMockSpectroCluster()
		c.Metadata.UID = "test-edge-vsphere-cluster-id"
		c.Spec.CloudType = "edge-vsphere"
		return c, http.StatusOK

	case "test-maas-cluster-id":
		// Batch 3g MaaS CRUD.
		c := getMockSpectroCluster()
		c.Metadata.UID = "test-maas-cluster-id"
		c.Spec.CloudType = "maas"
		return c, http.StatusOK

	case "test-aks-cluster-cloudconfig-error-id":
		// Drives resourceClusterAksRead's GetCloudConfigAks error branch:
		// CloudConfigRef points at AksCloudConfigGetErrorUID so the
		// subsequent cloud-config fetch fails.
		c := getMockSpectroCluster()
		c.Metadata.UID = "test-aks-cluster-cloudconfig-error-id"
		c.Spec.CloudType = "aks"
		c.Spec.CloudConfigRef = &models.V1ObjectReference{UID: AksCloudConfigGetErrorUID}
		return c, http.StatusOK

	case "test-edge-native-cluster-id":
		c := getMockSpectroCluster()
		c.Metadata.UID = "test-edge-native-cluster-id"
		c.Spec.CloudType = "edge-native"
		return c, http.StatusOK

	case "test-cloudstack-cluster-id":
		c := getMockSpectroCluster()
		c.Metadata.UID = "test-cloudstack-cluster-id"
		c.Spec.CloudType = "apache-cloudstack"
		return c, http.StatusOK

	case "cluster-uid-brownfield-import-running":
		// resourceClusterBrownfieldRead success path: State=="Running" +
		// ClusterImport populated → no "import pending" warning, and
		// getClusterImportInfo succeeds.
		c := getMockSpectroCluster()
		c.Status.State = "Running"
		c.Status.ClusterImport = &models.V1ClusterImport{
			ImportLink: "kubectl apply -f https://api.dev.spectrocloud.com/v1/spectroclusters/cluster-uid-brownfield-import-running/import/manifest",
		}
		return c, http.StatusOK

	case "cluster-uid-brownfield-import-pending":
		// Same as above but State!="Running" → drives the extra
		// "Cluster import pending" warning diagnostic.
		c := getMockSpectroCluster()
		c.Status.State = "Pending"
		c.Status.ClusterImport = &models.V1ClusterImport{
			ImportLink: "kubectl apply -f https://api.dev.spectrocloud.com/v1/spectroclusters/cluster-uid-brownfield-import-pending/import/manifest",
		}
		return c, http.StatusOK

	case "cluster-uid-overview-error", "cluster-uid-overview-missing-health":
		// GetCluster must still succeed for these — only the paired
		// overviewHandler branch differs. See overviewHandler below.
		return getMockSpectroCluster(), http.StatusOK

	default:
		// test-cluster-id and every other UID: return the well-populated
		// fixture the pre-batch-4 CRUD tests already depend on.
		return getMockSpectroCluster(), http.StatusOK
	}
}

// clusterVariablesPatchErrorUID drives the UpdateClusterProfileVariableInCluster error
// branch — any other UID gets the original blanket 204 success.
const clusterVariablesPatchErrorUID = "cluster-uid-variables-patch-error"

func clusterVariablesPatchHandler(w http.ResponseWriter, r *http.Request) {
	uid := mux.Vars(r)["uid"]
	w.Header().Set("Content-Type", "application/json")
	if uid == clusterVariablesPatchErrorUID {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(getError("500", "failed to update cluster profile variables"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// overviewHandler serves GET /v1/dashboard/spectroclusters/{uid}/overview.
// GetClusterOverview is called by resourceClusterStateRefreshFunc after
// Status.State == "Running" to determine whether to append "-Healthy" to
// the state string. Dispatch on UID so tests can drive both branches.
func overviewHandler(w http.ResponseWriter, r *http.Request) {
	uid := mux.Vars(r)["uid"]
	w.Header().Set("Content-Type", "application/json")

	// cluster-uid-overview-error drives resourceClusterBrownfieldRead's
	// "GetClusterOverview failed" branch (health_status → "Unknown").
	if uid == "cluster-uid-overview-error" {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(getError("500", "failed to get cluster overview"))
		return
	}

	summary := &models.V1SpectroClusterUIDSummary{
		Status: &models.V1SpectroClusterUIDStatusSummary{},
	}
	switch uid {
	case "cluster-uid-running-unhealthy":
		summary.Status.Health = &models.V1SpectroClusterHealthStatus{State: "Unhealthy"}
	case "cluster-uid-overview-missing-health":
		// Health left nil — drives resourceClusterBrownfieldRead's
		// "health info missing" branch (health_status → "Unknown").
	default:
		summary.Status.Health = &models.V1SpectroClusterHealthStatus{State: "Healthy"}
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(summary)
}

// -----------------------------------------------------------------------------
// UID-dispatched handlers for the four brownfield readCommonFieldsBrownfield
// policy/config lookups (GetClusterBackupConfig, GetClusterScanConfig,
// GetClusterRbacConfig, GetClusterNamespaceConfig). All four share the same
// two special UIDs so a single readCommonFieldsBrownfield test can drive the
// error branch or the "populated payload" branch regardless of which field
// it's exercising. Any other UID preserves the original static payload used
// by pre-existing tests.
// -----------------------------------------------------------------------------

const (
	clusterPolicyConfigErrorUID = "cluster-uid-policy-error"
	clusterPolicyConfigFullUID  = "cluster-uid-policy-full"
)

func clusterFeatureBackupHandler(w http.ResponseWriter, r *http.Request) {
	uid := mux.Vars(r)["uid"]
	w.Header().Set("Content-Type", "application/json")
	switch uid {
	case clusterPolicyConfigErrorUID:
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(getError("500", "failed to get cluster backup config"))
	case clusterPolicyConfigFullUID:
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(&models.V1ClusterBackup{
			Spec: &models.V1ClusterBackupSpec{
				Config: &models.V1ClusterBackupConfig{
					BackupPrefix:      "test-prefix",
					BackupLocationUID: "test-location-uid",
					DurationInHours:   24,
					Schedule: &models.V1ClusterFeatureSchedule{
						ScheduledRunTime: "0 1 * * *",
					},
				},
			},
		})
	default:
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(&models.V1ClusterBackup{Spec: &models.V1ClusterBackupSpec{}})
	}
}

func clusterFeatureScanHandler(w http.ResponseWriter, r *http.Request) {
	uid := mux.Vars(r)["uid"]
	w.Header().Set("Content-Type", "application/json")
	switch uid {
	case clusterPolicyConfigErrorUID:
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(getError("500", "failed to get cluster scan config"))
	case clusterPolicyConfigFullUID:
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(&models.V1ClusterComplianceScan{
			Spec: &models.V1ClusterComplianceScanSpec{
				DriverSpec: map[string]models.V1ComplianceScanDriverSpec{
					"other-driver": {},
				},
			},
		})
	default:
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(&models.V1ClusterComplianceScan{Spec: &models.V1ClusterComplianceScanSpec{}})
	}
}

func clusterConfigRbacsHandler(w http.ResponseWriter, r *http.Request) {
	uid := mux.Vars(r)["uid"]
	w.Header().Set("Content-Type", "application/json")
	if uid == clusterPolicyConfigErrorUID {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(getError("500", "failed to get cluster rbac config"))
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(&models.V1ClusterRbacs{Items: []*models.V1ClusterRbac{}})
}

func clusterConfigNamespacesHandler(w http.ResponseWriter, r *http.Request) {
	uid := mux.Vars(r)["uid"]
	w.Header().Set("Content-Type", "application/json")
	if uid == clusterPolicyConfigErrorUID {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(getError("500", "failed to get cluster namespace config"))
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(&models.V1ClusterNamespaceResources{Items: []*models.V1ClusterNamespaceResource{}})
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
			// UID-dispatched so tests can drive the UpdateClusterProfileVariableInCluster
			// error branch inside updateProfiles/updateClusterTemplateVariables
			// (cluster_common_profiles.go), which the previous blanket-204 response
			// could never exercise.
			Method:  "PATCH",
			Path:    "/v1/spectroclusters/{uid}/variables",
			Handler: clusterVariablesPatchHandler,
		},
		{
			// UID-dispatched — see clusterFeatureBackupHandler for the
			// clusterPolicyConfigErrorUID / clusterPolicyConfigFullUID
			// branches used by brownfield readCommonFieldsBrownfield tests.
			Method:  "GET",
			Path:    "/v1/spectroclusters/{uid}/features/backup",
			Handler: clusterFeatureBackupHandler,
		},
		{
			Method:  "GET",
			Path:    "/v1/spectroclusters/{uid}/features/complianceScan",
			Handler: clusterFeatureScanHandler,
		},
		{
			Method:  "GET",
			Path:    "/v1/spectroclusters/{uid}/config/rbacs",
			Handler: clusterConfigRbacsHandler,
		},
		{
			Method:  "GET",
			Path:    "/v1/spectroclusters/{uid}/config/namespaces",
			Handler: clusterConfigNamespacesHandler,
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

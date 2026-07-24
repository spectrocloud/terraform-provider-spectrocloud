package routes

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

// GkeCloudConfigGetErrorUID drives the GetCloudConfigGke API-error branch
// inside resourceClusterGkeUpdate (the unconditional fetch at the top of
// the function).
const GkeCloudConfigGetErrorUID = "gke-cloud-config-get-error"

// GkeCloudConfigUpdateErrorUID drives the UpdateCloudConfigGke API-error
// branch inside resourceClusterGkeUpdate's cloud_config HasChange block.
const GkeCloudConfigUpdateErrorUID = "gke-cloud-config-update-error"

// gkeCloudConfigGetHandler serves GET /v1/cloudconfigs/gke/{configUid},
// dispatching on the config UID so resourceClusterGkeUpdate's
// GetCloudConfigGke error branch can be exercised.
func gkeCloudConfigGetHandler(w http.ResponseWriter, r *http.Request) {
	configUID := mux.Vars(r)["configUid"]
	w.Header().Set("Content-Type", "application/json")
	if configUID == GkeCloudConfigGetErrorUID {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(getError("500", "failed to get gke cloud config"))
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(getMockGkeCloudConfig())
}

// gkeCloudConfigUpdateHandler serves PUT
// /v1/cloudconfigs/gke/{configUid}/clusterConfig, dispatching on the config
// UID so resourceClusterGkeUpdate's UpdateCloudConfigGke error branch can be
// exercised.
func gkeCloudConfigUpdateHandler(w http.ResponseWriter, r *http.Request) {
	configUID := mux.Vars(r)["configUid"]
	w.Header().Set("Content-Type", "application/json")
	if configUID == GkeCloudConfigUpdateErrorUID {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(getError("500", "failed to update gke cloud config"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// getMockGkeCloudConfig returns the payload for GET
// /v1/cloudconfigs/gke/{configUid}. GKE reuses V1GcpCloudConfig — see
// palette-sdk-go/client/cluster_gke.go — so there's no separate model.
//
// The mock returns a single worker pool. GKE doesn't distinguish
// control-plane pools the same way as unmanaged GCP (Google manages the
// control plane), so IsControlPlane on machine pools is effectively
// worker-only.
func getMockGkeCloudConfig() *models.V1GcpCloudConfig {
	worker := false
	region := "us-central1"
	project := "test-gcp-project"
	instance := "n1-standard-2"

	return &models.V1GcpCloudConfig{
		Metadata: &models.V1ObjectMeta{
			Name: "gke-cloud-config",
			UID:  MockCloudConfigUID,
		},
		Spec: &models.V1GcpCloudConfigSpec{
			CloudAccountRef: &models.V1ObjectReference{
				UID: MockGcpCloudAccountUID,
			},
			ClusterConfig: &models.V1GcpClusterConfig{
				Project: &project,
				Region:  &region,
			},
			MachinePoolConfig: []*models.V1GcpMachinePoolConfig{
				{
					Name:           "worker-pool",
					InstanceType:   &instance,
					IsControlPlane: &worker,
					Size:           2,
					RootDeviceSize: 60,
				},
			},
		},
	}
}

// GkeClusterRoutes wires the GKE cluster + cloud-config CRUD endpoints.
// Endpoints share the shape of GcpClusterRoutes; only the URL prefix and
// the Create response UID differ.
func GkeClusterRoutes() []Route {
	return []Route{
		{
			Method: "POST",
			Path:   "/v1/spectroclusters/gke",
			Response: ResponseData{
				StatusCode: http.StatusCreated,
				Payload:    map[string]string{"UID": "test-gke-cluster-id"},
			},
		},
		{
			// UID-dispatched — see gkeCloudConfigGetHandler for the
			// GkeCloudConfigGetErrorUID branch.
			Method:  "GET",
			Path:    "/v1/cloudconfigs/gke/{configUid}",
			Handler: gkeCloudConfigGetHandler,
		},
		{
			// UID-dispatched — see gkeCloudConfigUpdateHandler for the
			// GkeCloudConfigUpdateErrorUID branch.
			Method:  "PUT",
			Path:    "/v1/cloudconfigs/gke/{configUid}/clusterConfig",
			Handler: gkeCloudConfigUpdateHandler,
		},
		{
			Method: "POST",
			Path:   "/v1/cloudconfigs/gke/{configUid}/machinePools",
			Response: ResponseData{
				StatusCode: http.StatusCreated,
				Payload:    map[string]string{"UID": "test-gke-machine-pool-id"},
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/cloudconfigs/gke/{configUid}/machinePools/{machinePoolName}",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/cloudconfigs/gke/{configUid}/machinePools/{machinePoolName}",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
		{
			Method: "GET",
			Path:   "/v1/cloudconfigs/gke/{configUid}/machinePools/{machinePoolName}/machines",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload: &models.V1GcpMachines{
					Items: []*models.V1GcpMachine{},
				},
			},
		},
	}
}

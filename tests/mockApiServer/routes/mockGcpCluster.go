package routes

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

// mockGcpCloudAccountUID is the fixture UID the GCP wave2 CRUD tests
// expect back from flattenCloudConfigGcp. Kept as a constant because
// the test and the mock have to agree on the string.
const MockGcpCloudAccountUID = "test-gcp-account-id-1"

// GcpCloudConfigGetErrorUID drives the GetCloudConfigGcp API-error branch
// inside resourceClusterGcpUpdate (the unconditional fetch at the top of
// the function).
const GcpCloudConfigGetErrorUID = "gcp-cloud-config-get-error"

// gcpCloudConfigGetHandler serves GET /v1/cloudconfigs/gcp/{configUid},
// dispatching on the config UID so resourceClusterGcpUpdate's
// GetCloudConfigGcp error branch can be exercised.
func gcpCloudConfigGetHandler(w http.ResponseWriter, r *http.Request) {
	configUID := mux.Vars(r)["configUid"]
	w.Header().Set("Content-Type", "application/json")
	if configUID == GcpCloudConfigGetErrorUID {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(getError("500", "failed to get gcp cloud config"))
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(getMockGcpCloudConfig())
}

// getMockGcpCloudConfig returns the payload the mock's GET
// /v1/cloudconfigs/gcp/{configUid} route serves. It mirrors what the
// wave2 test's prepareGcpClusterResourceData will Set(), so a round-
// trip Read after Create asserts drift-free.
//
// The MachinePoolConfig has one control-plane pool (matching the
// defaultGcpMachinePool fixture) plus one worker pool — enough to
// exercise flattenMachinePoolConfigsGcp's is-control-plane branching
// on a single Read.
func getMockGcpCloudConfig() *models.V1GcpCloudConfig {
	cp := true
	worker := false
	region := "us-central1"
	project := "test-gcp-project"
	instance := "n1-standard-2"

	return &models.V1GcpCloudConfig{
		Metadata: &models.V1ObjectMeta{
			Name: "gcp-cloud-config",
			UID:  MockCloudConfigUID,
		},
		Spec: &models.V1GcpCloudConfigSpec{
			CloudAccountRef: &models.V1ObjectReference{
				UID: MockGcpCloudAccountUID,
			},
			ClusterConfig: &models.V1GcpClusterConfig{
				Project: &project,
				Region:  &region,
				Network: "test-network",
			},
			MachinePoolConfig: []*models.V1GcpMachinePoolConfig{
				{
					Name:                    "cp-pool",
					InstanceType:            &instance,
					IsControlPlane:          &cp,
					Size:                    1,
					UseControlPlaneAsWorker: true,
					RootDeviceSize:          65,
					Azs:                     []string{"us-central1-a"},
				},
				{
					Name:           "worker-pool",
					InstanceType:   &instance,
					IsControlPlane: &worker,
					Size:           2,
					RootDeviceSize: 65,
					Azs:            []string{"us-central1-a"},
				},
			},
		},
	}
}

// GcpClusterRoutes wires the GCP cluster + cloud-config CRUD endpoints.
// Structure mirrors AwsClusterRoutes so the wave2 tests can be written
// with the same shape as their AWS counterparts.
//
// The machinePools GET (last route) returns an empty V1GcpMachines list
// so flattenNodeMaintenanceStatus in resource_cluster_gcp.Read walks the
// empty-status path — nothing to reconcile.
func GcpClusterRoutes() []Route {
	return []Route{
		{
			Method: "POST",
			Path:   "/v1/spectroclusters/gcp",
			Response: ResponseData{
				StatusCode: http.StatusCreated,
				Payload:    map[string]string{"UID": "test-gcp-cluster-id"},
			},
		},
		{
			// UID-dispatched — see gcpCloudConfigGetHandler for the
			// GcpCloudConfigGetErrorUID branch.
			Method:  "GET",
			Path:    "/v1/cloudconfigs/gcp/{configUid}",
			Handler: gcpCloudConfigGetHandler,
		},
		{
			Method: "PUT",
			Path:   "/v1/cloudconfigs/gcp/{configUid}/clusterConfig",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
		{
			Method: "POST",
			Path:   "/v1/cloudconfigs/gcp/{configUid}/machinePools",
			Response: ResponseData{
				StatusCode: http.StatusCreated,
				Payload:    map[string]string{"UID": "test-gcp-machine-pool-id"},
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/cloudconfigs/gcp/{configUid}/machinePools/{machinePoolName}",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/cloudconfigs/gcp/{configUid}/machinePools/{machinePoolName}",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
		{
			// Node-status list — flattenNodeMaintenanceStatus reads this
			// during Read to reconcile per-node maintenance flags. Empty
			// payload is fine; the func handles no-machines gracefully.
			Method: "GET",
			Path:   "/v1/cloudconfigs/gcp/{configUid}/machinePools/{machinePoolName}/machines",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload: &models.V1GcpMachines{
					Items: []*models.V1GcpMachine{},
				},
			},
		},
	}
}

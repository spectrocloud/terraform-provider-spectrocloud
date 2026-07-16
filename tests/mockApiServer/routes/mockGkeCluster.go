package routes

import (
	"net/http"

	"github.com/spectrocloud/palette-sdk-go/api/models"
)

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
			Method: "GET",
			Path:   "/v1/cloudconfigs/gke/{configUid}",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    getMockGkeCloudConfig(),
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/cloudconfigs/gke/{configUid}/clusterConfig",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
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

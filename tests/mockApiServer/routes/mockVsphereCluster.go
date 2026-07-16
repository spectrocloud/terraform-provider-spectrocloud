package routes

import (
	"net/http"

	"github.com/spectrocloud/palette-sdk-go/api/models"
)

// MockVsphereCloudAccountUID — the fixture UID a vSphere Read expects
// echoed back through the CloudAccountRef. Test files reference this
// constant so both sides stay in lockstep.
const MockVsphereCloudAccountUID = "test-vsphere-account-id-1"

// getMockVsphereCloudConfig returns the payload for GET
// /v1/cloudconfigs/vsphere/{configUid}. Shape mirrors what the resource
// expects after Create → Read (a single CP+worker machine pool with
// placement and a compute size).
func getMockVsphereCloudConfig() *models.V1VsphereCloudConfig {
	cp := true
	worker := false
	return &models.V1VsphereCloudConfig{
		Metadata: &models.V1ObjectMeta{
			Name: "vsphere-cloud-config",
			UID:  MockCloudConfigUID,
		},
		Spec: &models.V1VsphereCloudConfigSpec{
			CloudAccountRef: &models.V1ObjectReference{
				UID: MockVsphereCloudAccountUID,
			},
			ClusterConfig: &models.V1VsphereClusterConfig{
				ControlPlaneEndpoint: &models.V1ControlPlaneEndPoint{
					Type: "VIP",
					Host: "10.0.0.100",
				},
				NtpServers: []string{"time.google.com"},
				Placement: &models.V1VspherePlacementConfig{
					Datacenter: "test-datacenter",
					Folder:     "test-folder",
				},
				SSHKeys: []string{"ssh-rsa AAAA"},
			},
			MachinePoolConfig: []*models.V1VsphereMachinePoolConfig{
				{
					Name:                    "cp-pool",
					IsControlPlane:          &cp,
					Size:                    1,
					UseControlPlaneAsWorker: true,
					InstanceType: &models.V1VsphereInstanceType{
						DiskGiB:   types32Ptr(60),
						MemoryMiB: types64Ptr(8192),
						NumCPUs:   types32Ptr(4),
					},
					Placements: []*models.V1VspherePlacementConfig{
						{
							Cluster:      "test-cluster",
							ResourcePool: "test-pool",
							Datastore:    "test-datastore",
							Network: &models.V1VsphereNetworkConfig{
								NetworkName: strPtr("test-network"),
							},
						},
					},
				},
				{
					Name:           "worker-pool",
					IsControlPlane: &worker,
					Size:           2,
					InstanceType: &models.V1VsphereInstanceType{
						DiskGiB:   types32Ptr(60),
						MemoryMiB: types64Ptr(8192),
						NumCPUs:   types32Ptr(4),
					},
					Placements: []*models.V1VspherePlacementConfig{
						{
							Cluster:      "test-cluster",
							ResourcePool: "test-pool",
							Datastore:    "test-datastore",
							Network: &models.V1VsphereNetworkConfig{
								NetworkName: strPtr("test-network"),
							},
						},
					},
				},
			},
		},
	}
}

// VsphereClusterRoutes wires the vSphere-specific endpoints. See the
// GCP/AWS variants for the underlying pattern.
func VsphereClusterRoutes() []Route {
	return []Route{
		{
			Method: "POST",
			Path:   "/v1/spectroclusters/vsphere",
			Response: ResponseData{
				StatusCode: http.StatusCreated,
				Payload:    map[string]string{"UID": "test-vsphere-cluster-id"},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/cloudconfigs/vsphere/{configUid}",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    getMockVsphereCloudConfig(),
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/cloudconfigs/vsphere/{configUid}/clusterConfig",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
		{
			Method: "POST",
			Path:   "/v1/cloudconfigs/vsphere/{configUid}/machinePools",
			Response: ResponseData{
				StatusCode: http.StatusCreated,
				Payload:    map[string]string{"UID": "test-vsphere-machine-pool-id"},
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/cloudconfigs/vsphere/{configUid}/machinePools/{machinePoolName}",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/cloudconfigs/vsphere/{configUid}/machinePools/{machinePoolName}",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
		{
			Method: "GET",
			Path:   "/v1/cloudconfigs/vsphere/{configUid}/machinePools/{machinePoolName}/machines",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload: &models.V1VsphereMachines{
					Items: []*models.V1VsphereMachine{},
				},
			},
		},
	}
}

// Tiny pointer helpers local to this file — the vSphere models take a
// mix of int32/int64 pointers.
func types32Ptr(v int32) *int32 { return &v }
func types64Ptr(v int64) *int64 { return &v }

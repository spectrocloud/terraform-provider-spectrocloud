package routes

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

// MockVsphereCloudAccountUID — the fixture UID a vSphere Read expects
// echoed back through the CloudAccountRef. Test files reference this
// constant so both sides stay in lockstep.
const MockVsphereCloudAccountUID = "test-vsphere-account-id-1"

// VsphereCloudConfigErrorUID drives the GetCloudConfigVsphere /
// GetCloudConfigEdgeVsphere API-error branches (both vsphere and edge
// vsphere resources hit the same /v1/cloudconfigs/vsphere/{configUid}
// endpoint under the hood).
const VsphereCloudConfigErrorUID = "vsphere-cloud-config-error"

// VsphereCloudConfigUpdateErrorUID drives the UpdateCloudConfigVsphere
// API-error branch inside resourceClusterVsphereUpdate's cloud_config
// HasChange block (PUT .../clusterConfig).
const VsphereCloudConfigUpdateErrorUID = "vsphere-cloud-config-update-error"

// vsphereCloudConfigGetHandler serves GET
// /v1/cloudconfigs/vsphere/{configUid}, dispatching on the config UID so
// the GetCloudConfigVsphere/GetCloudConfigEdgeVsphere error branches can
// be exercised.
func vsphereCloudConfigGetHandler(w http.ResponseWriter, r *http.Request) {
	configUID := mux.Vars(r)["configUid"]
	w.Header().Set("Content-Type", "application/json")
	if configUID == VsphereCloudConfigErrorUID {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(getError("500", "failed to get vsphere cloud config"))
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(getMockVsphereCloudConfig())
}

// vsphereCloudConfigUpdateHandler serves PUT
// /v1/cloudconfigs/vsphere/{configUid}/clusterConfig, dispatching on the
// config UID so resourceClusterVsphereUpdate's UpdateCloudConfigVsphere
// error branch can be exercised.
func vsphereCloudConfigUpdateHandler(w http.ResponseWriter, r *http.Request) {
	configUID := mux.Vars(r)["configUid"]
	w.Header().Set("Content-Type", "application/json")
	if configUID == VsphereCloudConfigUpdateErrorUID {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(getError("500", "failed to update vsphere cloud config"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

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
			// UID-dispatched — see vsphereCloudConfigGetHandler for the
			// VsphereCloudConfigErrorUID branch.
			Method:  "GET",
			Path:    "/v1/cloudconfigs/vsphere/{configUid}",
			Handler: vsphereCloudConfigGetHandler,
		},
		{
			// UID-dispatched — see vsphereCloudConfigUpdateHandler for the
			// VsphereCloudConfigUpdateErrorUID branch.
			Method:  "PUT",
			Path:    "/v1/cloudconfigs/vsphere/{configUid}/clusterConfig",
			Handler: vsphereCloudConfigUpdateHandler,
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

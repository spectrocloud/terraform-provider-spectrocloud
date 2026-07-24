package routes

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

// EdgeNativeCloudConfigErrorUID drives the GetCloudConfigEdgeNative
// API-error branch in flattenCloudConfigEdgeNative.
const EdgeNativeCloudConfigErrorUID = "edge-native-cloud-config-error"

// EdgeNativePoolMachinesFoundHostName / UID is the host name the mock
// machines-list handler reports for the "pool-to-change" and
// "pool-to-remove" machine pool names used by
// resourceClusterEdgeNativeUpdate machine_pool diff tests — see
// edgeNativePoolMachinesListHandler.
const (
	EdgeNativePoolChangeName       = "pool-to-change"
	EdgeNativePoolRemoveName       = "pool-to-remove"
	EdgeNativePoolChangeHostName   = "host-removed"
	EdgeNativePoolChangeMachineUID = "edge-native-machine-changed-uid"
	EdgeNativePoolRemoveHostName   = "doomed-host"
	EdgeNativePoolRemoveMachineUID = "edge-native-machine-removed-uid"

	// EdgeNativePoolListErrorName drives the GetNodeListInEdgeNativeMachinePool
	// API-error branch inside resourceClusterEdgeNativeUpdate's machine_pool
	// diff handling (both the "changed pool" and "removed pool" call sites).
	EdgeNativePoolListErrorName = "pool-list-error"
)

// edgeNativePoolMachinesListHandler serves GET
// .../machinePools/{machinePoolName}/machines, dispatching on the pool
// name so machine_pool Update tests can drive both the "changed pool,
// some hosts removed" and "pool removed entirely" node-deletion branches
// in resourceClusterEdgeNativeUpdate.
func edgeNativePoolMachinesListHandler(w http.ResponseWriter, r *http.Request) {
	poolName := mux.Vars(r)["machinePoolName"]
	w.Header().Set("Content-Type", "application/json")
	if poolName == EdgeNativePoolListErrorName {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(getError("500", "failed to list edge native machines"))
		return
	}
	w.WriteHeader(http.StatusOK)
	switch poolName {
	case EdgeNativePoolChangeName:
		_ = json.NewEncoder(w).Encode(&models.V1EdgeNativeMachines{
			Items: []*models.V1EdgeNativeMachine{
				{
					Metadata: &models.V1ObjectMeta{
						Name: EdgeNativePoolChangeHostName,
						UID:  EdgeNativePoolChangeMachineUID,
					},
				},
			},
		})
	case EdgeNativePoolRemoveName:
		_ = json.NewEncoder(w).Encode(&models.V1EdgeNativeMachines{
			Items: []*models.V1EdgeNativeMachine{
				{
					Metadata: &models.V1ObjectMeta{
						Name: EdgeNativePoolRemoveHostName,
						UID:  EdgeNativePoolRemoveMachineUID,
					},
				},
			},
		})
	default:
		_ = json.NewEncoder(w).Encode(&models.V1EdgeNativeMachines{Items: []*models.V1EdgeNativeMachine{}})
	}
}

// edgeNativeCloudConfigGetHandler serves GET
// /v1/cloudconfigs/edge-native/{configUid}, dispatching on the config UID
// so flattenCloudConfigEdgeNative's GetCloudConfigEdgeNative error branch
// can be exercised.
func edgeNativeCloudConfigGetHandler(w http.ResponseWriter, r *http.Request) {
	configUID := mux.Vars(r)["configUid"]
	w.Header().Set("Content-Type", "application/json")
	if configUID == EdgeNativeCloudConfigErrorUID {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(getError("500", "failed to get edge native cloud config"))
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(getMockEdgeNativeCloudConfig())
}

// getMockEdgeNativeCloudConfig — payload for GET
// /v1/cloudconfigs/edge-native/{configUid}. Structure includes at least
// one CP pool with a host and a small worker pool.
func getMockEdgeNativeCloudConfig() *models.V1EdgeNativeCloudConfig {
	return &models.V1EdgeNativeCloudConfig{
		Metadata: &models.V1ObjectMeta{
			Name: "edge-native-cloud-config",
			UID:  MockCloudConfigUID,
		},
		Spec: &models.V1EdgeNativeCloudConfigSpec{
			ClusterConfig: &models.V1EdgeNativeClusterConfig{
				ControlPlaneEndpoint: &models.V1EdgeNativeControlPlaneEndPoint{
					Host: "10.0.0.100",
					Type: "VIP",
				},
			},
			MachinePoolConfig: []*models.V1EdgeNativeMachinePoolConfig{
				{
					Name:                    "cp-pool",
					IsControlPlane:          true,
					Size:                    1,
					UseControlPlaneAsWorker: true,
					Hosts: []*models.V1EdgeNativeHost{
						{HostUID: strPtr("edge-host-1")},
					},
				},
			},
		},
	}
}

func EdgeNativeClusterRoutes() []Route {
	return []Route{
		{
			Method: "POST",
			Path:   "/v1/spectroclusters/edge-native",
			Response: ResponseData{
				StatusCode: http.StatusCreated,
				Payload:    map[string]string{"UID": "test-edge-native-cluster-id"},
			},
		},
		{
			// UID-dispatched — see edgeNativeCloudConfigGetHandler for the
			// EdgeNativeCloudConfigErrorUID branch.
			Method:  "GET",
			Path:    "/v1/cloudconfigs/edge-native/{configUid}",
			Handler: edgeNativeCloudConfigGetHandler,
		},
		{
			Method: "PUT",
			Path:   "/v1/cloudconfigs/edge-native/{configUid}/clusterConfig",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
		{
			Method: "POST",
			Path:   "/v1/cloudconfigs/edge-native/{configUid}/machinePools",
			Response: ResponseData{
				StatusCode: http.StatusCreated,
				Payload:    map[string]string{"UID": "test-edge-native-machine-pool-id"},
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/cloudconfigs/edge-native/{configUid}/machinePools/{machinePoolName}",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/cloudconfigs/edge-native/{configUid}/machinePools/{machinePoolName}",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
		{
			// UID-dispatched by machinePoolName — see
			// edgeNativePoolMachinesListHandler for the
			// pool-to-change/pool-to-remove branches used by
			// resourceClusterEdgeNativeUpdate machine_pool diff tests.
			Method:  "GET",
			Path:    "/v1/cloudconfigs/edge-native/{configUid}/machinePools/{machinePoolName}/machines",
			Handler: edgeNativePoolMachinesListHandler,
		},
		{
			Method: "DELETE",
			Path:   "/v1/cloudconfigs/edge-native/{configUid}/machinePools/{machinePoolName}/machines/{machineUid}",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
	}
}

// int32Ptr is a tiny helper local to this file. Other mocks use their own
// versions; we keep them file-scoped to avoid coupling.
// func int32Ptr(i int32) *int32 { return &i }

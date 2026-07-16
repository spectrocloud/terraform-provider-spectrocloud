package routes

import (
	"net/http"

	"github.com/spectrocloud/palette-sdk-go/api/models"
)

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
			Method: "GET",
			Path:   "/v1/cloudconfigs/edge-native/{configUid}",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    getMockEdgeNativeCloudConfig(),
			},
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
	}
}

// int32Ptr is a tiny helper local to this file. Other mocks use their own
// versions; we keep them file-scoped to avoid coupling.
// func int32Ptr(i int32) *int32 { return &i }

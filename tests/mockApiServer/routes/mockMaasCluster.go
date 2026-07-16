package routes

import (
	"net/http"

	"github.com/spectrocloud/palette-sdk-go/api/models"
)

const MockMaasCloudAccountUID = "test-maas-account-id-1"

// getMockMaasCloudConfig — payload for GET /v1/cloudconfigs/maas/{configUid}.
// Structure mirrors what prepareMaasClusterResourceData sets so the Read
// leg after Create asserts no drift.
func getMockMaasCloudConfig() *models.V1MaasCloudConfig {
	cp := true
	domain := "test-domain"
	return &models.V1MaasCloudConfig{
		Metadata: &models.V1ObjectMeta{
			Name: "maas-cloud-config",
			UID:  MockCloudConfigUID,
		},
		Spec: &models.V1MaasCloudConfigSpec{
			CloudAccountRef: &models.V1ObjectReference{
				UID: MockMaasCloudAccountUID,
			},
			ClusterConfig: &models.V1MaasClusterConfig{
				Domain:      &domain,
				EnableLxdVM: false,
			},
			MachinePoolConfig: []*models.V1MaasMachinePoolConfig{
				{
					Name:                    "cp-pool",
					IsControlPlane:          cp,
					Size:                    1,
					UseControlPlaneAsWorker: true,
					ResourcePool:            "default",
					Azs:                     []string{"us-east-1a"},
					InstanceType: &models.V1MaasInstanceType{
						MinCPU:     4,
						MinMemInMB: 8192,
					},
				},
				{
					Name:           "worker-pool",
					IsControlPlane: false,
					Size:           2,
					ResourcePool:   "default",
					Azs:            []string{"us-east-1a"},
					InstanceType: &models.V1MaasInstanceType{
						MinCPU:     4,
						MinMemInMB: 8192,
					},
				},
			},
		},
	}
}

func MaasClusterRoutes() []Route {
	return []Route{
		{
			Method: "POST",
			Path:   "/v1/spectroclusters/maas",
			Response: ResponseData{
				StatusCode: http.StatusCreated,
				Payload:    map[string]string{"UID": "test-maas-cluster-id"},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/cloudconfigs/maas/{configUid}",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    getMockMaasCloudConfig(),
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/cloudconfigs/maas/{configUid}/clusterConfig",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
		{
			Method: "POST",
			Path:   "/v1/cloudconfigs/maas/{configUid}/machinePools",
			Response: ResponseData{
				StatusCode: http.StatusCreated,
				Payload:    map[string]string{"UID": "test-maas-machine-pool-id"},
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/cloudconfigs/maas/{configUid}/machinePools/{machinePoolName}",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/cloudconfigs/maas/{configUid}/machinePools/{machinePoolName}",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
		{
			Method: "GET",
			Path:   "/v1/cloudconfigs/maas/{configUid}/machinePools/{machinePoolName}/machines",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload: &models.V1MaasMachines{
					Items: []*models.V1MaasMachine{},
				},
			},
		},
	}
}

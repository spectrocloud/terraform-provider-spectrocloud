package routes

import (
	"net/http"

	"github.com/spectrocloud/palette-sdk-go/api/models"
)

// getMockVirtualCloudConfig — payload for GET
// /v1/cloudconfigs/virtual/{configUid}. flattenCloudConfigVirtual reads
// the first pool's InstanceType (max/min CPU/mem/storage) so the test
// can assert the "resources" field lands correctly.
func getMockVirtualCloudConfig() *models.V1VirtualCloudConfig {
	cp := true
	return &models.V1VirtualCloudConfig{
		Metadata: &models.V1ObjectMeta{
			Name: "virtual-cloud-config",
			UID:  MockCloudConfigUID,
		},
		Spec: &models.V1VirtualCloudConfigSpec{
			ClusterConfig: &models.V1VirtualClusterConfig{},
			MachinePoolConfig: []*models.V1VirtualMachinePoolConfig{
				{
					Name:           "cp-pool",
					IsControlPlane: cp,
					Size:           1,
					InstanceType: &models.V1VirtualInstanceType{
						MaxCPU:        4,
						MaxMemInMiB:   8192,
						MaxStorageGiB: 40,
						MinCPU:        2,
						MinMemInMiB:   4096,
						MinStorageGiB: 20,
					},
				},
			},
		},
	}
}

// VirtualClusterRoutes wires the virtual cluster cloud-config GET so
// flattenCloudConfigVirtual can round-trip in unit tests. Update/Create
// paths aren't included here — the resource's Update/Create goes through
// waitForClusterCreation and is out of reach for unit tests.
func VirtualClusterRoutes() []Route {
	return []Route{
		{
			Method: "GET",
			Path:   "/v1/cloudconfigs/virtual/{configUid}",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    getMockVirtualCloudConfig(),
			},
		},
	}
}

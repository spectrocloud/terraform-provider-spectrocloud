package routes

import (
	"net/http"

	"github.com/spectrocloud/palette-sdk-go/api/models"
)

const MockCloudStackCloudAccountUID = "test-cloudstack-account-id-1"

// getMockCloudStackCloudConfig — payload for GET
// /v1/cloudconfigs/apache-cloudstack/{configUid}. CloudStack pool config
// embeds V1MachinePoolBaseConfig + V1CloudStackMachineConfig — the base
// carries the pool-level fields (Name, Size, IsControlPlane).
func getMockCloudStackCloudConfig() *models.V1CloudStackCloudConfig {
	cp := true
	return &models.V1CloudStackCloudConfig{
		Metadata: &models.V1ObjectMeta{
			Name: "cloudstack-cloud-config",
			UID:  MockCloudConfigUID,
		},
		Spec: &models.V1CloudStackCloudConfigSpec{
			CloudAccountRef: &models.V1ObjectReference{
				UID: MockCloudStackCloudAccountUID,
			},
			ClusterConfig: &models.V1CloudStackClusterConfig{
				ControlPlaneEndpoint: "10.0.0.100",
				SSHKeyName:           "test-ssh-key",
				Zones: []*models.V1CloudStackZoneSpec{
					{ID: "zone-1", Name: "zone-1"},
				},
			},
			MachinePoolConfig: []*models.V1CloudStackMachinePoolConfig{
				{
					V1MachinePoolBaseConfig: models.V1MachinePoolBaseConfig{
						Name:                    "cp-pool",
						IsControlPlane:          &cp,
						Size:                    1,
						UseControlPlaneAsWorker: true,
					},
					V1CloudStackMachineConfig: models.V1CloudStackMachineConfig{
						Offering: &models.V1CloudStackResource{Name: "Medium Instance"},
						Template: &models.V1CloudStackResource{Name: "ubuntu-2004"},
					},
				},
			},
		},
	}
}

func CloudStackClusterRoutes() []Route {
	return []Route{
		{
			Method: "POST",
			Path:   "/v1/spectroclusters/apache-cloudstack",
			Response: ResponseData{
				StatusCode: http.StatusCreated,
				Payload:    map[string]string{"UID": "test-cloudstack-cluster-id"},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/cloudconfigs/apache-cloudstack/{configUid}",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    getMockCloudStackCloudConfig(),
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/cloudconfigs/apache-cloudstack/{configUid}/clusterConfig",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
		{
			Method: "POST",
			Path:   "/v1/cloudconfigs/apache-cloudstack/{configUid}/machinePools",
			Response: ResponseData{
				StatusCode: http.StatusCreated,
				Payload:    map[string]string{"UID": "test-cloudstack-machine-pool-id"},
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/cloudconfigs/apache-cloudstack/{configUid}/machinePools/{machinePoolName}",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/cloudconfigs/apache-cloudstack/{configUid}/machinePools/{machinePoolName}",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
	}
}

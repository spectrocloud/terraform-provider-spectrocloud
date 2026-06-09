package routes

import (
	"net/http"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud"
)

const (
	// MockCloudConfigUID is shared by AWS and custom-cloud mock routes for cluster read tests.
	MockCloudConfigUID = "test-cloud-config-id"

	MockCustomCloudType       = "nutanix"
	MockCustomCloudConfigUID  = MockCloudConfigUID
	MockCustomCloudAccountUID = "test-custom-account-id-1"
)

func getMockCustomCloudConfig() *models.V1CustomCloudConfig {
	cp := true
	return &models.V1CustomCloudConfig{
		Metadata: &models.V1ObjectMeta{
			Name: "custom-cloud-config",
			UID:  MockCustomCloudConfigUID,
		},
		Spec: &models.V1CustomCloudConfigSpec{
			CloudAccountRef: &models.V1ObjectReference{
				UID: MockCustomCloudAccountUID,
			},
			ClusterConfig: &models.V1CustomClusterConfig{
				Values: spectrocloud.StringPtr(`kind: Cluster
metadata:
  name: test-custom-cluster`),
			},
			MachinePoolConfig: []*models.V1CustomMachinePoolConfig{
				{
					Name:                    "pool-1",
					Size:                    3,
					IsControlPlane:          &cp,
					UseControlPlaneAsWorker: true,
					Values: `kind: KubeadmControlPlane
metadata:
  name: pool-1
spec:
  replicas: 3`,
				},
			},
		},
	}
}

func CustomCloudClusterRoutes() []Route {
	return []Route{
		{
			Method: "POST",
			Path:   "/v1/spectroclusters/cloudTypes/{cloudType}",
			Response: ResponseData{
				StatusCode: http.StatusCreated,
				Payload:    map[string]string{"UID": "test-custom-cluster-id"},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/cloudconfigs/cloudTypes/{cloudType}/{configUid}",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    getMockCustomCloudConfig(),
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/cloudconfigs/cloudTypes/{cloudType}/{configUid}/clusterConfig",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
		{
			Method: "POST",
			Path:   "/v1/cloudconfigs/cloudTypes/{cloudType}/{configUid}/machinePools",
			Response: ResponseData{
				StatusCode: http.StatusCreated,
				Payload:    nil,
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/cloudconfigs/cloudTypes/{cloudType}/{configUid}/machinePools/{machinePoolName}",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/cloudconfigs/cloudTypes/{cloudType}/{configUid}/machinePools/{machinePoolName}",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
	}
}

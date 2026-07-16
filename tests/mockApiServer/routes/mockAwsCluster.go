package routes

import (
	"net/http"

	"github.com/spectrocloud/palette-sdk-go/api/models"
)

const (
	MockAwsCloudAccountUID = "test-aws-account-id-1"
)

func getMockAwsCloudConfig() *models.V1AwsCloudConfig {
	cp := true
	region := "us-east-1"
	return &models.V1AwsCloudConfig{
		Metadata: &models.V1ObjectMeta{
			Name: "aws-cloud-config",
			UID:  MockCloudConfigUID,
		},
		Spec: &models.V1AwsCloudConfigSpec{
			CloudAccountRef: &models.V1ObjectReference{
				UID: MockAwsCloudAccountUID,
			},
			ClusterConfig: &models.V1AwsClusterConfig{
				Region:     &region,
				VpcID:      "vpc-test123",
				SSHKeyName: "test-key",
			},
			MachinePoolConfig: []*models.V1AwsMachinePoolConfig{
				{
					Name:                    "cp-pool",
					IsControlPlane:          &cp,
					InstanceType:            "t3.large",
					Size:                    1,
					UseControlPlaneAsWorker: true,
					RootDeviceSize:          20,
				},
				{
					Name:         "worker-pool",
					InstanceType: "t3.large",
					Size:         2,
				},
			},
		},
	}
}

func AwsClusterRoutes() []Route {
	return []Route{
		{
			Method: "POST",
			Path:   "/v1/spectroclusters/aws",
			Response: ResponseData{
				StatusCode: http.StatusCreated,
				Payload:    map[string]string{"UID": "test-aws-cluster-id"},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/cloudconfigs/aws/{configUid}",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    getMockAwsCloudConfig(),
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/cloudconfigs/aws/{configUid}/clusterConfig",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
		{
			Method: "POST",
			Path:   "/v1/cloudconfigs/aws/{configUid}/machinePools",
			Response: ResponseData{
				StatusCode: http.StatusCreated,
				Payload:    map[string]string{"UID": "test-aws-machine-pool-id"},
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/cloudconfigs/aws/{configUid}/machinePools/{machinePoolName}",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/cloudconfigs/aws/{configUid}/machinePools/{machinePoolName}",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
		{
			Method: "GET",
			Path:   "/v1/cloudconfigs/aws/{configUid}/machinePools/{machinePoolName}/machines",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload: &models.V1AwsMachines{
					Items: []*models.V1AwsMachine{},
				},
			},
		},
	}
}

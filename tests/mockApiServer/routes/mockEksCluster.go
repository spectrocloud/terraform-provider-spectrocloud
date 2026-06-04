package routes

import (
	"net/http"

	"github.com/spectrocloud/palette-sdk-go/api/models"
)

const (
	MockEksCloudAccountUID = "test-aws-account-id-1"
	MockEksClusterUID      = "test-eks-cluster-id"
)

func getMockEksCloudConfig() *models.V1EksCloudConfig {
	region := "us-east-1"
	cp := true
	onDemand := "on-demand"
	return &models.V1EksCloudConfig{
		Metadata: &models.V1ObjectMeta{
			Name: "eks-cloud-config",
			UID:  MockCloudConfigUID,
		},
		Spec: &models.V1EksCloudConfigSpec{
			CloudAccountRef: &models.V1ObjectReference{
				UID: MockEksCloudAccountUID,
			},
			ClusterConfig: &models.V1EksClusterConfig{
				Region:     &region,
				VpcID:      "vpc-test123",
				SSHKeyName: "test-key",
				EndpointAccess: &models.V1EksClusterConfigEndpointAccess{
					Public:  true,
					Private: false,
				},
			},
			MachinePoolConfig: []*models.V1EksMachinePoolConfig{
				{
					Name:           "cp-pool",
					IsControlPlane: &cp,
					SubnetIds: map[string]string{
						"us-east-1a": "subnet-cp",
					},
				},
				{
					Name:         "worker-pool",
					InstanceType: "m5.large",
					Size:         2,
					AmiType:      "AL2023_x86_64_STANDARD",
					RootDeviceSize: 100,
					CapacityType: &onDemand,
					SubnetIds: map[string]string{
						"us-east-1a": "subnet-worker",
					},
				},
			},
			FargateProfiles: []*models.V1FargateProfile{},
		},
	}
}

func getMockSpectroClusterEks() *models.V1SpectroCluster {
	cluster := getMockSpectroCluster()
	cluster.Metadata.UID = MockEksClusterUID
	cluster.Spec.CloudType = "eks"
	return cluster
}

func EksClusterRoutes() []Route {
	return []Route{
		{
			Method: "GET",
			Path:   "/v1/spectroclusters/" + MockEksClusterUID,
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    getMockSpectroClusterEks(),
			},
		},
		{
			Method: "POST",
			Path:   "/v1/spectroclusters/eks",
			Response: ResponseData{
				StatusCode: http.StatusCreated,
				Payload:    map[string]string{"UID": MockEksClusterUID},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/cloudconfigs/eks/{configUid}",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    getMockEksCloudConfig(),
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/cloudconfigs/eks/{configUid}/clusterConfig",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/cloudconfigs/eks/{configUid}/fargateProfiles",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
		{
			Method: "POST",
			Path:   "/v1/cloudconfigs/eks/{configUid}/machinePools",
			Response: ResponseData{
				StatusCode: http.StatusCreated,
				Payload:    map[string]string{"UID": "test-eks-machine-pool-id"},
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/cloudconfigs/eks/{configUid}/machinePools/{machinePoolName}",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/cloudconfigs/eks/{configUid}/machinePools/{machinePoolName}",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
		{
			Method: "GET",
			Path:   "/v1/cloudconfigs/eks/{configUid}/machinePools/{machinePoolName}/machines",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload: &models.V1AwsMachines{
					Items: []*models.V1AwsMachine{},
				},
			},
		},
	}
}

package routes

import (
	"net/http"

	"github.com/spectrocloud/palette-sdk-go/api/models"
)

const (
	MockAksCloudAccountUID = "test-azure-account-id-1"
	MockAksClusterUID      = "test-aks-cluster-id"
)

func getMockAksCloudConfig() *models.V1AzureCloudConfig {
	region := "eastus"
	subID := "test-subscription-id"
	return &models.V1AzureCloudConfig{
		Metadata: &models.V1ObjectMeta{
			Name: "aks-cloud-config",
			UID:  MockCloudConfigUID,
		},
		Spec: &models.V1AzureCloudConfigSpec{
			CloudAccountRef: &models.V1ObjectReference{
				UID: MockAksCloudAccountUID,
			},
			ClusterConfig: &models.V1AzureClusterConfig{
				SubscriptionID: &subID,
				ResourceGroup:  "test-rg",
				Location:       &region,
				APIServerAccessProfile: &models.V1APIServerAccessProfile{
					EnablePrivateCluster: false,
				},
			},
			MachinePoolConfig: []*models.V1AzureMachinePoolConfig{
				{
					Name:         "worker-pool",
					InstanceType: "Standard_D2s_v3",
					Size:         2,
					OsDisk: &models.V1AzureOSDisk{
						DiskSizeGB: 128,
						ManagedDisk: &models.V1ManagedDisk{
							StorageAccountType: "Premium_LRS",
						},
					},
				},
			},
		},
	}
}

func getMockSpectroClusterAks() *models.V1SpectroCluster {
	cluster := getMockSpectroCluster()
	cluster.Metadata.UID = MockAksClusterUID
	cluster.Spec.CloudType = "aks"
	return cluster
}

func AksClusterRoutes() []Route {
	return []Route{
		{
			Method: "GET",
			Path:   "/v1/spectroclusters/" + MockAksClusterUID,
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    getMockSpectroClusterAks(),
			},
		},
		{
			Method: "POST",
			Path:   "/v1/spectroclusters/aks",
			Response: ResponseData{
				StatusCode: http.StatusCreated,
				Payload:    map[string]string{"UID": MockAksClusterUID},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/cloudconfigs/aks/{configUid}",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    getMockAksCloudConfig(),
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/cloudconfigs/aks/{configUid}/clusterConfig",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
		{
			Method: "POST",
			Path:   "/v1/cloudconfigs/aks/{configUid}/machinePools",
			Response: ResponseData{
				StatusCode: http.StatusCreated,
				Payload:    map[string]string{"UID": "test-aks-machine-pool-id"},
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/cloudconfigs/aks/{configUid}/machinePools/{machinePoolName}",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/cloudconfigs/aks/{configUid}/machinePools/{machinePoolName}",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
		{
			Method: "GET",
			Path:   "/v1/cloudconfigs/aks/{configUid}/machinePools/{machinePoolName}/machines",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload: &models.V1AzureMachines{
					Items: []*models.V1AzureMachine{},
				},
			},
		},
	}
}

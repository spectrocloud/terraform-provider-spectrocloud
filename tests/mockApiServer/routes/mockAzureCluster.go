package routes

import (
	"net/http"

	"github.com/spectrocloud/palette-sdk-go/api/models"
)

const (
	MockAzureCloudAccountUID = "test-azure-account-id-1"
	MockAzureClusterUID      = "test-azure-cluster-id"
)

func getMockAzureCloudConfig() *models.V1AzureCloudConfig {
	region := "eastus"
	subID := "test-subscription-id"
	return &models.V1AzureCloudConfig{
		Metadata: &models.V1ObjectMeta{
			Name: "azure-cloud-config",
			UID:  MockCloudConfigUID,
		},
		Spec: &models.V1AzureCloudConfigSpec{
			CloudAccountRef: &models.V1ObjectReference{
				UID: MockAzureCloudAccountUID,
			},
			ClusterConfig: &models.V1AzureClusterConfig{
				SubscriptionID: &subID,
				ResourceGroup:  "test-rg",
				Location:       &region,
			},
			MachinePoolConfig: []*models.V1AzureMachinePoolConfig{
				{
					Name:         "worker-pool",
					InstanceType: "Standard_D2s_v3",
					Size:         2,
					OsType:       models.V1OsTypeLinux.Pointer(),
					Azs:          []string{"eastus-1"},
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

func getMockSpectroClusterAzure() *models.V1SpectroCluster {
	cluster := getMockSpectroCluster()
	cluster.Metadata.UID = MockAzureClusterUID
	cluster.Spec.CloudType = "azure"
	return cluster
}

func AzureClusterRoutes() []Route {
	return []Route{
		{
			Method: "GET",
			Path:   "/v1/spectroclusters/" + MockAzureClusterUID,
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    getMockSpectroClusterAzure(),
			},
		},
		{
			Method: "POST",
			Path:   "/v1/spectroclusters/azure",
			Response: ResponseData{
				StatusCode: http.StatusCreated,
				Payload:    map[string]string{"UID": MockAzureClusterUID},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/cloudconfigs/azure/{configUid}",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    getMockAzureCloudConfig(),
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/cloudconfigs/azure/{configUid}/clusterConfig",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
		{
			Method: "POST",
			Path:   "/v1/cloudconfigs/azure/{configUid}/machinePools",
			Response: ResponseData{
				StatusCode: http.StatusCreated,
				Payload:    map[string]string{"UID": "test-azure-machine-pool-id"},
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/cloudconfigs/azure/{configUid}/machinePools/{machinePoolName}",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/cloudconfigs/azure/{configUid}/machinePools/{machinePoolName}",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
		{
			Method: "GET",
			Path:   "/v1/cloudconfigs/azure/{configUid}/machinePools/{machinePoolName}/machines",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload: &models.V1AzureMachines{
					Items: []*models.V1AzureMachine{},
				},
			},
		},
	}
}

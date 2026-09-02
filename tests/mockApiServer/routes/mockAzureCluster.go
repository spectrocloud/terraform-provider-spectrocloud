package routes

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

const (
	MockAzureCloudAccountUID = "test-azure-account-id-1"
	MockAzureClusterUID      = "test-azure-cluster-id"

	// AzureCloudConfigGetErrorUID drives the GetCloudConfigAzure API-error
	// branch: both the unconditional fetch inside resourceClusterAzureUpdate
	// and the fetch inside flattenCloudConfigAzure (resourceClusterAzureRead)
	// hit /v1/cloudconfigs/azure/{configUid}.
	AzureCloudConfigGetErrorUID = "azure-cloud-config-get-error"

	// AzureCloudConfigUpdateErrorUID drives the UpdateCloudConfigAzure
	// API-error branch inside resourceClusterAzureUpdate's cloud_config
	// HasChange block.
	AzureCloudConfigUpdateErrorUID = "azure-cloud-config-update-error"
)

// azureCloudConfigGetHandler serves GET /v1/cloudconfigs/azure/{configUid},
// dispatching on the config UID so GetCloudConfigAzure's error branch can be
// exercised.
func azureCloudConfigGetHandler(w http.ResponseWriter, r *http.Request) {
	configUID := mux.Vars(r)["configUid"]
	w.Header().Set("Content-Type", "application/json")
	if configUID == AzureCloudConfigGetErrorUID {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(getError("500", "failed to get azure cloud config"))
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(getMockAzureCloudConfig())
}

// azureCloudConfigUpdateHandler serves PUT
// /v1/cloudconfigs/azure/{configUid}/clusterConfig, dispatching on the config
// UID so resourceClusterAzureUpdate's UpdateCloudConfigAzure error branch can
// be exercised.
func azureCloudConfigUpdateHandler(w http.ResponseWriter, r *http.Request) {
	configUID := mux.Vars(r)["configUid"]
	w.Header().Set("Content-Type", "application/json")
	if configUID == AzureCloudConfigUpdateErrorUID {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(getError("500", "failed to update azure cloud config"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

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
			// UID-dispatched — see azureCloudConfigGetHandler for the
			// AzureCloudConfigGetErrorUID branch.
			Method:  "GET",
			Path:    "/v1/cloudconfigs/azure/{configUid}",
			Handler: azureCloudConfigGetHandler,
		},
		{
			// UID-dispatched — see azureCloudConfigUpdateHandler for the
			// AzureCloudConfigUpdateErrorUID branch.
			Method:  "PUT",
			Path:    "/v1/cloudconfigs/azure/{configUid}/clusterConfig",
			Handler: azureCloudConfigUpdateHandler,
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

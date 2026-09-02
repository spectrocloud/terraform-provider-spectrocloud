package routes

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

const (
	MockAksCloudAccountUID = "test-azure-account-id-1"
	MockAksClusterUID      = "test-aks-cluster-id"

	// AksCloudConfigGetErrorUID drives the GetCloudConfigAks API-error
	// branch: both the unconditional fetch inside resourceClusterAksUpdate
	// and the fetch inside resourceClusterAksRead hit
	// /v1/cloudconfigs/aks/{configUid}.
	AksCloudConfigGetErrorUID = "aks-cloud-config-get-error"

	// AksCloudConfigUpdateErrorUID drives the UpdateCloudConfigAks
	// API-error branch inside resourceClusterAksUpdate's cloud_config
	// HasChange block.
	AksCloudConfigUpdateErrorUID = "aks-cloud-config-update-error"
)

// aksCloudConfigGetHandler serves GET /v1/cloudconfigs/aks/{configUid},
// dispatching on the config UID so GetCloudConfigAks's error branch can be
// exercised.
func aksCloudConfigGetHandler(w http.ResponseWriter, r *http.Request) {
	configUID := mux.Vars(r)["configUid"]
	w.Header().Set("Content-Type", "application/json")
	if configUID == AksCloudConfigGetErrorUID {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(getError("500", "failed to get aks cloud config"))
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(getMockAksCloudConfig())
}

// aksCloudConfigUpdateHandler serves PUT
// /v1/cloudconfigs/aks/{configUid}/clusterConfig, dispatching on the config
// UID so resourceClusterAksUpdate's UpdateCloudConfigAks error branch can be
// exercised.
func aksCloudConfigUpdateHandler(w http.ResponseWriter, r *http.Request) {
	configUID := mux.Vars(r)["configUid"]
	w.Header().Set("Content-Type", "application/json")
	if configUID == AksCloudConfigUpdateErrorUID {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(getError("500", "failed to update aks cloud config"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

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
			Method:  "GET",
			Path:    "/v1/cloudconfigs/aks/{configUid}",
			Handler: aksCloudConfigGetHandler,
		},
		{
			Method:  "PUT",
			Path:    "/v1/cloudconfigs/aks/{configUid}/clusterConfig",
			Handler: aksCloudConfigUpdateHandler,
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

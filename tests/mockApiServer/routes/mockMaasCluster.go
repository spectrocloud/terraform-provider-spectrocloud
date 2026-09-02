package routes

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

const MockMaasCloudAccountUID = "test-maas-account-id-1"

// MaasCloudConfigErrorUID drives the GetCloudConfigMaas API-error branch:
// both the unconditional fetch inside resourceClusterMaasUpdate and the
// flattenCloudConfigMaas call inside resourceClusterMaasRead hit
// /v1/cloudconfigs/maas/{configUid}.
const MaasCloudConfigErrorUID = "maas-cloud-config-error"

// MaasCloudConfigUpdateErrorUID drives the UpdateCloudConfigMaas API-error
// branch inside resourceClusterMaasUpdate's cloud_config HasChange block.
const MaasCloudConfigUpdateErrorUID = "maas-cloud-config-update-error"

// maasCloudConfigGetHandler serves GET /v1/cloudconfigs/maas/{configUid},
// dispatching on the config UID so GetCloudConfigMaas's error branch can be
// exercised.
func maasCloudConfigGetHandler(w http.ResponseWriter, r *http.Request) {
	configUID := mux.Vars(r)["configUid"]
	w.Header().Set("Content-Type", "application/json")
	if configUID == MaasCloudConfigErrorUID {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(getError("500", "failed to get maas cloud config"))
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(getMockMaasCloudConfig())
}

// maasCloudConfigUpdateHandler serves PUT
// /v1/cloudconfigs/maas/{configUid}/clusterConfig, dispatching on the config
// UID so resourceClusterMaasUpdate's UpdateCloudConfigMaas error branch can
// be exercised.
func maasCloudConfigUpdateHandler(w http.ResponseWriter, r *http.Request) {
	configUID := mux.Vars(r)["configUid"]
	w.Header().Set("Content-Type", "application/json")
	if configUID == MaasCloudConfigUpdateErrorUID {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(getError("500", "failed to update maas cloud config"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

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
			Method:  "GET",
			Path:    "/v1/cloudconfigs/maas/{configUid}",
			Handler: maasCloudConfigGetHandler,
		},
		{
			Method:  "PUT",
			Path:    "/v1/cloudconfigs/maas/{configUid}/clusterConfig",
			Handler: maasCloudConfigUpdateHandler,
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

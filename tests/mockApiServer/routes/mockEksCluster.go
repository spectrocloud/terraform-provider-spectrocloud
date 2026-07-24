package routes

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

const (
	MockEksCloudAccountUID = "test-aws-account-id-1"
	MockEksClusterUID      = "test-eks-cluster-id"

	// EksCloudConfigGetErrorUID drives the GetCloudConfigEks API-error
	// branch inside resourceClusterEksUpdate (the unconditional fetch
	// after the cloud_config HasChange block).
	EksCloudConfigGetErrorUID = "eks-cloud-config-get-error"

	// EksCloudConfigUpdateErrorUID drives the UpdateCloudConfigEks
	// API-error branch inside resourceClusterEksUpdate's cloud_config
	// HasChange block.
	EksCloudConfigUpdateErrorUID = "eks-cloud-config-update-error"

	// EksFargateUpdateErrorUID drives the UpdateFargateProfilesEks
	// API-error branch inside resourceClusterEksUpdate's fargate_profile
	// HasChange block.
	EksFargateUpdateErrorUID = "eks-fargate-update-error"
)

// eksCloudConfigGetHandler serves GET /v1/cloudconfigs/eks/{configUid},
// dispatching on the config UID so resourceClusterEksUpdate's unconditional
// GetCloudConfigEks error branch can be exercised.
func eksCloudConfigGetHandler(w http.ResponseWriter, r *http.Request) {
	configUID := mux.Vars(r)["configUid"]
	w.Header().Set("Content-Type", "application/json")
	if configUID == EksCloudConfigGetErrorUID {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(getError("500", "failed to get eks cloud config"))
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(getMockEksCloudConfig())
}

// eksCloudConfigUpdateHandler serves PUT
// /v1/cloudconfigs/eks/{configUid}/clusterConfig, dispatching on the config
// UID so resourceClusterEksUpdate's UpdateCloudConfigEks error branch can be
// exercised.
func eksCloudConfigUpdateHandler(w http.ResponseWriter, r *http.Request) {
	configUID := mux.Vars(r)["configUid"]
	w.Header().Set("Content-Type", "application/json")
	if configUID == EksCloudConfigUpdateErrorUID {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(getError("500", "failed to update eks cloud config"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// eksFargateUpdateHandler serves PUT
// /v1/cloudconfigs/eks/{configUid}/fargateProfiles, dispatching on the
// config UID so resourceClusterEksUpdate's UpdateFargateProfilesEks error
// branch can be exercised.
func eksFargateUpdateHandler(w http.ResponseWriter, r *http.Request) {
	configUID := mux.Vars(r)["configUid"]
	w.Header().Set("Content-Type", "application/json")
	if configUID == EksFargateUpdateErrorUID {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(getError("500", "failed to update eks fargate profiles"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

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
					Name:           "worker-pool",
					InstanceType:   "m5.large",
					Size:           2,
					AmiType:        "AL2023_x86_64_STANDARD",
					RootDeviceSize: 100,
					CapacityType:   &onDemand,
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
			// UID-dispatched — see eksCloudConfigGetHandler for the
			// EksCloudConfigGetErrorUID branch.
			Method:  "GET",
			Path:    "/v1/cloudconfigs/eks/{configUid}",
			Handler: eksCloudConfigGetHandler,
		},
		{
			// UID-dispatched — see eksCloudConfigUpdateHandler for the
			// EksCloudConfigUpdateErrorUID branch.
			Method:  "PUT",
			Path:    "/v1/cloudconfigs/eks/{configUid}/clusterConfig",
			Handler: eksCloudConfigUpdateHandler,
		},
		{
			// UID-dispatched — see eksFargateUpdateHandler for the
			// EksFargateUpdateErrorUID branch.
			Method:  "PUT",
			Path:    "/v1/cloudconfigs/eks/{configUid}/fargateProfiles",
			Handler: eksFargateUpdateHandler,
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

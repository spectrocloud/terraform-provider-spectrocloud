package routes

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

const (
	MockAwsCloudAccountUID = "test-aws-account-id-1"

	// AwsMachinesListErrorConfigUID / AwsMachinesListFoundConfigUID drive the
	// resolveNodeID (resource_cluster_brownfield.go) API-error and
	// success-with-match branches. Any other configUID keeps the original
	// empty-Items response, which resolveNodeID interprets as "not found".
	AwsMachinesListErrorConfigUID = "aws-machines-list-error"
	AwsMachinesListFoundConfigUID = "aws-machines-list-found"
	AwsMachinesListFoundNodeName  = "ip-10-0-0-1"
	AwsMachinesListFoundNodeUID   = "aws-machine-uid-1"
)

func awsPoolMachinesListHandler(w http.ResponseWriter, r *http.Request) {
	configUID := mux.Vars(r)["configUid"]
	w.Header().Set("Content-Type", "application/json")
	switch configUID {
	case AwsMachinesListErrorConfigUID:
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(getError("500", "failed to list aws machines"))
	case AwsMachinesListFoundConfigUID:
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(&models.V1AwsMachines{
			Items: []*models.V1AwsMachine{
				{
					Metadata: &models.V1ObjectMeta{
						Name: AwsMachinesListFoundNodeName,
						UID:  AwsMachinesListFoundNodeUID,
					},
				},
			},
		})
	default:
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(&models.V1AwsMachines{Items: []*models.V1AwsMachine{}})
	}
}

// awsPoolMachineUIDGetHandler serves GET .../machines/{machineUid}, used by
// GetNodeMaintenanceStatusAws. It always reports the node's current
// maintenance action as "cordon" — since resourceNodeAction only calls
// ToggleMaintenanceOnNode when the requested action differs from the
// current one, a brownfield machine_pool Update test that requests
// action=="cordon" short-circuits here without needing a toggle/wait mock.
func awsPoolMachineUIDGetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(&models.V1AwsMachine{
		Metadata: &models.V1ObjectMeta{UID: mux.Vars(r)["machineUid"]},
		Status: &models.V1CloudMachineStatus{
			MaintenanceStatus: &models.V1MachineMaintenanceStatus{
				Action: "cordon",
				State:  "Completed",
			},
		},
	})
}

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
			// UID-dispatched — see awsPoolMachinesListHandler for the
			// AwsMachinesListErrorConfigUID / AwsMachinesListFoundConfigUID
			// branches used by resolveNodeID tests.
			Method:  "GET",
			Path:    "/v1/cloudconfigs/aws/{configUid}/machinePools/{machinePoolName}/machines",
			Handler: awsPoolMachinesListHandler,
		},
		{
			// GetNodeMaintenanceStatusAws — see awsPoolMachineUIDGetHandler.
			Method:  "GET",
			Path:    "/v1/cloudconfigs/aws/{configUid}/machinePools/{machinePoolName}/machines/{machineUid}",
			Handler: awsPoolMachineUIDGetHandler,
		},
	}
}

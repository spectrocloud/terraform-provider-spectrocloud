package spectrocloud

import (
	"github.com/spectrocloud/hapi/models"
)

var NodeMaintenanceLifecycleStates = []string{
	"Completed",
	"InProgress",
	"Initiated",
	"Failed",
}

type GetMaintenanceStatus func(string, string, string, string) (*models.V1MachineMaintenanceStatus, error)

type GetNodeStatusMap func(string, string, string) (map[string]models.V1CloudMachineStatus, error)

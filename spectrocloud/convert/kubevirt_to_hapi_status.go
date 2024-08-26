package convert

import (
	"encoding/json"
	"fmt"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	kubevirtapiv1 "kubevirt.io/api/core/v1"
)

func ToHapiVmStatusM(status kubevirtapiv1.VirtualMachineStatus) (*models.V1ClusterVirtualMachineStatus, error) {
	var hapiVmStatus models.V1ClusterVirtualMachineStatus

	// Marshal the input spec to JSON
	specJson, err := json.Marshal(status)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal kubevirtapiv1.VirtualMachineSpec to JSON: %v", err)
	}

	// Unmarshal the JSON to the desired HAPI VM spec
	err = json.Unmarshal(specJson, &hapiVmStatus)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to models.V1ClusterVirtualMachineSpec: %v", err)
	}

	return &hapiVmStatus, nil
}

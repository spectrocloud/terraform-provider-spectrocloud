package convert

import (
	"fmt"

	"encoding/json"

	"github.com/spectrocloud/hapi/models"
	kubevirtapiv1 "kubevirt.io/api/core/v1"
)

func ToHapiVmSpecM(spec kubevirtapiv1.VirtualMachineSpec) (*models.V1ClusterVirtualMachineSpec, error) {
	var hapiVmSpec models.V1ClusterVirtualMachineSpec

	// Marshal the input spec to JSON
	specJson, err := json.Marshal(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal kubevirtapiv1.VirtualMachineSpec to JSON: %v", err)
	}

	// Unmarshal the JSON to the desired HAPI VM spec
	err = json.Unmarshal(specJson, &hapiVmSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to models.V1ClusterVirtualMachineSpec: %v", err)
	}

	return &hapiVmSpec, nil
}

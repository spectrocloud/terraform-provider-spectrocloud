package convert

import (
	"encoding/json"
	"fmt"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	kubevirtapiv1 "kubevirt.io/api/core/v1"
)

func ToKubevirtVMSpecM(vm *models.V1ClusterVirtualMachineSpec) (kubevirtapiv1.VirtualMachineSpec, error) {
	var kubevirtVMSpec kubevirtapiv1.VirtualMachineSpec

	// Marshal the input spec to JSON
	hapiClusterVMSpecJSON, err := json.Marshal(vm)
	if err != nil {
		return kubevirtVMSpec, fmt.Errorf("failed to marshal models.V1ClusterVirtualMachineSpec to JSON: %v", err)
	}

	// Unmarshal the JSON to the desired Kubevirt VM spec
	err = json.Unmarshal(hapiClusterVMSpecJSON, &kubevirtVMSpec)
	if err != nil {
		return kubevirtVMSpec, fmt.Errorf("failed to unmarshal JSON to kubevirtapiv1.VirtualMachineSpec: %v", err)
	}

	return kubevirtVMSpec, nil
}

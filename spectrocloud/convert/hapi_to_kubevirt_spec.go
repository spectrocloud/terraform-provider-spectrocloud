package convert

import (
	"encoding/json"
	"fmt"

	"github.com/spectrocloud/palette-sdk-go/api/models"
)

func ToKubevirtVMSpecM(vm *models.V1ClusterVirtualMachineSpec) (models.V1VMVirtualMachineInstanceSpec, error) {
	var kubevirtVMSpec models.V1VMVirtualMachineInstanceSpec

	// Marshal the input spec to JSON
	hapiClusterVMSpecJSON, err := json.Marshal(vm)
	if err != nil {
		return kubevirtVMSpec, fmt.Errorf("failed to marshal models.V1ClusterVirtualMachineSpec to JSON: %v", err)
	}

	// Unmarshal the JSON to the desired Kubevirt VM spec
	err = json.Unmarshal(hapiClusterVMSpecJSON, &kubevirtVMSpec)
	if err != nil {
		return kubevirtVMSpec, fmt.Errorf("failed to unmarshal JSON to models.V1ClusterVirtualMachineSpec: %v", err)
	}

	return kubevirtVMSpec, nil
}

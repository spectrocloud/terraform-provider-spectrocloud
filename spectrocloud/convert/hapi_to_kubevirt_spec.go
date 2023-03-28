package convert

import (
	"encoding/json"
	"fmt"

	"github.com/spectrocloud/hapi/models"
	kubevirtapiv1 "kubevirt.io/api/core/v1"

	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
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

func ToKubevirtVMSpec(i **models.V1ClusterVirtualMachineSpec) kubevirtapiv1.VirtualMachineSpec {
	// return stub
	hapiClusterVMSpec := *i
	VMSpec := &kubevirtapiv1.VirtualMachineSpec{
		Running:             types.Ptr(hapiClusterVMSpec.Running),
		RunStrategy:         ToKubevirtVMRunStrategy(hapiClusterVMSpec.RunStrategy),
		Instancetype:        ToKubevirtVMInstancetype(hapiClusterVMSpec.Instancetype),
		Preference:          nil,
		Template:            nil,
		DataVolumeTemplates: ToKubevirtDataVolumeTemplate(hapiClusterVMSpec.DataVolumeTemplates),
	}
	return *VMSpec
	//return kubevirtapiv1.VirtualMachineSpec{}
}

func ToKubevirtVMRunStrategy(runStrategy string) *kubevirtapiv1.VirtualMachineRunStrategy {
	return types.Ptr(kubevirtapiv1.VirtualMachineRunStrategy(runStrategy))
}

func ToKubevirtVMInstancetype(InsType *models.V1VMInstancetypeMatcher) *kubevirtapiv1.InstancetypeMatcher {
	var InsTypeMatcher *kubevirtapiv1.InstancetypeMatcher
	if InsType != nil {
		InsTypeMatcher = &kubevirtapiv1.InstancetypeMatcher{
			Name:            InsType.Name,
			Kind:            InsType.Kind,
			RevisionName:    InsType.RevisionName,
			InferFromVolume: InsType.InferFromVolume,
		}
	}

	return InsTypeMatcher
}

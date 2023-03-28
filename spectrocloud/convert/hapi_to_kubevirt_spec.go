package convert

import (
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
	kubevirtapiv1 "kubevirt.io/api/core/v1"
)

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

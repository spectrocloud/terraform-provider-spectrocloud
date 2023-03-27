package convert

import (
	"github.com/spectrocloud/hapi/models"
	kubevirtapiv1 "kubevirt.io/api/core/v1"
)

func ToKubevirtVMSpec(i **models.V1ClusterVirtualMachineSpec) kubevirtapiv1.VirtualMachineSpec {
	// return stub
	return kubevirtapiv1.VirtualMachineSpec{}
}

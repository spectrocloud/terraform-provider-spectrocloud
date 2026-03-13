package spectrocloud

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/schemas/virtual_machines/schema/virtualmachine"

	"encoding/json"

	"github.com/spectrocloud/palette-sdk-go/api/models"
)

func resourceVirtualMachineV2() *schema.Resource {
	return &schema.Resource{
		Schema: virtualmachine.VirtualMachineFields(),
	}
}

// ToHapiVm_V2 converts the JSON "data" from VirtualMachineResourceDataToJSON to HAPI VM.
// It is equivalent to: FromResourceData(d) then ToHapiVm(vm), but takes the serialized data instead of ResourceData.
func ToHapiVm2(data []byte) (*models.V1ClusterVirtualMachine, error) {
	// Work on a copy of the input
	jsonCopy := make([]byte, len(data))
	copy(jsonCopy, data)

	var m map[string]interface{}
	if err := json.Unmarshal(jsonCopy, &m); err != nil {
		return nil, err
	}
	var GracePeriodSeconds int64
	if vm.DeletionGracePeriodSeconds != nil {
		GracePeriodSeconds = *vm.DeletionGracePeriodSeconds
	}
	Spec, err := ToHapiVmSpecM_V2(vm.Spec)
	if err != nil {
		return nil, err
	}
	vm, err := virtualmachine.FromTFMap(m)

	return ToHapiVm(vm)
}

func ToHapiVmSpecM_V2(data []byte) (*models.V1ClusterVirtualMachineSpec, error) {
	var hapiVmSpec models.V1ClusterVirtualMachineSpec

	// Marshal the input spec to JSON
	specJson, err := json.Marshal(data)
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

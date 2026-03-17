package virtualmachineinstance

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	// kubevirtapiv1 "kubevirt.io/api/core/v1"
)

func ExpandVirtualMachineInstanceTemplateSpec(d *schema.ResourceData) (*models.V1VMVirtualMachineInstanceTemplateSpec, error) {
	result := &models.V1VMVirtualMachineInstanceTemplateSpec{}
	// Template metadata not used (removed for template).
	result.Metadata = nil

	if spec, err := expandVirtualMachineInstanceSpec(d); err == nil {
		result.Spec = spec
	} else {
		return result, err
	}

	return result, nil
}

func FlattenVirtualMachineInstanceTemplateSpec(in kubevirtapiv1.VirtualMachineInstanceTemplateSpec, resourceData *schema.ResourceData) []interface{} {
	att := make(map[string]interface{})

	// Since we removed metadata support in VM instance Spec metadata is not set.
	att["spec"] = flattenVirtualMachineInstanceSpec(in.Spec, resourceData)

	return []interface{}{att}
}

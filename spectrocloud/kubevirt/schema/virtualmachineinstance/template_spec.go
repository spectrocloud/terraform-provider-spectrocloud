package virtualmachineinstance

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

func ExpandVirtualMachineInstanceTemplateSpec(d *schema.ResourceData) (*models.V1VMVirtualMachineInstanceTemplateSpec, error) {
	result := &models.V1VMVirtualMachineInstanceTemplateSpec{}

	// we have removed metadata for template hence set empty metadata object (TBD)***
	result.Metadata = &models.V1VMObjectMeta{}

	if spec, err := expandVirtualMachineInstanceSpec(d); err == nil {
		result.Spec = spec
	} else {
		return result, err
	}

	return result, nil
}

// func FlattenVirtualMachineInstanceTemplateSpec(in kubevirtapiv1.VirtualMachineInstanceTemplateSpec, resourceData *schema.ResourceData) []interface{} {
// 	att := make(map[string]interface{})

// 	// Since we removed metadata support in VM instance Spec metadata is not set.
// 	att["spec"] = flattenVirtualMachineInstanceSpec(in.Spec, resourceData)

// 	return []interface{}{att}
// }

// FlattenVirtualMachineInstanceTemplateSpecFromVM builds the same shape as FlattenVirtualMachineInstanceTemplateSpec from Palette V1VMVirtualMachineInstanceTemplateSpec.
func FlattenVirtualMachineInstanceTemplateSpecFromVM(in *models.V1VMVirtualMachineInstanceTemplateSpec, resourceData *schema.ResourceData) []interface{} {
	att := make(map[string]interface{})
	if in == nil {
		att["spec"] = []interface{}{map[string]interface{}{}}
		return []interface{}{att}
	}
	att["spec"] = flattenVirtualMachineInstanceSpecFromVM(in.Spec, resourceData)
	return []interface{}{att}
}

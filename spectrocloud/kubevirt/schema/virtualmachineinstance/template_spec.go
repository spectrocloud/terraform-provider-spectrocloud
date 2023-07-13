package virtualmachineinstance

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	kubevirtapiv1 "kubevirt.io/api/core/v1"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/schema/k8s"
)

func virtualMachineInstanceTemplateSpecFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"metadata": k8s.NamespacedMetadataSchema("VirtualMachineInstanceTemplateSpec", false),
		"spec":     virtualMachineInstanceSpecSchema(),
	}
}

func VirtualMachineInstanceTemplateSpecSchema() *schema.Schema {
	fields := virtualMachineInstanceTemplateSpecFields()

	return &schema.Schema{
		Type: schema.TypeList,

		Description: "Template is the direct specification of VirtualMachineInstance.",
		Optional:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: fields,
		},
	}

}

func ExpandVirtualMachineInstanceTemplateSpec(d *schema.ResourceData) (*kubevirtapiv1.VirtualMachineInstanceTemplateSpec, error) {
	result := &kubevirtapiv1.VirtualMachineInstanceTemplateSpec{}

	// we have removed metadata for template hence trying to apply same metadata (TBD)***
	result.ObjectMeta = k8s.ConvertToBasicMetadata(d)

	if spec, err := expandVirtualMachineInstanceSpec(d); err == nil {
		result.Spec = spec
	} else {
		return result, err
	}

	return result, nil
}

func FlattenVirtualMachineInstanceTemplateSpec(in kubevirtapiv1.VirtualMachineInstanceTemplateSpec, resourceData *schema.ResourceData) []interface{} {
	att := make(map[string]interface{})

	att["metadata"] = k8s.FlattenMetadata(in.ObjectMeta, resourceData)
	att["spec"] = flattenVirtualMachineInstanceSpec(in.Spec, resourceData)

	return []interface{}{att}
}

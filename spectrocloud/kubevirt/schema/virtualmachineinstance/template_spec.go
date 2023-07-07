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
	//if len(virtualMachine) == 0 || virtualMachine[0] == nil {
	//	return nil, nil
	//}

	result := &kubevirtapiv1.VirtualMachineInstanceTemplateSpec{}

	//in := virtualMachine[0].(map[string]interface{})

	// we have removed metadata for template hence trying to apply same metadata (TBD)***
	result.ObjectMeta = k8s.ConvertToBasicMetadata(d)

	//if v, ok := in["metadata"].([]interface{}); ok {
	//	result.ObjectMeta = k8s.ExpandMetadata(d)
	//}
	//if v, ok := in["spec"].([]interface{}); ok {
	//	spec, err := expandVirtualMachineInstanceSpec(v)
	//	if err != nil {
	//		return result, err
	//	}
	//	result.Spec = spec
	//}
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
	att["spec"] = flattenVirtualMachineInstanceSpec(in.Spec)

	return []interface{}{att}
}

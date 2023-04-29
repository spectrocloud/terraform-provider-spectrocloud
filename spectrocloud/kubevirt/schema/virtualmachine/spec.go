package virtualmachine

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	kubevirtapiv1 "kubevirt.io/api/core/v1"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/schema/virtualmachineinstance"
)

func virtualMachineSpecFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		// "running": &schema.Schema{
		// 	Type:        schema.TypeBool,
		// 	Description: "Running controls whether the associatied VirtualMachineInstance is created or not.",
		// 	Optional:    true,
		// },
		"run_strategy": {
			Type:        schema.TypeString,
			Description: "Running state indicates the requested running state of the VirtualMachineInstance, mutually exclusive with Running.",
			Optional:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"",
				"Always",
				"Halted",
				"Manual",
				"RerunOnFailure",
			}, false),
		},
		"template":              virtualmachineinstance.VirtualMachineInstanceTemplateSpecSchema(),
		"data_volume_templates": dataVolumeTemplatesSchema(),
	}
}

func virtualMachineSpecSchema() *schema.Schema {
	fields := virtualMachineSpecFields()

	return &schema.Schema{
		Type: schema.TypeList,

		Description: "VirtualMachineSpec describes how the proper VirtualMachine should look like.",
		Optional:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: fields,
		},
	}

}

func expandVirtualMachineSpec(virtualMachine []interface{}) (kubevirtapiv1.VirtualMachineSpec, error) {
	result := kubevirtapiv1.VirtualMachineSpec{}

	if len(virtualMachine) == 0 || virtualMachine[0] == nil {
		return result, nil
	}

	in := virtualMachine[0].(map[string]interface{})

	// if v, ok := in["running"].(bool); ok {
	// 	result.Running = &v
	// }
	if v, ok := in["run_strategy"].(string); ok {
		if v != "" {
			runStrategy := kubevirtapiv1.VirtualMachineRunStrategy(v)
			result.RunStrategy = &runStrategy
		}
	}
	if v, ok := in["template"].([]interface{}); ok {
		template, err := virtualmachineinstance.ExpandVirtualMachineInstanceTemplateSpec(v)
		if err != nil {
			return result, err
		}
		result.Template = template
	}
	if v, ok := in["data_volume_templates"].([]interface{}); ok {
		dataVolumeTemplates, err := expandDataVolumeTemplates(v)
		if err != nil {
			return result, err
		}
		result.DataVolumeTemplates = dataVolumeTemplates
	}

	return result, nil
}

func flattenVirtualMachineSpec(in kubevirtapiv1.VirtualMachineSpec) []interface{} {
	att := make(map[string]interface{})

	if in.RunStrategy != nil {
		att["run_strategy"] = string(*in.RunStrategy)
	}
	if in.Template != nil {
		att["template"] = virtualmachineinstance.FlattenVirtualMachineInstanceTemplateSpec(*in.Template)
	} else {
		att["template"] = []interface{}{} // Set to empty value
	}
	if in.DataVolumeTemplates != nil {
		att["data_volume_templates"] = flattenDataVolumeTemplates(in.DataVolumeTemplates)
	} else {
		att["data_volume_templates"] = []interface{}{} // Set to empty value
	}

	return []interface{}{att}
}

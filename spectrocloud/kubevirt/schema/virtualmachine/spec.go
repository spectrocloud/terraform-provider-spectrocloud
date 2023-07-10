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

func expandVirtualMachineSpec(d *schema.ResourceData) (kubevirtapiv1.VirtualMachineSpec, error) {
	result := kubevirtapiv1.VirtualMachineSpec{}

	if v, ok := d.GetOk("run_strategy"); ok {
		if v.(string) != "" {
			runStrategy := kubevirtapiv1.VirtualMachineRunStrategy(v.(string))
			result.RunStrategy = &runStrategy
		}
	}

	if template, err := virtualmachineinstance.ExpandVirtualMachineInstanceTemplateSpec(d); err == nil && template != nil {
		result.Template = template
	} else {
		return result, err
	}

	if v, ok := d.GetOk("data_volume_templates"); ok {
		dataVolumeTemplates, err := expandDataVolumeTemplates(v.([]interface{}))
		if err != nil {
			return result, err
		}
		result.DataVolumeTemplates = dataVolumeTemplates
	}

	return result, nil
}

func flattenVirtualMachineSpec(in kubevirtapiv1.VirtualMachineSpec, resourceData *schema.ResourceData) []interface{} {
	att := make(map[string]interface{})

	if in.RunStrategy != nil {
		att["run_strategy"] = string(*in.RunStrategy)
	}
	if in.Template != nil {
		att["template"] = virtualmachineinstance.FlattenVirtualMachineInstanceTemplateSpec(*in.Template, resourceData)
	} else {
		att["template"] = []interface{}{} // Set to empty value
	}
	if in.DataVolumeTemplates != nil {
		att["data_volume_templates"] = flattenDataVolumeTemplates(in.DataVolumeTemplates, resourceData)
	} else {
		att["data_volume_templates"] = []interface{}{} // Set to empty value
	}

	return []interface{}{att}
}

func FlattenVMMToSpectroSchema(in kubevirtapiv1.VirtualMachineSpec, resourceData *schema.ResourceData) error {
	VM := flattenVirtualMachineSpec(in, resourceData)[0].(map[string]interface{})
	VMTemplate := VM["template"]
	VMTemplateSpec := VMTemplate.([]interface{})[0].(map[string]interface{})["spec"]
	VMTemplateSpecAttributes := VMTemplateSpec.([]interface{})[0].(map[string]interface{})
	if err := resourceData.Set("run_strategy", VM["run_strategy"]); err != nil {
		return err
	}
	if err := resourceData.Set("node_selector", VMTemplateSpecAttributes["node_selector"]); err != nil {
		return err
	}
	if err := resourceData.Set("affinity", VMTemplateSpecAttributes["affinity"]); err != nil {
		return err
	}
	if err := resourceData.Set("scheduler_name", VMTemplateSpecAttributes["scheduler_name"]); err != nil {
		return err
	}
	if err := resourceData.Set("hostname", VMTemplateSpecAttributes["hostname"]); err != nil {
		return err
	}
	if err := resourceData.Set("subdomain", VMTemplateSpecAttributes["subdomain"]); err != nil {
		return err
	}
	if err := resourceData.Set("dns_policy", VMTemplateSpecAttributes["dns_policy"]); err != nil {
		return err
	}
	if err := resourceData.Set("priority_class_name", VMTemplateSpecAttributes["priority_class_name"]); err != nil {
		return err
	}
	if err := resourceData.Set("network", VMTemplateSpecAttributes["network"]); err != nil {
		return err
	}
	if err := resourceData.Set("volume", VMTemplateSpecAttributes["volume"]); err != nil {
		return err
	}

	// Setting up domain
	vmDomain := VMTemplateSpecAttributes["domain"].([]interface{})[0].(map[string]interface{})
	vmTolerations := VMTemplateSpecAttributes["tolerations"]
	resource := vmDomain["resources"]
	cpu := vmDomain["cpu"]
	memory := vmDomain["memory"]
	device := vmDomain["devices"].([]interface{})[0].(map[string]interface{})
	disks := device["disk"]
	interfaces := device["interface"]

	if err := resourceData.Set("cpu", cpu); err != nil {
		return err
	}
	if err := resourceData.Set("memory", memory); err != nil {
		return err
	}
	if err := resourceData.Set("resources", resource); err != nil {
		return err
	}
	if err := resourceData.Set("disk", disks); err != nil {
		return err
	}
	if err := resourceData.Set("interface", interfaces); err != nil {
		return err
	}

	if err := resourceData.Set("tolerations", vmTolerations); err != nil {
		return err
	}

	if v, ok := VMTemplateSpecAttributes["eviction_strategy"]; ok != true {
		if err := resourceData.Set("eviction_strategy", v); err != nil {
			return err
		}
	}
	if v, ok := VMTemplateSpecAttributes["termination_grace_period_seconds"]; ok != true {
		if err := resourceData.Set("termination_grace_period_seconds", v); err != nil {
			return err
		}
	}
	if v, ok := VMTemplateSpecAttributes["liveness_probe"]; ok != true {
		if err := resourceData.Set("liveness_probe", v); err != nil {
			return err
		}
	}
	if v, ok := VMTemplateSpecAttributes["readiness_probe"]; ok != true {
		if err := resourceData.Set("readiness_probe", v); err != nil {
			return err
		}
	}
	if v, ok := VMTemplateSpecAttributes["pod_dns_config"]; ok != true {
		if err := resourceData.Set("pod_dns_config", v); err != nil {
			return err
		}
	}

	if err := resourceData.Set("data_volume_templates", VM["data_volume_templates"]); err != nil {
		return err
	}
	return nil
}

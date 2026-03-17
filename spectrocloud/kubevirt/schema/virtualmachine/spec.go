package virtualmachine

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/schema/virtualmachineinstance"
)

func ExpandVirtualMachineSpec(d *schema.ResourceData) (*models.V1ClusterVirtualMachineSpec, error) {
	result := &models.V1ClusterVirtualMachineSpec{}

	if v, ok := d.GetOk("run_strategy"); ok && v.(string) != "" {
		result.RunStrategy = v.(string)
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

// func FlattenVirtualMachineSpec(in kubevirtapiv1.VirtualMachineSpec, resourceData *schema.ResourceData) []interface{} {
// 	att := make(map[string]interface{})

// 	if in.RunStrategy != nil {
// 		att["run_strategy"] = string(*in.RunStrategy)
// 	}
// 	if in.Template != nil {
// 		att["template"] = virtualmachineinstance.FlattenVirtualMachineInstanceTemplateSpec(*in.Template, resourceData)
// 	} else {
// 		att["template"] = []interface{}{} // Set to empty value
// 	}
// 	if in.DataVolumeTemplates != nil {
// 		att["data_volume_templates"] = flattenDataVolumeTemplatesK8s(in.DataVolumeTemplates, resourceData)
// 	} else {
// 		att["data_volume_templates"] = []interface{}{} // Set to empty value
// 	}

// 	return []interface{}{att}
// }

// FlattenVirtualMachineSpecFromVM builds the same []interface{} shape as FlattenVirtualMachineSpec but from Palette V1ClusterVirtualMachineSpec.
func FlattenVirtualMachineSpecFromVM(in *models.V1ClusterVirtualMachineSpec, resourceData *schema.ResourceData) []interface{} {
	att := make(map[string]interface{})
	if in == nil {
		return []interface{}{att}
	}
	if in.RunStrategy != "" {
		att["run_strategy"] = in.RunStrategy
	}
	if in.Template != nil {
		att["template"] = virtualmachineinstance.FlattenVirtualMachineInstanceTemplateSpecFromVM(in.Template, resourceData)
	} else {
		att["template"] = []interface{}{map[string]interface{}{"spec": []interface{}{map[string]interface{}{}}}}
	}
	if len(in.DataVolumeTemplates) > 0 {
		att["data_volume_templates"] = flattenDataVolumeTemplatesFromVM(in.DataVolumeTemplates, resourceData)
	} else {
		att["data_volume_templates"] = []interface{}{}
	}
	return []interface{}{att}
}

// FlattenVMMToSpectroSchemaFromVM flattens Palette V1ClusterVirtualMachineSpec into resourceData (same keys as FlattenVMMToSpectroSchema).
func FlattenVMMToSpectroSchemaFromVM(in *models.V1ClusterVirtualMachineSpec, resourceData *schema.ResourceData) error {
	if in == nil {
		return nil
	}
	VMList := FlattenVirtualMachineSpecFromVM(in, resourceData)
	if len(VMList) == 0 {
		return nil
	}
	VM := VMList[0].(map[string]interface{})
	// template spec
	VMTemplate := VM["template"]
	templateList, ok := VMTemplate.([]interface{})
	if !ok || len(templateList) == 0 {
		return nil
	}
	VMTemplateSpec := templateList[0].(map[string]interface{})["spec"]
	specList, ok := VMTemplateSpec.([]interface{})
	if !ok || len(specList) == 0 {
		return nil
	}
	VMTemplateSpecAttributes := specList[0].(map[string]interface{})

	// domain spec (may be missing if template spec is empty)
	vmDomain, _ := VMTemplateSpecAttributes["domain"].([]interface{})
	var vmTolerations, resource, cpu, memory, firmware, features, disks, interfaces interface{}
	if len(vmDomain) > 0 {
		if domainMap, ok := vmDomain[0].(map[string]interface{}); ok {
			resource = domainMap["resources"]
			cpu = domainMap["cpu"]
			memory = domainMap["memory"]
			firmware = domainMap["firmware"]
			features = domainMap["features"]
			if devList, ok := domainMap["devices"].([]interface{}); ok && len(devList) > 0 {
				if device, ok := devList[0].(map[string]interface{}); ok {
					disks = device["disk"]
					interfaces = device["interface"]
				}
			}
		}
	}
	vmTolerations = VMTemplateSpecAttributes["tolerations"]

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
	if err := resourceData.Set("cpu", cpu); err != nil {
		return err
	}
	if err := resourceData.Set("memory", memory); err != nil {
		return err
	}
	if err := resourceData.Set("firmware", firmware); err != nil {
		return err
	}
	if err := resourceData.Set("features", features); err != nil {
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
	if v, ok := VMTemplateSpecAttributes["eviction_strategy"]; ok {
		_ = resourceData.Set("eviction_strategy", v)
	}
	if v, ok := VMTemplateSpecAttributes["termination_grace_period_seconds"]; ok {
		_ = resourceData.Set("termination_grace_period_seconds", v)
	}
	if v, ok := VMTemplateSpecAttributes["liveness_probe"]; ok {
		_ = resourceData.Set("liveness_probe", v)
	}
	if v, ok := VMTemplateSpecAttributes["readiness_probe"]; ok {
		_ = resourceData.Set("readiness_probe", v)
	}
	if v, ok := VMTemplateSpecAttributes["pod_dns_config"]; ok {
		_ = resourceData.Set("pod_dns_config", v)
	}
	if err := resourceData.Set("data_volume_templates", VM["data_volume_templates"]); err != nil {
		return err
	}
	return nil
}

// func FlattenVMMToSpectroSchema(in kubevirtapiv1.VirtualMachineSpec, resourceData *schema.ResourceData) error {
// 	VM := FlattenVirtualMachineSpec(in, resourceData)[0].(map[string]interface{})
// 	// template spec
// 	VMTemplate := VM["template"]
// 	VMTemplateSpec := VMTemplate.([]interface{})[0].(map[string]interface{})["spec"]
// 	VMTemplateSpecAttributes := VMTemplateSpec.([]interface{})[0].(map[string]interface{})

// 	// domain spec
// 	vmDomain := VMTemplateSpecAttributes["domain"].([]interface{})[0].(map[string]interface{})
// 	vmTolerations := VMTemplateSpecAttributes["tolerations"]
// 	resource := vmDomain["resources"]
// 	cpu := vmDomain["cpu"]
// 	memory := vmDomain["memory"]
// 	firmware := vmDomain["firmware"]
// 	features := vmDomain["features"]
// 	device := vmDomain["devices"].([]interface{})[0].(map[string]interface{})
// 	disks := device["disk"]
// 	interfaces := device["interface"]

// 	// Not checking key exist for all required attributes, this will be revamped. if needed.

// 	if err := resourceData.Set("run_strategy", VM["run_strategy"]); err != nil {
// 		return err
// 	}
// 	if err := resourceData.Set("node_selector", VMTemplateSpecAttributes["node_selector"]); err != nil {
// 		return err
// 	}
// 	if err := resourceData.Set("affinity", VMTemplateSpecAttributes["affinity"]); err != nil {
// 		return err
// 	}
// 	if err := resourceData.Set("scheduler_name", VMTemplateSpecAttributes["scheduler_name"]); err != nil {
// 		return err
// 	}
// 	if err := resourceData.Set("hostname", VMTemplateSpecAttributes["hostname"]); err != nil {
// 		return err
// 	}
// 	if err := resourceData.Set("subdomain", VMTemplateSpecAttributes["subdomain"]); err != nil {
// 		return err
// 	}
// 	if err := resourceData.Set("dns_policy", VMTemplateSpecAttributes["dns_policy"]); err != nil {
// 		return err
// 	}
// 	if err := resourceData.Set("priority_class_name", VMTemplateSpecAttributes["priority_class_name"]); err != nil {
// 		return err
// 	}
// 	if err := resourceData.Set("network", VMTemplateSpecAttributes["network"]); err != nil {
// 		return err
// 	}
// 	if err := resourceData.Set("volume", VMTemplateSpecAttributes["volume"]); err != nil {
// 		return err
// 	}
// 	if err := resourceData.Set("cpu", cpu); err != nil {
// 		return err
// 	}
// 	if err := resourceData.Set("memory", memory); err != nil {
// 		return err
// 	}
// 	if err := resourceData.Set("firmware", firmware); err != nil {
// 		return err
// 	}
// 	if err := resourceData.Set("features", features); err != nil {
// 		return err
// 	}
// 	if err := resourceData.Set("resources", resource); err != nil {
// 		return err
// 	}
// 	if err := resourceData.Set("disk", disks); err != nil {
// 		return err
// 	}
// 	if err := resourceData.Set("interface", interfaces); err != nil {
// 		return err
// 	}
// 	if err := resourceData.Set("tolerations", vmTolerations); err != nil {
// 		return err
// 	}

// 	// checking key exist for all optional attributes
// 	if v, ok := VMTemplateSpecAttributes["eviction_strategy"]; !ok {
// 		if err := resourceData.Set("eviction_strategy", v); err != nil {
// 			return err
// 		}
// 	}
// 	if v, ok := VMTemplateSpecAttributes["termination_grace_period_seconds"]; !ok {
// 		if err := resourceData.Set("termination_grace_period_seconds", v); err != nil {
// 			return err
// 		}
// 	}
// 	if v, ok := VMTemplateSpecAttributes["liveness_probe"]; !ok {
// 		if err := resourceData.Set("liveness_probe", v); err != nil {
// 			return err
// 		}
// 	}
// 	if v, ok := VMTemplateSpecAttributes["readiness_probe"]; !ok {
// 		if err := resourceData.Set("readiness_probe", v); err != nil {
// 			return err
// 		}
// 	}
// 	if v, ok := VMTemplateSpecAttributes["pod_dns_config"]; !ok {
// 		if err := resourceData.Set("pod_dns_config", v); err != nil {
// 			return err
// 		}
// 	}

// 	if err := resourceData.Set("data_volume_templates", VM["data_volume_templates"]); err != nil {
// 		return err
// 	}
// 	return nil
// }

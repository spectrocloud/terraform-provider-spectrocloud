package virtualmachine

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	kubevirtapiv1 "kubevirt.io/api/core/v1"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/schema/datavolume"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/schema/k8s"
)

func DataVolumeFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"metadata": k8s.NamespacedMetadataSchema("DataVolume", false),
		"spec":     datavolume.DataVolumeSpecSchema(),
	}
}

func dataVolumeTemplatesSchema() *schema.Schema {
	fields := DataVolumeFields()

	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "dataVolumeTemplates is a list of dataVolumes that the VirtualMachineInstance template can reference.",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: fields,
		},
	}
}

func expandDataVolumeTemplates(dataVolumes []interface{}) ([]kubevirtapiv1.DataVolumeTemplateSpec, error) {
	result := make([]kubevirtapiv1.DataVolumeTemplateSpec, len(dataVolumes))

	if len(dataVolumes) == 0 || dataVolumes[0] == nil {
		return result, nil
	}

	for i, dataVolume := range dataVolumes {
		in := dataVolume.(map[string]interface{})

		if v, ok := in["metadata"].([]interface{}); ok {
			result[i].ObjectMeta = k8s.ExpandMetadata(v)
		}
		if v, ok := in["spec"].([]interface{}); ok {
			spec, err := datavolume.ExpandDataVolumeSpec(v)
			if err != nil {
				return result, err
			}
			result[i].Spec = spec
		}
	}

	return result, nil
}

func flattenDataVolumeTemplates(in []kubevirtapiv1.DataVolumeTemplateSpec, resourceData *schema.ResourceData) []interface{} {
	att := make([]interface{}, len(in))

	for i, v := range in {
		c := make(map[string]interface{})
		c["metadata"] = k8s.FlattenMetadataDataVolume(v.ObjectMeta)
		c["spec"] = datavolume.FlattenDataVolumeSpec(v.Spec)
		att[i] = c
	}

	return att
}

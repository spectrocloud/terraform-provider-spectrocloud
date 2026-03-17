package virtualmachine

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spectrocloud/palette-sdk-go/api/models"
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

func expandDataVolumeTemplates(dataVolumes []interface{}) ([]*models.V1VMDataVolumeTemplateSpec, error) {
	result := make([]*models.V1VMDataVolumeTemplateSpec, len(dataVolumes))

	if len(dataVolumes) == 0 || dataVolumes[0] == nil {
		return result, nil
	}

	for i, dataVolume := range dataVolumes {
		in := dataVolume.(map[string]interface{})
		item := &models.V1VMDataVolumeTemplateSpec{}
		if v, ok := in["metadata"].([]interface{}); ok {
			item.Metadata = k8s.ExpandMetadata(v)
		}
		if v, ok := in["spec"].([]interface{}); ok {
			spec, err := datavolume.ExpandDataVolumeSpec(v)
			if err != nil {
				return result, err
			}
			item.Spec = spec
		}
		result[i] = item
	}

	return result, nil
}

// flattenDataVolumeTemplatesK8s flattens KubeVirt []DataVolumeTemplateSpec to the same schema shape as flattenDataVolumeTemplates.
// func flattenDataVolumeTemplatesK8s(in []kubevirtapiv1.DataVolumeTemplateSpec, resourceData *schema.ResourceData) []interface{} {
// 	att := make([]interface{}, len(in))
// 	for i, v := range in {
// 		c := make(map[string]interface{})
// 		c["metadata"] = k8s.FlattenMetadataDataVolume(v.ObjectMeta)
// 		c["spec"] = datavolume.FlattenDataVolumeSpec(v.Spec)
// 		att[i] = c
// 	}
// 	return att
// }

// flattenDataVolumeTemplatesFromVM flattens []*V1VMDataVolumeTemplateSpec (API response) to the same shape as flattenDataVolumeTemplates.
func flattenDataVolumeTemplatesFromVM(in []*models.V1VMDataVolumeTemplateSpec, resourceData *schema.ResourceData) []interface{} {
	if len(in) == 0 {
		return []interface{}{}
	}
	att := make([]interface{}, 0, len(in))
	for _, v := range in {
		if v == nil {
			continue
		}
		c := make(map[string]interface{})
		c["metadata"] = k8s.FlattenMetadataDataVolumeFromVM(v.Metadata)
		c["spec"] = datavolume.FlattenDataVolumeSpecFromVM(v.Spec)
		att = append(att, c)
	}
	return att
}

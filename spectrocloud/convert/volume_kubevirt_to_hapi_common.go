package convert

import (
	"math"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/schema/datavolume"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/utils"
)

// ToHapiVolume builds a Palette VM add-volume request from Terraform resource data
// without converting through CDI (cdiv1.DataVolume) types.
func ToHapiVolume(d *schema.ResourceData, addVolumeOptions *models.V1VMAddVolumeOptions) (*models.V1VMAddVolumeEntity, error) {
	meta := expandDataVolumeMetadataToVM(d)
	spec, err := datavolume.ExpandDataVolumeSpec(d.Get("spec").([]interface{}))
	if err != nil {
		return nil, err
	}

	return &models.V1VMAddVolumeEntity{
		AddVolumeOptions: addVolumeOptions,
		DataVolumeTemplate: &models.V1VMDataVolumeTemplateSpec{
			Metadata: meta,
			Spec:     spec,
		},
		Persist: true,
	}, nil
}

// expandDataVolumeMetadataToVM maps the `metadata` block to models.V1VMObjectMeta,
// matching the field coverage of k8s.ExpandMetadataToObjectMeta for user-settable fields
// plus uid/generation when present in state.
func expandDataVolumeMetadataToVM(d *schema.ResourceData) *models.V1VMObjectMeta {
	in := d.Get("metadata").([]interface{})
	if len(in) < 1 || in[0] == nil {
		return &models.V1VMObjectMeta{}
	}
	m := in[0].(map[string]interface{})
	meta := &models.V1VMObjectMeta{}

	if v, ok := m["annotations"].(map[string]string); ok && len(v) > 0 {
		meta.Annotations = v
	} else if v, ok := m["annotations"].(map[string]interface{}); ok && len(v) > 0 {
		meta.Annotations = utils.ExpandStringMap(v)
	}
	if v, ok := m["labels"].(map[string]string); ok && len(v) > 0 {
		meta.Labels = v
	} else if v, ok := m["labels"].(map[string]interface{}); ok && len(v) > 0 {
		meta.Labels = utils.ExpandStringMap(v)
	}
	if v, ok := m["generate_name"]; ok && v != nil && v.(string) != "" {
		meta.GenerateName = v.(string)
	}
	if v, ok := m["name"]; ok && v != nil {
		meta.Name = v.(string)
	}
	if v, ok := m["namespace"]; ok && v != nil {
		meta.Namespace = v.(string)
	}
	if v, ok := m["resource_version"]; ok && v != nil {
		meta.ResourceVersion = v.(string)
	}
	if v, ok := m["uid"].(string); ok && v != "" {
		meta.UID = v
	}
	if v, ok := m["generation"]; ok && v != nil {
		meta.Generation = expandInt64FromInterface(v)
	}
	return meta
}

func expandInt64FromInterface(v interface{}) int64 {
	switch g := v.(type) {
	case int:
		return int64(g)
	case int64:
		return g
	case int32:
		return int64(g)
	case uint64:
		if g > uint64(math.MaxInt64) {
			return math.MaxInt64
		}
		return int64(g)
	default:
		return 0
	}
}

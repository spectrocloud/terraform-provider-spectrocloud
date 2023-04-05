package k8s

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	api "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/utils"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

func persistentVolumeClaimSpecFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"access_modes": {
			Type:        schema.TypeSet,
			Description: "A set of the desired access modes the volume should have. More info: http://kubernetes.io/docs/user-guide/persistent-volumes#access-modes-1",
			Required:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{
					"ReadWriteOnce",
					"ReadOnlyMany",
					"ReadWriteMany",
				}, false),
			},
			Set: schema.HashString,
		},
		"resources": {
			Type:        schema.TypeList,
			Description: "A list of the minimum resources the volume should have. More info: http://kubernetes.io/docs/user-guide/persistent-volumes#resources",
			Required:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"limits": {
						Type:        schema.TypeMap,
						Description: "Map describing the maximum amount of compute resources allowed. More info: http://kubernetes.io/docs/user-guide/compute-resources/",
						Optional:    true,
						ForceNew:    true,
					},
					// This is the only field the API will allow modifying in-place, so ForceNew is not used.
					"requests": {
						Type:        schema.TypeMap,
						Description: "Map describing the minimum amount of compute resources required. If this is omitted for a container, it defaults to `limits` if that is explicitly specified, otherwise to an implementation-defined value. More info: http://kubernetes.io/docs/user-guide/compute-resources/",
						Optional:    true,
					},
				},
			},
		},
		"selector": {
			Type:        schema.TypeList,
			Description: "A label query over volumes to consider for binding.",
			Optional:    true,
			ForceNew:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: labelSelectorFields(false),
			},
		},
		"volume_name": {
			Type:        schema.TypeString,
			Description: "The binding reference to the PersistentVolume backing this claim.",
			Optional:    true,
			ForceNew:    true,
			Computed:    true,
		},
		"storage_class_name": {
			Type:        schema.TypeString,
			Description: "Name of the storage class requested by the claim",
			Optional:    true,
			Computed:    true,
		},
		"volume_mode": {
			Type:        schema.TypeString,
			Description: "volumeMode defines what type of volume is required by the claim. Value of Filesystem is implied when not included in claim spec.",
			Optional:    true,
			Computed:    true,
		},
	}
}

func PersistentVolumeClaimSpecSchema() *schema.Schema {
	fields := persistentVolumeClaimSpecFields()

	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "PVC is a pointer to the PVC Spec we want to use.",
		Required:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: fields,
		},
	}

}

// Flatteners

func FlattenPersistentVolumeClaimSpec(in v1.PersistentVolumeClaimSpec) []interface{} {
	att := make(map[string]interface{})
	att["access_modes"] = flattenPersistentVolumeAccessModes(in.AccessModes)
	att["resources"] = flattenResourceRequirements(in.Resources)
	if in.Selector != nil {
		att["selector"] = flattenLabelSelector(in.Selector)
	}
	if in.VolumeName != "" {
		att["volume_name"] = in.VolumeName
	}
	if in.StorageClassName != nil {
		att["storage_class_name"] = *in.StorageClassName
	}
	if in.VolumeMode != nil {
		att["volume_mode"] = *in.VolumeMode
	}
	return []interface{}{att}
}

func flattenResourceRequirements(in v1.ResourceRequirements) []interface{} {
	att := make(map[string]interface{})
	if len(in.Limits) > 0 {
		att["limits"] = utils.FlattenStringMap(utils.FlattenResourceList(in.Limits))
	}
	if len(in.Requests) > 0 {
		att["requests"] = utils.FlattenStringMap(utils.FlattenResourceList(in.Requests))
	}
	return []interface{}{att}
}

// Expanders

func ExpandPersistentVolumeClaimSpec(l []interface{}) (*v1.PersistentVolumeClaimSpec, error) {
	obj := &v1.PersistentVolumeClaimSpec{}
	if len(l) == 0 || l[0] == nil {
		return obj, nil
	}
	in := l[0].(map[string]interface{})
	resourceRequirements, err := expandResourceRequirements(in["resources"].([]interface{}))
	if err != nil {
		return nil, err
	}
	obj.AccessModes = expandPersistentVolumeAccessModes(in["access_modes"].(*schema.Set).List())
	obj.Resources = *resourceRequirements
	if v, ok := in["selector"].([]interface{}); ok && len(v) > 0 {
		obj.Selector = expandLabelSelector(v)
	}
	if v, ok := in["volume_name"].(string); ok {
		obj.VolumeName = v
	}
	if v, ok := in["storage_class_name"].(string); ok && v != "" {
		obj.StorageClassName = utils.PtrToString(v)
	}
	if v, ok := in["volume_mode"].(string); ok && v != "" {
		switch v {
		case string(v1.PersistentVolumeBlock):
			obj.VolumeMode = types.Ptr(v1.PersistentVolumeBlock)
		case string(v1.PersistentVolumeFilesystem):
			obj.VolumeMode = types.Ptr(v1.PersistentVolumeFilesystem)
		default:
			return nil, fmt.Errorf("invalid volume mode: %s", v)
		}
	}
	return obj, nil
}

func expandResourceRequirements(l []interface{}) (*v1.ResourceRequirements, error) {
	obj := &v1.ResourceRequirements{}
	if len(l) == 0 || l[0] == nil {
		return obj, nil
	}
	in := l[0].(map[string]interface{})
	if v, ok := in["limits"].(map[string]interface{}); ok && len(v) > 0 {
		rl, err := utils.ExpandMapToResourceList(v)
		if err != nil {
			return obj, err
		}
		obj.Limits = *rl
	}
	if v, ok := in["requests"].(map[string]interface{}); ok && len(v) > 0 {
		rq, err := utils.ExpandMapToResourceList(v)
		if err != nil {
			return obj, err
		}
		obj.Requests = *rq
	}
	return obj, nil
}

func expandPersistentVolumeAccessModes(s []interface{}) []api.PersistentVolumeAccessMode {
	out := make([]api.PersistentVolumeAccessMode, len(s))
	for i, v := range s {
		out[i] = api.PersistentVolumeAccessMode(v.(string))
	}
	return out
}

func flattenPersistentVolumeAccessModes(in []api.PersistentVolumeAccessMode) *schema.Set {
	var out = make([]interface{}, len(in))
	for i, v := range in {
		out[i] = string(v)
	}
	return schema.NewSet(schema.HashString, out)
}

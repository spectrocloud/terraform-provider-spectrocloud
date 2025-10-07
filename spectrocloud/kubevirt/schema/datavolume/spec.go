package datavolume

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cdiv1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/schema/k8s"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/utils"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

func dataVolumeSpecFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"source":  dataVolumeSourceSchema(),
		"pvc":     k8s.PersistentVolumeClaimSpecSchema(),
		"storage": dataVolumeStorageSchema(),
		"content_type": {
			Type:        schema.TypeString,
			Description: "ContentType options: \"kubevirt\", \"archive\".",
			Optional:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"kubevirt",
				"archive",
			}, false),
		},
	}
}

func dataVolumeStorageSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "Storage is the requested storage specification for the DataVolume.",
		Optional:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"access_modes": {
					Type:        schema.TypeSet,
					Description: "A set of the desired access modes the volume should have. More info: http://kubernetes.io/docs/concepts/storage/persistent-volumes#access-modes-1",
					Optional:    true,
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
					Description: "A list of the minimum resources the volume should have. More info: http://kubernetes.io/docs/concepts/storage/persistent-volumes#resources",
					Optional:    true,
					MaxItems:    1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"limits": {
								Type:        schema.TypeMap,
								Description: "Map describing the maximum amount of compute resources allowed. More info: http://kubernetes.io/docs/user-guide/compute-resources/",
								Optional:    true,
							},
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
					MaxItems:    1,
					Elem: &schema.Resource{
						Schema: labelSelectorFields(),
					},
				},
				"volume_name": {
					Type:        schema.TypeString,
					Description: "The binding reference to the PersistentVolume backing this claim.",
					Optional:    true,
				},
				"storage_class_name": {
					Type:        schema.TypeString,
					Description: "Name of the storage class requested by the claim",
					Optional:    true,
				},
				"volume_mode": {
					Type:        schema.TypeString,
					Description: "volumeMode defines what type of volume is required by the claim. Value of Filesystem is implied when not included in claim spec.",
					Optional:    true,
					ValidateFunc: validation.StringInSlice([]string{
						"Block",
						"Filesystem",
					}, false),
				},
			},
		},
	}
}

func labelSelectorFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"match_expressions": {
			Type:        schema.TypeList,
			Description: "A list of label selector requirements. The requirements are ANDed.",
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"key": {
						Type:        schema.TypeString,
						Description: "The label key that the selector applies to.",
						Optional:    true,
					},
					"operator": {
						Type:        schema.TypeString,
						Description: "A key's relationship to a set of values. Valid operators are `In`, `NotIn`, `Exists` and `DoesNotExist`.",
						Optional:    true,
					},
					"values": {
						Type:        schema.TypeSet,
						Description: "An array of string values. If the operator is `In` or `NotIn`, the values array must be non-empty. If the operator is `Exists` or `DoesNotExist`, the values array must be empty.",
						Optional:    true,
						Elem:        &schema.Schema{Type: schema.TypeString},
						Set:         schema.HashString,
					},
				},
			},
		},
		"match_labels": {
			Type:        schema.TypeMap,
			Description: "A map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of `match_expressions`, whose key field is \"key\", the operator is \"In\", and the values array contains only \"value\".",
			Optional:    true,
		},
	}
}

func DataVolumeSpecSchema() *schema.Schema {
	fields := dataVolumeSpecFields()

	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "DataVolumeSpec defines our specification for a DataVolume type",
		Required:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: fields,
		},
	}

}

func ExpandDataVolumeSpec(dataVolumeSpec []interface{}) (cdiv1.DataVolumeSpec, error) {
	result := cdiv1.DataVolumeSpec{}

	if len(dataVolumeSpec) == 0 || dataVolumeSpec[0] == nil {
		return result, nil
	}

	in := dataVolumeSpec[0].(map[string]interface{})

	result.Source = expandDataVolumeSource(in["source"].([]interface{}))

	if v, ok := in["pvc"].([]interface{}); ok && len(v) > 0 {
		p, err := k8s.ExpandPersistentVolumeClaimSpec(v)
		if err != nil {
			return result, err
		}
		result.PVC = p
	}

	if v, ok := in["storage"].([]interface{}); ok && len(v) > 0 {
		storage, err := expandDataVolumeStorage(v)
		if err != nil {
			return result, err
		}
		result.Storage = storage
	}

	if v, ok := in["content_type"].(string); ok {
		result.ContentType = cdiv1.DataVolumeContentType(v)
	}

	return result, nil
}

func expandDataVolumeStorage(storage []interface{}) (*cdiv1.StorageSpec, error) {
	if len(storage) == 0 || storage[0] == nil {
		return nil, nil
	}

	in := storage[0].(map[string]interface{})
	result := &cdiv1.StorageSpec{}

	if v, ok := in["access_modes"].(*schema.Set); ok && v.Len() > 0 {
		result.AccessModes = expandPersistentVolumeAccessModes(v.List())
	}

	if v, ok := in["resources"].([]interface{}); ok && len(v) > 0 {
		resources, err := expandResourceRequirements(v)
		if err != nil {
			return nil, err
		}
		result.Resources = v1.VolumeResourceRequirements{
			Limits:   resources.Limits,
			Requests: resources.Requests,
		}
	}

	if v, ok := in["selector"].([]interface{}); ok && len(v) > 0 {
		result.Selector = expandLabelSelector(v)
	}

	if v, ok := in["volume_name"].(string); ok && v != "" {
		result.VolumeName = v
	}

	if v, ok := in["storage_class_name"].(string); ok && v != "" {
		result.StorageClassName = &v
	}

	if v, ok := in["volume_mode"].(string); ok && v != "" {
		switch v {
		case "Block":
			result.VolumeMode = types.Ptr(v1.PersistentVolumeBlock)
		case "Filesystem":
			result.VolumeMode = types.Ptr(v1.PersistentVolumeFilesystem)
		}
	}

	return result, nil
}

func expandPersistentVolumeAccessModes(s []interface{}) []v1.PersistentVolumeAccessMode {
	out := make([]v1.PersistentVolumeAccessMode, len(s))
	for i, v := range s {
		out[i] = v1.PersistentVolumeAccessMode(v.(string))
	}
	return out
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

func expandLabelSelector(l []interface{}) *metav1.LabelSelector {
	if len(l) == 0 || l[0] == nil {
		return &metav1.LabelSelector{}
	}
	in := l[0].(map[string]interface{})
	obj := &metav1.LabelSelector{}
	if v, ok := in["match_labels"].(map[string]interface{}); ok && len(v) > 0 {
		obj.MatchLabels = utils.ExpandStringMap(v)
	}
	if v, ok := in["match_expressions"].([]interface{}); ok && len(v) > 0 {
		obj.MatchExpressions = expandLabelSelectorRequirement(v)
	}
	return obj
}

func expandLabelSelectorRequirement(l []interface{}) []metav1.LabelSelectorRequirement {
	if len(l) == 0 || l[0] == nil {
		return []metav1.LabelSelectorRequirement{}
	}
	obj := make([]metav1.LabelSelectorRequirement, len(l))
	for i, n := range l {
		in := n.(map[string]interface{})
		obj[i] = metav1.LabelSelectorRequirement{
			Key:      in["key"].(string),
			Operator: metav1.LabelSelectorOperator(in["operator"].(string)),
			Values:   utils.SliceOfString(in["values"].(*schema.Set).List()),
		}
	}
	return obj
}

func FlattenDataVolumeSpec(spec cdiv1.DataVolumeSpec) []interface{} {
	att := map[string]interface{}{
		"source":       flattenDataVolumeSource(spec.Source),
		"content_type": string(spec.ContentType),
	}

	if spec.PVC != nil {
		att["pvc"] = k8s.FlattenPersistentVolumeClaimSpec(*spec.PVC)
	}

	if spec.Storage != nil {
		att["storage"] = flattenDataVolumeStorage(*spec.Storage)
	}

	return []interface{}{att}
}

func flattenDataVolumeStorage(in cdiv1.StorageSpec) []interface{} {
	att := make(map[string]interface{})

	if len(in.AccessModes) > 0 {
		att["access_modes"] = flattenPersistentVolumeAccessModes(in.AccessModes)
	}

	if len(in.Resources.Limits) > 0 || len(in.Resources.Requests) > 0 {
		att["resources"] = flattenResourceRequirements(v1.ResourceRequirements{
			Limits:   in.Resources.Limits,
			Requests: in.Resources.Requests,
		})
	}

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
		att["volume_mode"] = string(*in.VolumeMode)
	}

	return []interface{}{att}
}

func flattenPersistentVolumeAccessModes(in []v1.PersistentVolumeAccessMode) *schema.Set {
	var out = make([]interface{}, len(in))
	for i, v := range in {
		out[i] = string(v)
	}
	return schema.NewSet(schema.HashString, out)
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

func flattenLabelSelector(in *metav1.LabelSelector) []interface{} {
	att := make(map[string]interface{})
	if len(in.MatchLabels) > 0 {
		att["match_labels"] = utils.FlattenStringMap(in.MatchLabels)
	}
	if len(in.MatchExpressions) > 0 {
		att["match_expressions"] = flattenLabelSelectorRequirement(in.MatchExpressions)
	}
	return []interface{}{att}
}

func flattenLabelSelectorRequirement(in []metav1.LabelSelectorRequirement) []interface{} {
	att := make([]interface{}, len(in))
	for i, n := range in {
		m := make(map[string]interface{})
		m["key"] = n.Key
		m["operator"] = n.Operator
		m["values"] = utils.NewStringSet(schema.HashString, n.Values)
		att[i] = m
	}
	return att
}

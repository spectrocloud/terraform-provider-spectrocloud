package datavolume

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cdiv1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"

	"github.com/spectrocloud/palette-sdk-go/api/models"
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

func ExpandDataVolumeSpec(dataVolumeSpec []interface{}) (*models.V1VMDataVolumeSpec, error) {
	result := &models.V1VMDataVolumeSpec{}

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
		result.Pvc = pvcSpecK8sToModel(p)
	}

	if v, ok := in["storage"].([]interface{}); ok && len(v) > 0 {
		storage, err := expandDataVolumeStorage(v)
		if err != nil {
			return result, err
		}
		result.Storage = storage
	}

	if v, ok := in["content_type"].(string); ok {
		result.ContentType = v
	}

	return result, nil
}

// ExpandDataVolumeSpecToK8s expands the spec schema into CDI cdiv1.DataVolumeSpec for native DataVolume resources.
// func ExpandDataVolumeSpecToK8s(dataVolumeSpec []interface{}) (cdiv1.DataVolumeSpec, error) {
// 	var result cdiv1.DataVolumeSpec
// 	if len(dataVolumeSpec) == 0 || dataVolumeSpec[0] == nil {
// 		return result, nil
// 	}
// 	in := dataVolumeSpec[0].(map[string]interface{})

// 	if v, ok := in["source"].([]interface{}); ok {
// 		result.Source = expandDataVolumeSourceToK8s(v)
// 	}
// 	if v, ok := in["pvc"].([]interface{}); ok && len(v) > 0 {
// 		p, err := k8s.ExpandPersistentVolumeClaimSpec(v)
// 		if err != nil {
// 			return result, err
// 		}
// 		result.PVC = p
// 	}
// 	if v, ok := in["storage"].([]interface{}); ok && len(v) > 0 {
// 		storage, err := expandDataVolumeStorageToK8s(v)
// 		if err != nil {
// 			return result, err
// 		}
// 		result.Storage = storage
// 	}
// 	if v, ok := in["content_type"].(string); ok {
// 		result.ContentType = cdiv1.DataVolumeContentType(v)
// 	}
// 	return result, nil
// }

// ExpandDataVolumeSpecToK8s expands the spec schema into CDI cdiv1.DataVolumeSpec for native DataVolume resources.
func ExpandDataVolumeSpecToK8s(dataVolumeSpec []interface{}) (*models.V1VMDataVolumeSpec, error) {
	var result *models.V1VMDataVolumeSpec
	if len(dataVolumeSpec) == 0 || dataVolumeSpec[0] == nil {
		return result, nil
	}
	in := dataVolumeSpec[0].(map[string]interface{})

	if v, ok := in["source"].([]interface{}); ok {
		result.Source = expandDataVolumeSource(v)
	}
	if v, ok := in["pvc"].([]interface{}); ok && len(v) > 0 {
		p, err := k8s.ExpandPersistentVolumeClaimSpec(v)
		if err != nil {
			return result, err
		}
		// result.PVC = p
		result.Pvc = pvcSpecK8sToModel(p)
	}
	if v, ok := in["storage"].([]interface{}); ok && len(v) > 0 {
		storage, err := expandDataVolumeStorageToK8s(v)
		if err != nil {
			return result, err
		}
		result.Storage = storage
	}
	if v, ok := in["content_type"].(string); ok {
		result.ContentType = v
	}
	return result, nil
}

// func expandDataVolumeStorageToK8s(storage []interface{}) (*cdiv1.StorageSpec, error) {
// 	if len(storage) == 0 || storage[0] == nil {
// 		return nil, nil
// 	}
// 	pvc, err := k8s.ExpandPersistentVolumeClaimSpec(storage)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &cdiv1.StorageSpec{
// 		AccessModes:      pvc.AccessModes,
// 		Resources:        pvc.Resources,
// 		Selector:         pvc.Selector,
// 		VolumeName:       pvc.VolumeName,
// 		StorageClassName: pvc.StorageClassName,
// 		VolumeMode:       pvc.VolumeMode,
// 	}, nil
// }

func expandDataVolumeStorageToK8s(storage []interface{}) (*models.V1VMStorageSpec, error) {
	if len(storage) == 0 || storage[0] == nil {
		return nil, nil
	}
	pvc, err := k8s.ExpandPersistentVolumeClaimSpec(storage)
	if err != nil {
		return nil, err
	}
	m := pvcSpecK8sToModel(pvc)
	if m == nil {
		return nil, nil
	}
	return &models.V1VMStorageSpec{
		AccessModes:      m.AccessModes,
		Resources:        m.Resources,
		Selector:         m.Selector,
		VolumeName:       m.VolumeName,
		StorageClassName: m.StorageClassName,
		VolumeMode:       m.VolumeMode,
	}, nil
}

// pvcSpecK8sToModel converts k8s PersistentVolumeClaimSpec to Palette SDK model.
func pvcSpecK8sToModel(p *v1.PersistentVolumeClaimSpec) *models.V1VMPersistentVolumeClaimSpec {
	if p == nil {
		return nil
	}
	out := &models.V1VMPersistentVolumeClaimSpec{}
	for _, m := range p.AccessModes {
		out.AccessModes = append(out.AccessModes, string(m))
	}
	if len(p.Resources.Limits) > 0 || len(p.Resources.Requests) > 0 {
		out.Resources = &models.V1VMCoreResourceRequirements{
			Limits:   make(map[string]models.V1VMQuantity),
			Requests: make(map[string]models.V1VMQuantity),
		}
		for k, q := range p.Resources.Limits {
			out.Resources.Limits[string(k)] = models.V1VMQuantity(q.String())
		}
		for k, q := range p.Resources.Requests {
			out.Resources.Requests[string(k)] = models.V1VMQuantity(q.String())
		}
	}
	if p.Selector != nil {
		out.Selector = labelSelectorK8sToModel(p.Selector)
	}
	if p.StorageClassName != nil {
		out.StorageClassName = *p.StorageClassName
	}
	if p.VolumeMode != nil {
		out.VolumeMode = string(*p.VolumeMode)
	}
	out.VolumeName = p.VolumeName
	return out
}

func labelSelectorK8sToModel(s *metav1.LabelSelector) *models.V1VMLabelSelector {
	if s == nil {
		return nil
	}
	out := &models.V1VMLabelSelector{
		MatchLabels: s.MatchLabels,
	}
	for _, e := range s.MatchExpressions {
		op := string(e.Operator)
		out.MatchExpressions = append(out.MatchExpressions, &models.V1VMLabelSelectorRequirement{
			Key:      types.Ptr(e.Key),
			Operator: types.Ptr(op),
			Values:   e.Values,
		})
	}
	return out
}

func expandDataVolumeStorage(storage []interface{}) (*models.V1VMStorageSpec, error) {
	if len(storage) == 0 || storage[0] == nil {
		return nil, nil
	}

	in := storage[0].(map[string]interface{})
	result := &models.V1VMStorageSpec{}

	if v, ok := in["access_modes"].(*schema.Set); ok && v.Len() > 0 {
		result.AccessModes = expandPersistentVolumeAccessModes(v.List())
	}

	if v, ok := in["resources"].([]interface{}); ok && len(v) > 0 {
		resources, err := expandResourceRequirements(v)
		if err != nil {
			return nil, err
		}
		result.Resources = &models.V1VMCoreResourceRequirements{
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
		result.StorageClassName = v
	}

	if v, ok := in["volume_mode"].(string); ok && v != "" {
		switch v {
		case "Block":
			result.VolumeMode = string(v1.PersistentVolumeBlock)
		case "Filesystem":
			result.VolumeMode = string(v1.PersistentVolumeFilesystem)
		}
	}

	return result, nil
}

func expandPersistentVolumeAccessModes(s []interface{}) []string {
	out := make([]string, len(s))
	for i, v := range s {
		out[i] = v.(string)
	}
	return out
}

func resourceListToVMQuantityMap(rl *v1.ResourceList) map[string]models.V1VMQuantity {
	if rl == nil {
		return nil
	}
	out := make(map[string]models.V1VMQuantity, len(*rl))
	for k, q := range *rl {
		out[string(k)] = models.V1VMQuantity(q.String())
	}
	return out
}

func expandResourceRequirements(l []interface{}) (*models.V1VMCoreResourceRequirements, error) {
	obj := &models.V1VMCoreResourceRequirements{}
	if len(l) == 0 || l[0] == nil {
		return obj, nil
	}
	in := l[0].(map[string]interface{})
	if v, ok := in["limits"].(map[string]interface{}); ok && len(v) > 0 {
		rl, err := utils.ExpandMapToResourceList(v)
		if err != nil {
			return obj, err
		}
		obj.Limits = resourceListToVMQuantityMap(rl)
	}
	if v, ok := in["requests"].(map[string]interface{}); ok && len(v) > 0 {
		rq, err := utils.ExpandMapToResourceList(v)
		if err != nil {
			return obj, err
		}
		obj.Requests = resourceListToVMQuantityMap(rq)
	}
	return obj, nil
}

func expandLabelSelector(l []interface{}) *models.V1VMLabelSelector {
	if len(l) == 0 || l[0] == nil {
		return &models.V1VMLabelSelector{}
	}
	in := l[0].(map[string]interface{})
	obj := &models.V1VMLabelSelector{}
	if v, ok := in["match_labels"].(map[string]interface{}); ok && len(v) > 0 {
		obj.MatchLabels = utils.ExpandStringMap(v)
	}
	if v, ok := in["match_expressions"].([]interface{}); ok && len(v) > 0 {
		obj.MatchExpressions = expandLabelSelectorRequirement(v)
	}
	return obj
}

func expandLabelSelectorRequirement(l []interface{}) []*models.V1VMLabelSelectorRequirement {
	if len(l) == 0 || l[0] == nil {
		return []*models.V1VMLabelSelectorRequirement{}
	}
	obj := make([]*models.V1VMLabelSelectorRequirement, len(l))
	for i, n := range l {
		in := n.(map[string]interface{})
		obj[i] = &models.V1VMLabelSelectorRequirement{
			Key:      types.Ptr(in["key"].(string)),
			Operator: types.Ptr(in["operator"].(string)),
			Values:   utils.SliceOfString(in["values"].(*schema.Set).List()),
		}
	}
	return obj
}

// func FlattenDataVolumeSpec(spec cdiv1.DataVolumeSpec) []interface{} {
// 	att := map[string]interface{}{
// 		"source":       flattenDataVolumeSource(spec.Source),
// 		"content_type": string(spec.ContentType),
// 	}

// 	if spec.PVC != nil {
// 		att["pvc"] = k8s.FlattenPersistentVolumeClaimSpec(*spec.PVC)
// 	}

// 	if spec.Storage != nil {
// 		att["storage"] = flattenDataVolumeStorage(*spec.Storage)
// 	}

// 	return []interface{}{att}
// }

// FlattenDataVolumeSpec flattens Palette HAPI models.V1VMDataVolumeSpec to Terraform state.
func FlattenDataVolumeSpec(spec *models.V1VMDataVolumeSpec) []interface{} {
	att := map[string]interface{}{
		"source":       flattenDataVolumeSourceFromVM(spec.Source),
		"content_type": string(spec.ContentType),
	}

	if spec.Pvc != nil {
		att["pvc"] = flattenPVCSpecFromVM(spec.Pvc)
	}

	if spec.Storage != nil {
		att["storage"] = flattenDataVolumeStorageFromVM(spec.Storage)
	}

	return []interface{}{att}
}

// FlattenDataVolumeSpecFromVM flattens Palette V1VMDataVolumeSpec to the same shape as FlattenDataVolumeSpec.
func FlattenDataVolumeSpecFromVM(spec *models.V1VMDataVolumeSpec) []interface{} {
	if spec == nil {
		return []interface{}{map[string]interface{}{
			"source":       []interface{}{},
			"content_type": "",
		}}
	}
	att := map[string]interface{}{
		"source":       flattenDataVolumeSourceFromVM(spec.Source),
		"content_type": spec.ContentType,
	}
	if spec.Pvc != nil {
		att["pvc"] = flattenPVCSpecFromVM(spec.Pvc)
	}
	if spec.Storage != nil {
		att["storage"] = flattenDataVolumeStorageFromVM(spec.Storage)
	}
	return []interface{}{att}
}

func flattenPVCSpecFromVM(in *models.V1VMPersistentVolumeClaimSpec) []interface{} {
	if in == nil {
		return nil
	}
	att := make(map[string]interface{})
	if len(in.AccessModes) > 0 {
		out := make([]interface{}, len(in.AccessModes))
		for i, s := range in.AccessModes {
			out[i] = s
		}
		att["access_modes"] = schema.NewSet(schema.HashString, out)
	}
	if in.Resources != nil && (len(in.Resources.Limits) > 0 || len(in.Resources.Requests) > 0) {
		resAtt := make(map[string]interface{})
		if len(in.Resources.Limits) > 0 {
			resAtt["limits"] = utils.FlattenStringMap(vmQuantityMapToStringMap(in.Resources.Limits))
		}
		if len(in.Resources.Requests) > 0 {
			resAtt["requests"] = utils.FlattenStringMap(vmQuantityMapToStringMap(in.Resources.Requests))
		}
		att["resources"] = []interface{}{resAtt}
	}
	if in.Selector != nil {
		att["selector"] = flattenLabelSelectorFromVM(in.Selector)
	}
	if in.VolumeName != "" {
		att["volume_name"] = in.VolumeName
	}
	if in.StorageClassName != "" {
		att["storage_class_name"] = in.StorageClassName
	}
	if in.VolumeMode != "" {
		att["volume_mode"] = in.VolumeMode
	}
	return []interface{}{att}
}

func vmQuantityMapToStringMap(m map[string]models.V1VMQuantity) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = string(v)
	}
	return out
}

func flattenDataVolumeStorageFromVM(in *models.V1VMStorageSpec) []interface{} {
	if in == nil {
		return nil
	}
	att := make(map[string]interface{})
	if len(in.AccessModes) > 0 {
		out := make([]interface{}, len(in.AccessModes))
		for i, s := range in.AccessModes {
			out[i] = s
		}
		att["access_modes"] = schema.NewSet(schema.HashString, out)
	}
	if in.Resources != nil && (len(in.Resources.Limits) > 0 || len(in.Resources.Requests) > 0) {
		resAtt := make(map[string]interface{})
		if len(in.Resources.Limits) > 0 {
			resAtt["limits"] = utils.FlattenStringMap(vmQuantityMapToStringMap(in.Resources.Limits))
		}
		if len(in.Resources.Requests) > 0 {
			resAtt["requests"] = utils.FlattenStringMap(vmQuantityMapToStringMap(in.Resources.Requests))
		}
		att["resources"] = []interface{}{resAtt}
	}
	if in.Selector != nil {
		att["selector"] = flattenLabelSelectorFromVM(in.Selector)
	}
	if in.VolumeName != "" {
		att["volume_name"] = in.VolumeName
	}
	if in.StorageClassName != "" {
		att["storage_class_name"] = in.StorageClassName
	}
	if in.VolumeMode != "" {
		att["volume_mode"] = in.VolumeMode
	}
	return []interface{}{att}
}

func flattenLabelSelectorFromVM(s *models.V1VMLabelSelector) []interface{} {
	if s == nil {
		return nil
	}
	att := make(map[string]interface{})
	if len(s.MatchLabels) > 0 {
		att["match_labels"] = utils.FlattenStringMap(s.MatchLabels)
	}
	if len(s.MatchExpressions) > 0 {
		att["match_expressions"] = flattenLabelSelectorRequirementFromVM(s.MatchExpressions)
	}
	return []interface{}{att}
}

func flattenLabelSelectorRequirementFromVM(in []*models.V1VMLabelSelectorRequirement) []interface{} {
	if len(in) == 0 {
		return nil
	}
	att := make([]interface{}, len(in))
	for i, n := range in {
		m := make(map[string]interface{})
		if n.Key != nil {
			m["key"] = *n.Key
		}
		if n.Operator != nil {
			m["operator"] = *n.Operator
		}
		m["values"] = utils.NewStringSet(schema.HashString, n.Values)
		att[i] = m
	}
	return att
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

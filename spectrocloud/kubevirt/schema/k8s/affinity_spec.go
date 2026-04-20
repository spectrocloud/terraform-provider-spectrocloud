package k8s

import (
	"math"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	v1 "k8s.io/api/core/v1"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/utils"
)

// safeInt32 converts int to int32 with bounds checking to prevent overflow
func safeInt32(value int) int32 {
	if value > math.MaxInt32 {
		return math.MaxInt32
	}
	if value < math.MinInt32 {
		return math.MinInt32
	}
	return int32(value)
}

func affinityFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"node_affinity": {
			Type:        schema.TypeList,
			Description: "Node affinity scheduling rules for the pod.",
			Optional:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: nodeAffinityFields(),
			},
		},
		"pod_affinity": {
			Type:        schema.TypeList,
			Description: "Inter-pod topological affinity. rules that specify that certain pods should be placed in the same topological domain (e.g. same node, same rack, same zone, same power domain, etc.)",
			Optional:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: podAffinityFields(),
			},
		},
		"pod_anti_affinity": {
			Type:        schema.TypeList,
			Description: "Inter-pod topological affinity. rules that specify that certain pods should be placed in the same topological domain (e.g. same node, same rack, same zone, same power domain, etc.)",
			Optional:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: podAffinityFields(),
			},
		},
	}
}

func AffinitySchema() *schema.Schema {
	fields := affinityFields()

	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		MaxItems:    1,
		Description: "Optional pod scheduling constraints.",
		Elem: &schema.Resource{
			Schema: fields,
		},
	}
}

func nodeAffinityFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"required_during_scheduling_ignored_during_execution": {
			Type:        schema.TypeList,
			Description: "If the affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to a node label update), the system may or may not try to eventually evict the pod from its node.",
			Optional:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: nodeSelectorFields(),
			},
		},
		"preferred_during_scheduling_ignored_during_execution": {
			Type:        schema.TypeList,
			Description: "The scheduler will prefer to schedule pods to nodes that satisfy the affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, RequiredDuringScheduling affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding 'weight' to the sum if the node matches the corresponding MatchExpressions; the node(s) with the highest sum are the most preferred.",
			Optional:    true,
			Elem: &schema.Resource{
				Schema: preferredSchedulingTermFields(),
			},
		},
	}
}

func nodeSelectorFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"node_selector_term": {
			Type:        schema.TypeList,
			Description: "List of node selector terms. The terms are ORed.",
			Optional:    true,
			Elem: &schema.Resource{
				Schema: nodeSelectorRequirementsFields(),
			},
		},
	}
}

func preferredSchedulingTermFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"weight": {
			Type:        schema.TypeInt,
			Description: "weight is in the range 1-100",
			Required:    true,
		},
		"preference": {
			Type:        schema.TypeList,
			Description: "A node selector term, associated with the corresponding weight.",
			Required:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: nodeSelectorRequirementsFields(),
			},
		},
	}
}

func nodeSelectorRequirementsFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"match_expressions": {
			Type:        schema.TypeList,
			Description: "List of node selector requirements. The requirements are ANDed.",
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"key": {
						Type:        schema.TypeString,
						Description: "The label key that the selector applies to.",
						Optional:    true,
					},
					"operator": {
						Type:         schema.TypeString,
						Description:  "Operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.",
						Optional:     true,
						ValidateFunc: validation.StringInSlice([]string{"In", "NotIn", "Exists", "DoesNotExist", "Gt", "Lt"}, false),
					},
					"values": {
						Type:        schema.TypeSet,
						Description: "Values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be interpreted as an integer. This array is replaced during a strategic merge patch.",
						Optional:    true,
						Elem:        &schema.Schema{Type: schema.TypeString},
						Set:         schema.HashString,
					},
				},
			},
		},
	}
}

func podAffinityFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"required_during_scheduling_ignored_during_execution": {
			Type:        schema.TypeList,
			Description: "If the affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to a pod label update), the system may or may not try to eventually evict the pod from its node. When there are multiple elements, the lists of nodes corresponding to each PodAffinityTerm are intersected, i.e. all terms must be satisfied.",
			Optional:    true,
			Elem: &schema.Resource{
				Schema: podAffinityTermFields(),
			},
		},
		"preferred_during_scheduling_ignored_during_execution": {
			Type:        schema.TypeList,
			Description: "The scheduler will prefer to schedule pods to nodes that satisfy the anti-affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, RequiredDuringScheduling anti-affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding 'weight' to the sum if the node matches the corresponding MatchExpressions; the node(s) with the highest sum are the most preferred.",
			Optional:    true,
			Elem: &schema.Resource{
				Schema: weightedPodAffinityTermFields(),
			},
		},
	}
}

func podAffinityTermFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"label_selector": {
			Type:        schema.TypeList,
			Description: "A label query over a set of resources, in this case pods.",
			Optional:    true,
			Elem: &schema.Resource{
				Schema: labelSelectorFields(true),
			},
		},
		"namespaces": {
			Type:        schema.TypeSet,
			Description: "namespaces specifies which namespaces the labelSelector applies to (matches against); null or empty list means 'this pod's namespace'",
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Set:         schema.HashString,
		},
		"topology_key": {
			Type:         schema.TypeString,
			Description:  "empty topology key is interpreted by the scheduler as 'all topologies'",
			Optional:     true,
			ValidateFunc: validation.StringMatch(regexp.MustCompile(`^.+$`), "value cannot be empty"),
		},
	}
}

func weightedPodAffinityTermFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"weight": {
			Type:        schema.TypeInt,
			Description: "weight associated with matching the corresponding podAffinityTerm, in the range 1-100",
			Required:    true,
		},
		"pod_affinity_term": {
			Type:        schema.TypeList,
			Description: "A pod affinity term, associated with the corresponding weight",
			Required:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: podAffinityTermFields(),
			},
		},
	}
}

// Flatteners

func FlattenAffinity(in *v1.Affinity) []interface{} {
	att := make(map[string]interface{})
	if in == nil {
		return nil
	}
	if in.NodeAffinity != nil {
		att["node_affinity"] = flattenNodeAffinity(in.NodeAffinity)
	}
	if in.PodAffinity != nil {
		att["pod_affinity"] = flattenPodAffinity(in.PodAffinity)
	}
	if in.PodAntiAffinity != nil {
		att["pod_anti_affinity"] = flattenPodAntiAffinity(in.PodAntiAffinity)
	}
	if len(att) > 0 {
		return []interface{}{att}
	}
	return []interface{}{}
}

func flattenNodeAffinity(in *v1.NodeAffinity) []interface{} {
	att := make(map[string]interface{})
	if in.RequiredDuringSchedulingIgnoredDuringExecution != nil {
		att["required_during_scheduling_ignored_during_execution"] = flattenNodeSelector(in.RequiredDuringSchedulingIgnoredDuringExecution)
	}
	if in.PreferredDuringSchedulingIgnoredDuringExecution != nil {
		att["preferred_during_scheduling_ignored_during_execution"] = flattenPreferredSchedulingTerm(in.PreferredDuringSchedulingIgnoredDuringExecution)
	}
	if len(att) > 0 {
		return []interface{}{att}
	}
	return []interface{}{}
}

func flattenPodAffinity(in *v1.PodAffinity) []interface{} {
	att := make(map[string]interface{})
	if len(in.RequiredDuringSchedulingIgnoredDuringExecution) > 0 {
		att["required_during_scheduling_ignored_during_execution"] = flattenPodAffinityTerms(in.RequiredDuringSchedulingIgnoredDuringExecution)
	}
	if len(in.PreferredDuringSchedulingIgnoredDuringExecution) > 0 {
		att["preferred_during_scheduling_ignored_during_execution"] = flattenWeightedPodAffinityTerms(in.PreferredDuringSchedulingIgnoredDuringExecution)
	}
	if len(att) > 0 {
		return []interface{}{att}
	}
	return []interface{}{}
}

func flattenPodAntiAffinity(in *v1.PodAntiAffinity) []interface{} {
	att := make(map[string]interface{})
	if len(in.RequiredDuringSchedulingIgnoredDuringExecution) > 0 {
		att["required_during_scheduling_ignored_during_execution"] = flattenPodAffinityTerms(in.RequiredDuringSchedulingIgnoredDuringExecution)
	}
	if len(in.PreferredDuringSchedulingIgnoredDuringExecution) > 0 {
		att["preferred_during_scheduling_ignored_during_execution"] = flattenWeightedPodAffinityTerms(in.PreferredDuringSchedulingIgnoredDuringExecution)
	}
	if len(att) > 0 {
		return []interface{}{att}
	}
	return []interface{}{}
}

func flattenNodeSelector(in *v1.NodeSelector) []interface{} {
	att := make(map[string]interface{})
	if len(in.NodeSelectorTerms) > 0 {
		att["node_selector_term"] = flattenNodeSelectorTerms(in.NodeSelectorTerms)
	}
	if len(att) > 0 {
		return []interface{}{att}
	}
	return []interface{}{}
}

func flattenPreferredSchedulingTerm(in []v1.PreferredSchedulingTerm) []interface{} {
	att := make([]interface{}, len(in))
	for i, n := range in {
		m := make(map[string]interface{})
		m["weight"] = int(n.Weight)
		m["preference"] = flattenNodeSelectorTerm(n.Preference)
		att[i] = m
	}
	return att
}

func flattenPodAffinityTerms(in []v1.PodAffinityTerm) []interface{} {
	att := make([]interface{}, len(in))
	for i, n := range in {
		m := make(map[string]interface{})
		m["namespaces"] = utils.NewStringSet(schema.HashString, n.Namespaces)
		m["topology_key"] = n.TopologyKey
		if n.LabelSelector != nil {
			m["label_selector"] = flattenLabelSelector(n.LabelSelector)
		}
		att[i] = m
	}
	return att
}

func flattenWeightedPodAffinityTerms(in []v1.WeightedPodAffinityTerm) []interface{} {
	att := make([]interface{}, len(in))
	for i, n := range in {
		m := make(map[string]interface{})
		m["weight"] = int(n.Weight)
		m["pod_affinity_term"] = flattenPodAffinityTerms([]v1.PodAffinityTerm{n.PodAffinityTerm})
		att[i] = m
	}
	return att
}

func flattenNodeSelectorTerms(in []v1.NodeSelectorTerm) []interface{} {
	att := make([]interface{}, len(in))
	for i, n := range in {
		att[i] = flattenNodeSelectorTerm(n)[0]
	}
	return att
}

func flattenNodeSelectorTerm(in v1.NodeSelectorTerm) []interface{} {
	att := make(map[string]interface{})
	if len(in.MatchExpressions) > 0 {
		att["match_expressions"] = flattenNodeSelectorRequirementList(in.MatchExpressions)
	}
	if len(in.MatchFields) > 0 {
		att["match_fields"] = flattenNodeSelectorRequirementList(in.MatchFields)
	}
	return []interface{}{att}
}

func flattenNodeSelectorRequirementList(in []v1.NodeSelectorRequirement) []interface{} {
	att := make([]interface{}, len(in))
	for i, v := range in {
		m := map[string]interface{}{}
		m["key"] = v.Key
		m["values"] = utils.NewStringSet(schema.HashString, v.Values)
		m["operator"] = string(v.Operator)
		att[i] = m
	}
	return att
}

// FlattenAffinityFromVM flattens *models.V1VMAffinity to the same shape as FlattenAffinity.
func FlattenAffinityFromVM(in *models.V1VMAffinity) []interface{} {
	if in == nil {
		return nil
	}
	att := make(map[string]interface{})
	if in.NodeAffinity != nil {
		if nodeAff := flattenNodeAffinityFromVM(in.NodeAffinity); len(nodeAff) > 0 {
			att["node_affinity"] = nodeAff
		}
	}
	if in.PodAffinity != nil {
		if podAff := flattenPodAffinityFromVM(in.PodAffinity); len(podAff) > 0 {
			att["pod_affinity"] = podAff
		}
	}
	if in.PodAntiAffinity != nil {
		if podAntiAff := flattenPodAntiAffinityFromVM(in.PodAntiAffinity); len(podAntiAff) > 0 {
			att["pod_anti_affinity"] = podAntiAff
		}
	}
	if len(att) > 0 {
		return []interface{}{att}
	}
	return []interface{}{}
}

func flattenNodeAffinityFromVM(in *models.V1VMNodeAffinity) []interface{} {
	if in == nil {
		return nil
	}
	att := make(map[string]interface{})
	if in.RequiredDuringSchedulingIgnoredDuringExecution != nil {
		if v := flattenNodeSelectorFromVM(in.RequiredDuringSchedulingIgnoredDuringExecution); len(v) > 0 {
			att["required_during_scheduling_ignored_during_execution"] = v
		}
	}
	if len(in.PreferredDuringSchedulingIgnoredDuringExecution) > 0 {
		if v := flattenPreferredSchedulingTermsFromVM(in.PreferredDuringSchedulingIgnoredDuringExecution); len(v) > 0 {
			att["preferred_during_scheduling_ignored_during_execution"] = v
		}
	}
	if len(att) > 0 {
		return []interface{}{att}
	}
	return []interface{}{}
}

func flattenNodeSelectorFromVM(in *models.V1VMNodeSelector) []interface{} {
	if in == nil || len(in.NodeSelectorTerms) == 0 {
		return nil
	}
	att := make(map[string]interface{})
	att["node_selector_term"] = flattenNodeSelectorTermsFromVM(in.NodeSelectorTerms)
	return []interface{}{att}
}

func flattenNodeSelectorTermsFromVM(in []*models.V1VMNodeSelectorTerm) []interface{} {
	if len(in) == 0 {
		return nil
	}
	att := make([]interface{}, len(in))
	for i, n := range in {
		if n != nil {
			att[i] = flattenNodeSelectorTermFromVM(n)[0]
		}
	}
	return att
}

func flattenNodeSelectorTermFromVM(in *models.V1VMNodeSelectorTerm) []interface{} {
	if in == nil {
		return []interface{}{map[string]interface{}{}}
	}
	att := make(map[string]interface{})
	if len(in.MatchExpressions) > 0 {
		att["match_expressions"] = flattenNodeSelectorRequirementListFromVM(in.MatchExpressions)
	}
	if len(in.MatchFields) > 0 {
		att["match_fields"] = flattenNodeSelectorRequirementListFromVM(in.MatchFields)
	}
	return []interface{}{att}
}

func flattenNodeSelectorRequirementListFromVM(in []*models.V1VMNodeSelectorRequirement) []interface{} {
	if len(in) == 0 {
		return nil
	}
	att := make([]interface{}, len(in))
	for i, v := range in {
		if v == nil {
			continue
		}
		m := map[string]interface{}{}
		if v.Key != nil {
			m["key"] = *v.Key
		}
		if len(v.Values) > 0 {
			m["values"] = utils.NewStringSet(schema.HashString, v.Values)
		}
		if v.Operator != nil {
			m["operator"] = *v.Operator
		}
		att[i] = m
	}
	return att
}

func flattenPreferredSchedulingTermsFromVM(in []*models.V1VMPreferredSchedulingTerm) []interface{} {
	if len(in) == 0 {
		return nil
	}
	att := make([]interface{}, len(in))
	for i, n := range in {
		if n == nil {
			continue
		}
		m := make(map[string]interface{})
		if n.Weight != nil {
			m["weight"] = int(*n.Weight)
		}
		if n.Preference != nil {
			m["preference"] = flattenNodeSelectorTermFromVM(n.Preference)[0]
		}
		att[i] = m
	}
	return att
}

func flattenPodAffinityFromVM(in *models.V1VMPodAffinity) []interface{} {
	if in == nil {
		return nil
	}
	att := make(map[string]interface{})
	if len(in.RequiredDuringSchedulingIgnoredDuringExecution) > 0 {
		att["required_during_scheduling_ignored_during_execution"] = flattenPodAffinityTermsFromVM(in.RequiredDuringSchedulingIgnoredDuringExecution)
	}
	if len(in.PreferredDuringSchedulingIgnoredDuringExecution) > 0 {
		att["preferred_during_scheduling_ignored_during_execution"] = flattenWeightedPodAffinityTermsFromVM(in.PreferredDuringSchedulingIgnoredDuringExecution)
	}
	if len(att) > 0 {
		return []interface{}{att}
	}
	return []interface{}{}
}

func flattenPodAntiAffinityFromVM(in *models.V1PodAntiAffinity) []interface{} {
	if in == nil {
		return nil
	}
	att := make(map[string]interface{})
	if len(in.RequiredDuringSchedulingIgnoredDuringExecution) > 0 {
		att["required_during_scheduling_ignored_during_execution"] = flattenPodAffinityTermsFromVM(in.RequiredDuringSchedulingIgnoredDuringExecution)
	}
	if len(in.PreferredDuringSchedulingIgnoredDuringExecution) > 0 {
		att["preferred_during_scheduling_ignored_during_execution"] = flattenWeightedPodAffinityTermsFromVM(in.PreferredDuringSchedulingIgnoredDuringExecution)
	}
	if len(att) > 0 {
		return []interface{}{att}
	}
	return []interface{}{}
}

func flattenPodAffinityTermsFromVM(in []*models.V1VMPodAffinityTerm) []interface{} {
	if len(in) == 0 {
		return nil
	}
	att := make([]interface{}, len(in))
	for i, n := range in {
		if n == nil {
			continue
		}
		m := make(map[string]interface{})
		if len(n.Namespaces) > 0 {
			m["namespaces"] = utils.NewStringSet(schema.HashString, n.Namespaces)
		}
		if n.TopologyKey != nil && *n.TopologyKey != "" {
			m["topology_key"] = *n.TopologyKey
		}
		if n.LabelSelector != nil {
			m["label_selector"] = flattenLabelSelectorFromVM(n.LabelSelector)
		}
		att[i] = m
	}
	return att
}

func flattenWeightedPodAffinityTermsFromVM(in []*models.V1VMWeightedPodAffinityTerm) []interface{} {
	if len(in) == 0 {
		return nil
	}
	att := make([]interface{}, len(in))
	for i, n := range in {
		if n == nil {
			continue
		}
		m := make(map[string]interface{})
		if n.Weight != nil {
			m["weight"] = int(*n.Weight)
		}
		if n.PodAffinityTerm != nil {
			m["pod_affinity_term"] = flattenPodAffinityTermsFromVM([]*models.V1VMPodAffinityTerm{n.PodAffinityTerm})
		}
		att[i] = m
	}
	return att
}

func flattenLabelSelectorFromVM(in *models.V1VMLabelSelector) []interface{} {
	if in == nil {
		return nil
	}
	att := make(map[string]interface{})
	if len(in.MatchExpressions) > 0 {
		att["match_expressions"] = flattenLabelSelectorRequirementFromVM(in.MatchExpressions)
	}
	if len(in.MatchLabels) > 0 {
		att["match_labels"] = in.MatchLabels
	}
	return []interface{}{att}
}

func flattenLabelSelectorRequirementFromVM(in []*models.V1VMLabelSelectorRequirement) []interface{} {
	if len(in) == 0 {
		return nil
	}
	att := make([]interface{}, len(in))
	for i, v := range in {
		if v == nil {
			continue
		}
		m := map[string]interface{}{}
		if v.Key != nil {
			m["key"] = *v.Key
		}
		if v.Operator != nil {
			m["operator"] = *v.Operator
		}
		if len(v.Values) > 0 {
			m["values"] = utils.NewStringSet(schema.HashString, v.Values)
		}
		att[i] = m
	}
	return att
}

// Expanders

func ExpandAffinity(a []interface{}) *models.V1VMAffinity {
	if len(a) == 0 || a[0] == nil {
		return &models.V1VMAffinity{}
	}
	in := a[0].(map[string]interface{})
	obj := models.V1VMAffinity{}
	if v, ok := in["node_affinity"].([]interface{}); ok && len(v) > 0 {
		obj.NodeAffinity = expandNodeAffinity(v)
	}
	if v, ok := in["pod_affinity"].([]interface{}); ok && len(v) > 0 {
		obj.PodAffinity = expandPodAffinityToVM(v)
	}
	if v, ok := in["pod_anti_affinity"].([]interface{}); ok && len(v) > 0 {
		obj.PodAntiAffinity = expandPodAntiAffinityToVM(v)
	}
	return &obj
}

func expandNodeAffinity(a []interface{}) *models.V1VMNodeAffinity {
	if len(a) == 0 || a[0] == nil {
		return &models.V1VMNodeAffinity{}
	}
	in := a[0].(map[string]interface{})
	obj := models.V1VMNodeAffinity{}
	if v, ok := in["required_during_scheduling_ignored_during_execution"].([]interface{}); ok && len(v) > 0 {
		obj.RequiredDuringSchedulingIgnoredDuringExecution = expandNodeSelector(v)
	}
	if v, ok := in["preferred_during_scheduling_ignored_during_execution"].([]interface{}); ok && len(v) > 0 {
		obj.PreferredDuringSchedulingIgnoredDuringExecution = expandPreferredSchedulingTermsToVM(v)
	}
	return &obj
}

func expandPreferredSchedulingTermsToVM(t []interface{}) []*models.V1VMPreferredSchedulingTerm {
	if len(t) == 0 || t[0] == nil {
		return nil
	}
	obj := make([]*models.V1VMPreferredSchedulingTerm, len(t))
	for i, n := range t {
		in := n.(map[string]interface{})
		term := &models.V1VMPreferredSchedulingTerm{}
		if v, ok := in["weight"].(int); ok {
			w := safeInt32(v)
			term.Weight = &w
		}
		if v, ok := in["preference"].([]interface{}); ok && len(v) > 0 {
			term.Preference = expandNodeSelectorTerm(v)
		}
		obj[i] = term
	}
	return obj
}

func expandPodAffinityToVM(a []interface{}) *models.V1VMPodAffinity {
	if len(a) == 0 || a[0] == nil {
		return &models.V1VMPodAffinity{}
	}
	in := a[0].(map[string]interface{})
	obj := &models.V1VMPodAffinity{}
	if v, ok := in["required_during_scheduling_ignored_during_execution"].([]interface{}); ok && len(v) > 0 {
		obj.RequiredDuringSchedulingIgnoredDuringExecution = expandPodAffinityTermsToVM(v)
	}
	if v, ok := in["preferred_during_scheduling_ignored_during_execution"].([]interface{}); ok && len(v) > 0 {
		obj.PreferredDuringSchedulingIgnoredDuringExecution = expandWeightedPodAffinityTermsToVM(v)
	}
	return obj
}

func expandPodAntiAffinityToVM(a []interface{}) *models.V1PodAntiAffinity {
	if len(a) == 0 || a[0] == nil {
		return &models.V1PodAntiAffinity{}
	}
	in := a[0].(map[string]interface{})
	obj := &models.V1PodAntiAffinity{}
	if v, ok := in["required_during_scheduling_ignored_during_execution"].([]interface{}); ok && len(v) > 0 {
		obj.RequiredDuringSchedulingIgnoredDuringExecution = expandPodAffinityTermsToVM(v)
	}
	if v, ok := in["preferred_during_scheduling_ignored_during_execution"].([]interface{}); ok && len(v) > 0 {
		obj.PreferredDuringSchedulingIgnoredDuringExecution = expandWeightedPodAffinityTermsToVM(v)
	}
	return obj
}

func expandPodAffinityTermsToVM(t []interface{}) []*models.V1VMPodAffinityTerm {
	if len(t) == 0 || t[0] == nil {
		return nil
	}
	obj := make([]*models.V1VMPodAffinityTerm, len(t))
	for i, n := range t {
		in := n.(map[string]interface{})
		term := &models.V1VMPodAffinityTerm{}
		if v, ok := in["label_selector"].([]interface{}); ok && len(v) > 0 {
			term.LabelSelector = expandLabelSelectorToVM(v)
		}
		if v, ok := in["namespaces"].(*schema.Set); ok {
			term.Namespaces = utils.SliceOfString(v.List())
		}
		if v, ok := in["topology_key"].(string); ok {
			term.TopologyKey = &v
		}
		obj[i] = term
	}
	return obj
}

func expandWeightedPodAffinityTermsToVM(t []interface{}) []*models.V1VMWeightedPodAffinityTerm {
	if len(t) == 0 || t[0] == nil {
		return nil
	}
	obj := make([]*models.V1VMWeightedPodAffinityTerm, len(t))
	for i, n := range t {
		in := n.(map[string]interface{})
		w := &models.V1VMWeightedPodAffinityTerm{}
		if v, ok := in["weight"].(int); ok {
			weight := safeInt32(v)
			w.Weight = &weight
		}
		if v, ok := in["pod_affinity_term"].([]interface{}); ok && len(v) > 0 {
			terms := expandPodAffinityTermsToVM(v)
			if len(terms) > 0 {
				w.PodAffinityTerm = terms[0]
			}
		}
		obj[i] = w
	}
	return obj
}

func expandNodeSelector(s []interface{}) *models.V1VMNodeSelector {
	if len(s) == 0 || s[0] == nil {
		return &models.V1VMNodeSelector{}
	}
	in := s[0].(map[string]interface{})
	obj := models.V1VMNodeSelector{}
	if v, ok := in["node_selector_term"].([]interface{}); ok && len(v) > 0 {
		terms := expandNodeSelectorTerms(v)
		obj.NodeSelectorTerms = make([]*models.V1VMNodeSelectorTerm, len(terms))
		for i := range terms {
			obj.NodeSelectorTerms[i] = &terms[i]
		}
	}
	return &obj
}

func expandNodeSelectorTerms(l []interface{}) []models.V1VMNodeSelectorTerm {
	if len(l) == 0 || l[0] == nil {
		return []models.V1VMNodeSelectorTerm{}
	}
	obj := make([]models.V1VMNodeSelectorTerm, len(l))
	for i, n := range l {
		obj[i] = *expandNodeSelectorTerm([]interface{}{n})
	}
	return obj
}

func expandNodeSelectorTerm(l []interface{}) *models.V1VMNodeSelectorTerm {
	if len(l) == 0 || l[0] == nil {
		return &models.V1VMNodeSelectorTerm{}
	}
	in := l[0].(map[string]interface{})
	obj := models.V1VMNodeSelectorTerm{}
	if v, ok := in["match_expressions"].([]interface{}); ok && len(v) > 0 {
		obj.MatchExpressions = expandNodeSelectorRequirementList(v)
	}
	if v, ok := in["match_fields"].([]interface{}); ok && len(v) > 0 {
		obj.MatchFields = expandNodeSelectorRequirementList(v)
	}
	return &obj
}

func expandNodeSelectorRequirementList(in []interface{}) []*models.V1VMNodeSelectorRequirement {
	att := []*models.V1VMNodeSelectorRequirement{}
	if len(in) < 1 {
		return att
	}
	att = make([]*models.V1VMNodeSelectorRequirement, len(in))
	for i, c := range in {
		p := c.(map[string]interface{})
		att[i] = &models.V1VMNodeSelectorRequirement{
			Key:      utils.PtrToString(p["key"].(string)),
			Operator: utils.PtrToString(p["operator"].(string)),
			Values:   utils.ExpandStringSlice(p["values"].(*schema.Set).List()),
		}
	}
	return att
}

package convert

import (
	"github.com/spectrocloud/hapi/models"
	k8sv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

func ToHapiVmAffinity(affinity *k8sv1.Affinity) *models.V1VMAffinity {
	return &models.V1VMAffinity{
		NodeAffinity:    ToHapiVmNodeAffinity(affinity.NodeAffinity),
		PodAffinity:     ToHapiVmPodAffinity(affinity.PodAffinity),
		PodAntiAffinity: ToHapiVmPodAntiAffinity(affinity.PodAntiAffinity),
	}
}

func ToHapiVmNodeAffinity(affinity *k8sv1.NodeAffinity) *models.V1VMNodeAffinity {
	if affinity == nil {
		return nil
	}

	return &models.V1VMNodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution:  ToHapiVmNodeSelector(affinity.RequiredDuringSchedulingIgnoredDuringExecution),
		PreferredDuringSchedulingIgnoredDuringExecution: ToHapiVmPreferredSchedulingTerms(affinity.PreferredDuringSchedulingIgnoredDuringExecution),
	}
}

func ToHapiVmPreferredSchedulingTerms(execution []k8sv1.PreferredSchedulingTerm) []*models.V1VMPreferredSchedulingTerm {
	ret := make([]*models.V1VMPreferredSchedulingTerm, len(execution))
	for i, term := range execution {
		ret[i] = &models.V1VMPreferredSchedulingTerm{
			Preference: ToHapiVmNodeSelectorTerm(term.Preference),
			Weight:     &term.Weight,
		}
	}
	return ret
}

func ToHapiVmNodeSelectorTerm(preference k8sv1.NodeSelectorTerm) *models.V1VMNodeSelectorTerm {
	return &models.V1VMNodeSelectorTerm{
		MatchExpressions: ToHapiVmNodeSelectorRequirements(preference.MatchExpressions),
		MatchFields:      ToHapiVmNodeSelectorRequirements(preference.MatchFields),
	}
}

func ToHapiVmNodeSelector(execution *k8sv1.NodeSelector) *models.V1VMNodeSelector {
	if execution == nil {
		return nil
	}
	return &models.V1VMNodeSelector{
		NodeSelectorTerms: ToHapiVmNodeSelectorTerms(execution.NodeSelectorTerms),
	}
}

func ToHapiVmNodeSelectorTerms(terms []k8sv1.NodeSelectorTerm) []*models.V1VMNodeSelectorTerm {
	ret := make([]*models.V1VMNodeSelectorTerm, len(terms))
	for i, term := range terms {
		ret[i] = &models.V1VMNodeSelectorTerm{
			MatchExpressions: ToHapiVmNodeSelectorRequirements(term.MatchExpressions),
			MatchFields:      ToHapiVmNodeSelectorRequirements(term.MatchFields),
		}
	}
	return ret
}

func ToHapiVmNodeSelectorRequirements(expressions []k8sv1.NodeSelectorRequirement) []*models.V1VMNodeSelectorRequirement {
	ret := make([]*models.V1VMNodeSelectorRequirement, len(expressions))
	for i, expression := range expressions {
		ret[i] = &models.V1VMNodeSelectorRequirement{
			Key:      &expression.Key,
			Operator: types.Ptr(string(expression.Operator)),
			Values:   expression.Values,
		}
	}
	return ret
}

func ToHapiVmPodAntiAffinity(affinity *k8sv1.PodAntiAffinity) *models.V1PodAntiAffinity {
	if affinity == nil {
		return nil
	}

	return &models.V1PodAntiAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution:  ToHapiVmPodAffinityTerm(affinity.RequiredDuringSchedulingIgnoredDuringExecution),
		PreferredDuringSchedulingIgnoredDuringExecution: ToHapiVmPodAffinityTerms(affinity.PreferredDuringSchedulingIgnoredDuringExecution),
	}
}

func ToHapiVmPodAffinity(affinity *k8sv1.PodAffinity) *models.V1VMPodAffinity {
	if affinity == nil {
		return nil
	}
	return &models.V1VMPodAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution:  ToHapiVmPodAffinityTerm(affinity.RequiredDuringSchedulingIgnoredDuringExecution),
		PreferredDuringSchedulingIgnoredDuringExecution: ToHapiVmPodAffinityTerms(affinity.PreferredDuringSchedulingIgnoredDuringExecution),
	}
}

func ToHapiVmPodAffinityTerms(execution []k8sv1.WeightedPodAffinityTerm) []*models.V1VMWeightedPodAffinityTerm {
	ret := make([]*models.V1VMWeightedPodAffinityTerm, len(execution))
	for i, term := range execution {
		ret[i] = &models.V1VMWeightedPodAffinityTerm{
			Weight: &term.Weight,
			PodAffinityTerm: &models.V1VMPodAffinityTerm{
				LabelSelector:     ToHapiVmLabelSelector(term.PodAffinityTerm.LabelSelector),
				NamespaceSelector: ToHapiVmLabelSelector(term.PodAffinityTerm.NamespaceSelector),
				Namespaces:        term.PodAffinityTerm.Namespaces,
				TopologyKey:       &term.PodAffinityTerm.TopologyKey,
			},
		}
	}
	return ret
}

func ToHapiVmPodAffinityTerm(execution []k8sv1.PodAffinityTerm) []*models.V1VMPodAffinityTerm {
	ret := make([]*models.V1VMPodAffinityTerm, len(execution))
	for i, term := range execution {
		ret[i] = &models.V1VMPodAffinityTerm{
			LabelSelector:     ToHapiVMLabelSelector(term.LabelSelector),
			NamespaceSelector: ToHapiVMLabelSelector(term.NamespaceSelector),
			Namespaces:        term.Namespaces,
			TopologyKey:       &term.TopologyKey,
		}
	}
	return ret
}

func ToHapiVMLabelSelector(selector *metav1.LabelSelector) *models.V1VMLabelSelector {
	return &models.V1VMLabelSelector{
		MatchExpressions: ToHapiVMLabelSelectorRequirement(selector.MatchExpressions),
		MatchLabels:      selector.MatchLabels,
	}
}

func ToHapiVMLabelSelectorRequirement(expressions []metav1.LabelSelectorRequirement) []*models.V1VMLabelSelectorRequirement {
	ret := make([]*models.V1VMLabelSelectorRequirement, len(expressions))
	for i, expression := range expressions {
		ret[i] = &models.V1VMLabelSelectorRequirement{
			Key:      &expression.Key,
			Operator: types.Ptr(string(expression.Operator)),
			Values:   expression.Values,
		}
	}
	return ret
}

func ToHapiVMNodeAffinity(affinity *k8sv1.NodeAffinity) *models.V1VMNodeAffinity {
	return &models.V1VMNodeAffinity{
		PreferredDuringSchedulingIgnoredDuringExecution: ToHapiVMPreferredSchedulingTerm(affinity.PreferredDuringSchedulingIgnoredDuringExecution),
		RequiredDuringSchedulingIgnoredDuringExecution:  ToHapiVMNodeSelector(affinity.RequiredDuringSchedulingIgnoredDuringExecution),
	}
}

func ToHapiVMNodeSelector(execution *k8sv1.NodeSelector) *models.V1VMNodeSelector {
	return &models.V1VMNodeSelector{
		NodeSelectorTerms: ToHapiVMNodeSelectorTerms(execution.NodeSelectorTerms),
	}
}

func ToHapiVMNodeSelectorTerms(terms []k8sv1.NodeSelectorTerm) []*models.V1VMNodeSelectorTerm {
	ret := make([]*models.V1VMNodeSelectorTerm, len(terms))
	for i, term := range terms {
		ret[i] = ToHapiVMNodeSelectorTerm(term)
	}
	return ret
}

func ToHapiVMPreferredSchedulingTerm(execution []k8sv1.PreferredSchedulingTerm) []*models.V1VMPreferredSchedulingTerm {
	ret := make([]*models.V1VMPreferredSchedulingTerm, len(execution))
	for i, term := range execution {
		ret[i] = &models.V1VMPreferredSchedulingTerm{
			Preference: ToHapiVMNodeSelectorTerm(term.Preference),
			Weight:     &term.Weight,
		}
	}
	return ret
}

func ToHapiVMNodeSelectorTerm(preference k8sv1.NodeSelectorTerm) *models.V1VMNodeSelectorTerm {
	return &models.V1VMNodeSelectorTerm{
		MatchFields:      ToHapiVMNodeSelectorRequirement(preference.MatchFields),
		MatchExpressions: ToHapiVMNodeSelectorRequirement(preference.MatchExpressions),
	}
}

func ToHapiVMNodeSelectorRequirement(expressions []k8sv1.NodeSelectorRequirement) []*models.V1VMNodeSelectorRequirement {
	ret := make([]*models.V1VMNodeSelectorRequirement, len(expressions))
	for i, expression := range expressions {
		ret[i] = &models.V1VMNodeSelectorRequirement{
			Key:      &expression.Key,
			Operator: types.Ptr(string(expression.Operator)),
			Values:   expression.Values,
		}
	}
	return ret
}

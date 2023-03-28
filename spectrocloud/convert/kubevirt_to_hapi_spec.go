package convert

import (
	"github.com/spectrocloud/hapi/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubevirtapiv1 "kubevirt.io/api/core/v1"

	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

func ToHapiVmSpec(spec *kubevirtapiv1.VirtualMachineSpec) *models.V1ClusterVirtualMachineSpec {
	var Running bool
	if spec.Running != nil {
		Running = *spec.Running
	}

	Metadata := ToHapiVmObjectMeta(spec.Template.ObjectMeta)

	var Spec *models.V1VMVirtualMachineInstanceSpec
	if spec.Template != nil {
		var TerminationGracePeriodSeconds int64
		if spec.Template.Spec.TerminationGracePeriodSeconds != nil {
			TerminationGracePeriodSeconds = *spec.Template.Spec.TerminationGracePeriodSeconds
		}

		var EvictionStrategy string
		if spec.Template.Spec.EvictionStrategy != nil {
			EvictionStrategy = string(*spec.Template.Spec.EvictionStrategy)
		}

		var StartStrategy string
		if spec.Template.Spec.StartStrategy != nil {
			StartStrategy = string(*spec.Template.Spec.StartStrategy)
		}
		Spec = &models.V1VMVirtualMachineInstanceSpec{
			AccessCredentials:             ToHapiVmAccessCredentials(spec.Template.Spec.AccessCredentials),
			Affinity:                      ToHapiVmAffinity(spec.Template.Spec.Affinity),
			DNSConfig:                     ToHapiVmDNSConfig(spec.Template.Spec.DNSConfig),
			DNSPolicy:                     string(spec.Template.Spec.DNSPolicy),
			Domain:                        ToHapiVmDomain(spec.Template.Spec.Domain),
			EvictionStrategy:              EvictionStrategy,
			Hostname:                      spec.Template.Spec.Hostname,
			LivenessProbe:                 ToHapiVmProbe(spec.Template.Spec.LivenessProbe),
			Networks:                      ToHapiVmNetworks(spec.Template.Spec.Networks),
			NodeSelector:                  spec.Template.Spec.NodeSelector,
			PriorityClassName:             spec.Template.Spec.PriorityClassName,
			ReadinessProbe:                ToHapiVmProbe(spec.Template.Spec.ReadinessProbe),
			SchedulerName:                 spec.Template.Spec.SchedulerName,
			StartStrategy:                 StartStrategy,
			Subdomain:                     spec.Template.Spec.Subdomain,
			TerminationGracePeriodSeconds: TerminationGracePeriodSeconds,
			Tolerations:                   ToHapiVmTolerations(spec.Template.Spec.Tolerations),
			TopologySpreadConstraints:     ToHapiVmTopologySpreadConstraints(spec.Template.Spec.TopologySpreadConstraints),
			Volumes:                       ToHapiVmVolumes(spec.Template.Spec.Volumes),
		}
	}

	var RunStrategy string
	if spec.RunStrategy != nil {
		RunStrategy = string(*spec.RunStrategy)
	}

	return &models.V1ClusterVirtualMachineSpec{
		DataVolumeTemplates: ToHapiVmDataVolumeTemplates(spec.DataVolumeTemplates),
		Instancetype:        nil,
		Preference:          ToHapiVmPreferenceMatcher(spec.Preference),
		RunStrategy:         RunStrategy,
		Running:             Running,
		Template: &models.V1VMVirtualMachineInstanceTemplateSpec{
			Metadata: Metadata,
			Spec:     Spec,
		},
	}
}

func ToHapiVmPreferenceMatcher(preference *kubevirtapiv1.PreferenceMatcher) *models.V1VMPreferenceMatcher {
	if preference == nil {
		return nil
	}
	return &models.V1VMPreferenceMatcher{
		InferFromVolume: preference.InferFromVolume,
		Kind:            preference.Kind,
	}
}

func ToHapiVmObjectMeta(meta metav1.ObjectMeta) *models.V1VMObjectMeta {
	return &models.V1VMObjectMeta{
		Annotations: meta.Annotations,
		//TODO: DeletionGracePeriodSeconds: GracePeriodSeconds,
		Finalizers:      meta.Finalizers,
		GenerateName:    meta.GenerateName,
		Generation:      meta.Generation,
		Labels:          meta.Labels,
		ManagedFields:   ToHapiVMManagedFields(meta.ManagedFields),
		Name:            meta.Name,
		Namespace:       meta.Namespace,
		OwnerReferences: ToHapiVMOwnerReferences(meta.OwnerReferences),
		ResourceVersion: meta.ResourceVersion,
		UID:             string(meta.UID),
	}
}

func ToHapiVMOwnerReferences(references []metav1.OwnerReference) []*models.V1VMOwnerReference {
	var result []*models.V1VMOwnerReference
	for _, reference := range references {
		result = append(result, &models.V1VMOwnerReference{
			APIVersion:         types.Ptr(reference.APIVersion),
			Controller:         *reference.Controller,
			BlockOwnerDeletion: *reference.BlockOwnerDeletion,
			Kind:               types.Ptr(reference.Kind),
			Name:               types.Ptr(reference.Name),
			UID:                types.Ptr(string(reference.UID)),
		})
	}
	return result
}

func ToHapiVMManagedFields(fields []metav1.ManagedFieldsEntry) []*models.V1VMManagedFieldsEntry {
	var result []*models.V1VMManagedFieldsEntry
	for _, field := range fields {
		result = append(result, &models.V1VMManagedFieldsEntry{
			APIVersion: field.APIVersion,
			FieldsType: field.FieldsType,
			FieldsV1:   ToHapiVmFieldsV1(field.FieldsV1),
			Manager:    field.Manager,
			Operation:  string(field.Operation),
			// TODO: Time:       field.Time,
			Subresource: field.Subresource,
		})
	}
	return result
}

func ToHapiVMPreferenceMatcher(preference *kubevirtapiv1.PreferenceMatcher) *models.V1VMPreferenceMatcher {
	return &models.V1VMPreferenceMatcher{
		InferFromVolume: preference.InferFromVolume,
		Kind:            preference.Kind,
		Name:            preference.Name,
		RevisionName:    preference.RevisionName,
	}
}

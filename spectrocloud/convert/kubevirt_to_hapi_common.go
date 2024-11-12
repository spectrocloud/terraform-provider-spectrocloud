package convert

import (
	"encoding/base64"

	"github.com/go-openapi/strfmt"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubevirtapiv1 "kubevirt.io/api/core/v1"

	"github.com/spectrocloud/terraform-provider-spectrocloud/util/ptr"
)

func ToHapiVm(vm *kubevirtapiv1.VirtualMachine) (*models.V1ClusterVirtualMachine, error) {
	var GracePeriodSeconds int64
	if vm.DeletionGracePeriodSeconds != nil {
		GracePeriodSeconds = *vm.DeletionGracePeriodSeconds
	}

	Spec, err := ToHapiVmSpecM(vm.Spec)
	if err != nil {
		return nil, err
	}
	Status, err := ToHapiVmStatusM(vm.Status)
	if err != nil {
		return nil, err
	}

	hapiVM := &models.V1ClusterVirtualMachine{
		Metadata: &models.V1VMObjectMeta{
			Annotations:                vm.Annotations,
			DeletionGracePeriodSeconds: GracePeriodSeconds,
			Finalizers:                 vm.Finalizers,
			GenerateName:               vm.GenerateName,
			Generation:                 vm.Generation,
			Labels:                     vm.Labels,
			ManagedFields:              ToHapiVmManagedFields(vm.ManagedFields),
			Name:                       vm.ObjectMeta.Name,
			Namespace:                  vm.ObjectMeta.Namespace,
			OwnerReferences:            ToHapiVmOwnerReferences(vm.OwnerReferences),
			ResourceVersion:            vm.ResourceVersion,
			UID:                        string(vm.UID),
		},
		Spec:   Spec,
		Status: Status,
	}
	return hapiVM, nil
}

func ToHapiVmOwnerReferences(references []metav1.OwnerReference) []*models.V1VMOwnerReference {
	ret := make([]*models.V1VMOwnerReference, len(references))
	for i, reference := range references {
		ret[i] = &models.V1VMOwnerReference{
			APIVersion:         ptr.To(reference.APIVersion),
			BlockOwnerDeletion: *reference.BlockOwnerDeletion,
			Controller:         *reference.Controller,
			Kind:               ptr.To(reference.Kind),
			Name:               ptr.To(reference.Name),
			UID:                ptr.To(string(reference.UID)),
		}
	}

	return ret
}

func ToHapiVmManagedFields(fields []metav1.ManagedFieldsEntry) []*models.V1VMManagedFieldsEntry {
	ret := make([]*models.V1VMManagedFieldsEntry, len(fields))
	for i, field := range fields {
		ret[i] = &models.V1VMManagedFieldsEntry{
			APIVersion: field.APIVersion,
			FieldsType: field.FieldsType,
			FieldsV1:   ToHapiVmFieldsV1(field.FieldsV1),
			Manager:    field.Manager,
			Operation:  string(field.Operation),
			// TODO: Time:       ToHapiV1Time(field.Time),
		}
	}

	return ret
}

func ToHapiV1Time(t metav1.Time) models.V1Time {
	return models.V1Time(t.Time)
}

func ToHapiVmFieldsV1(v1 *metav1.FieldsV1) *models.V1VMFieldsV1 {
	return &models.V1VMFieldsV1{
		Raw: ByteToStrfmtBase64(v1.Raw),
	}
}

func ByteToStrfmtBase64(raw []byte) []strfmt.Base64 {
	var res []strfmt.Base64
	encoded := base64.StdEncoding.EncodeToString(raw)
	res = append(res, strfmt.Base64(encoded))
	return res
}

func ToHapiVmQuantityDivisor(divisor resource.Quantity) models.V1VMQuantity {
	return models.V1VMQuantity(divisor.String())
}

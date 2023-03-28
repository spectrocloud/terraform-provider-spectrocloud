package convert

import (
	"encoding/base64"

	"github.com/go-openapi/strfmt"
	"github.com/spectrocloud/hapi/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kubevirtapiv1 "kubevirt.io/api/core/v1"
)

func ToKubevirtVM(hapiVM *models.V1ClusterVirtualMachine) *kubevirtapiv1.VirtualMachine {
	if hapiVM == nil {
		return nil
	}

	Spec, _ := ToKubevirtVMSpecM(hapiVM.Spec)
	Status, _ := ToKubevirtVMStatusM(hapiVM.Status)
	kubevirtVM := &kubevirtapiv1.VirtualMachine{
		TypeMeta: metav1.TypeMeta{
			Kind:       hapiVM.Kind,
			APIVersion: hapiVM.APIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:                       hapiVM.Metadata.Name,
			GenerateName:               hapiVM.Metadata.GenerateName,
			Namespace:                  hapiVM.Metadata.Namespace,
			UID:                        types.UID(hapiVM.Metadata.UID),
			ResourceVersion:            hapiVM.Metadata.ResourceVersion,
			Generation:                 hapiVM.Metadata.Generation,
			DeletionGracePeriodSeconds: &hapiVM.Metadata.DeletionGracePeriodSeconds,
			Labels:                     hapiVM.Metadata.Labels,
			Annotations:                hapiVM.Metadata.Annotations,
			OwnerReferences:            ToKubevirtVMOwnerReferences(hapiVM.Metadata.OwnerReferences),
			Finalizers:                 hapiVM.Metadata.Finalizers,
			ManagedFields:              ToKubevirtVMManagedFields(hapiVM.Metadata.ManagedFields),
		},
		Spec:   Spec,
		Status: Status,
	}
	return kubevirtVM
}

func ToKubevirtVMOwnerReferences(references []*models.V1VMOwnerReference) []metav1.OwnerReference {
	ret := make([]metav1.OwnerReference, len(references))
	for i, reference := range references {
		ret[i] = metav1.OwnerReference{
			APIVersion:         *reference.APIVersion,
			BlockOwnerDeletion: &reference.BlockOwnerDeletion,
			Controller:         &reference.Controller,
			Kind:               *reference.Kind,
			Name:               *reference.Name,
			UID:                types.UID(*reference.UID),
		}
	}

	return ret
}

func ToKubevirtVMManagedFields(fields []*models.V1VMManagedFieldsEntry) []metav1.ManagedFieldsEntry {
	ret := make([]metav1.ManagedFieldsEntry, len(fields))
	for i, field := range fields {
		ret[i] = metav1.ManagedFieldsEntry{
			APIVersion: field.APIVersion,
			FieldsType: field.FieldsType,
			FieldsV1:   ToKvVmFieldsV1(field.FieldsV1),
			Manager:    field.Manager,
			Operation:  ToKvVmManagedFieldsOperationType(field.Operation),
			// TODO: Time:       ToKubevirtTime(field.Time),
		}
	}

	return ret
}

func ToKvVmFieldsV1(v1 *models.V1VMFieldsV1) *metav1.FieldsV1 {
	return &metav1.FieldsV1{
		Raw: StrfmtBase64ToByte(v1.Raw),
	}
}

func StrfmtBase64ToByte(raw []strfmt.Base64) []byte {
	var res []byte
	for _, s := range raw {
		decoded, err := base64.StdEncoding.DecodeString(string(s))
		if err != nil {
			// TODO: Handle error
		}
		res = append(res, decoded...)
	}
	return res
}

func ToKvVmManagedFieldsOperationType(operation string) metav1.ManagedFieldsOperationType {
	switch operation {
	case "Apply":
		return metav1.ManagedFieldsOperationApply
	case "Update":
		return metav1.ManagedFieldsOperationUpdate
	}
	return metav1.ManagedFieldsOperationApply
}

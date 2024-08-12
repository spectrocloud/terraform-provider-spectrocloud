package convert

import (
	"encoding/base64"
	"errors"

	"github.com/go-openapi/strfmt"
	"github.com/spectrocloud/palette-api-go/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kubevirtapiv1 "kubevirt.io/api/core/v1"
)

func ToKubevirtVM(hapiVM *models.V1ClusterVirtualMachine) (*kubevirtapiv1.VirtualMachine, error) {
	if hapiVM == nil {
		return nil, errors.New("hapiVM is nil")
	}

	Spec, err := ToKubevirtVMSpecM(hapiVM.Spec)
	if err != nil {
		return nil, err
	}
	Status, err := ToKubevirtVMStatusM(hapiVM.Status)
	if err != nil {
		return nil, err
	}
	OwnerReferences, err := ToKubevirtVMOwnerReferences(hapiVM.Metadata.OwnerReferences)
	if err != nil {
		return nil, err
	}
	ManagedFields, err := ToKubevirtVMManagedFields(hapiVM.Metadata.ManagedFields)
	if err != nil {
		return nil, err
	}
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
			OwnerReferences:            OwnerReferences,
			Finalizers:                 hapiVM.Metadata.Finalizers,
			ManagedFields:              ManagedFields,
		},
		Spec:   Spec,
		Status: Status,
	}
	return kubevirtVM, nil
}

func ToKubevirtVMOwnerReferences(references []*models.V1VMOwnerReference) ([]metav1.OwnerReference, error) {
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

	return ret, nil
}

func ToKubevirtVMManagedFields(fields []*models.V1VMManagedFieldsEntry) ([]metav1.ManagedFieldsEntry, error) {
	ret := make([]metav1.ManagedFieldsEntry, len(fields))
	for i, field := range fields {
		FieldsV1, err := ToKvVmFieldsV1(field.FieldsV1)
		if err != nil {
			return nil, err
		}
		ret[i] = metav1.ManagedFieldsEntry{
			APIVersion: field.APIVersion,
			FieldsType: field.FieldsType,
			FieldsV1:   FieldsV1,
			Manager:    field.Manager,
			Operation:  ToKvVmManagedFieldsOperationType(field.Operation),
			// TODO: Time:       ToKubevirtTime(field.Time),
		}
	}

	return ret, nil
}

func ToKvVmFieldsV1(v1 *models.V1VMFieldsV1) (*metav1.FieldsV1, error) {
	Raw, err := StrfmtBase64ToByte(v1.Raw)
	if err != nil {
		return nil, err
	}
	Fields := &metav1.FieldsV1{
		Raw: Raw,
	}

	return Fields, nil
}

func StrfmtBase64ToByte(raw []strfmt.Base64) ([]byte, error) {
	var res []byte
	for _, s := range raw {
		decoded, err := base64.StdEncoding.DecodeString(string(s))
		if err != nil {
			return nil, err
		}
		res = append(res, decoded...)
	}
	return res, nil
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

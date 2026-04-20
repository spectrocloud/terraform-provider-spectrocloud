package datavolume

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/schema/k8s"
)

func DataVolumeFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"cluster_uid": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The cluster UID to which the virtual machine belongs to.",
		},
		"cluster_context": {
			Type:     schema.TypeString,
			Required: true,
		},
		"vm_name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The name of the virtual machine to which the data volume belongs to.",
		},
		"vm_namespace": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The namespace of the virtual machine to which the data volume belongs to.",
		},
		"add_volume_options": DataVolumeOptionsSchema(),
		"metadata":           k8s.NamespacedMetadataSchema("DataVolume", false),
		"spec":               DataVolumeSpecSchema(),
		"status":             dataVolumeStatusSchema(),
	}
}

// ToResourceDataFromVM writes metadata, spec, and status for refresh after create/update:
// status is normalized from the current resource config (expand + flatten), matching legacy ToResourceData.
func ToResourceDataFromVMTemplate(t *models.V1VMDataVolumeTemplateSpec, resourceData *schema.ResourceData) error {
	if t == nil {
		return nil
	}
	if err := resourceData.Set("metadata", k8s.FlattenMetadataDataVolumeFromVM(t.Metadata)); err != nil {
		return err
	}
	if err := resourceData.Set("spec", FlattenDataVolumeSpecFromVM(t.Spec)); err != nil {
		return err
	}
	// HAPI does not carry status
	// st := expandDataVolumeStatus(resourceData.Get("status").([]interface{}))
	// if err := resourceData.Set("status", flattenDataVolumeStatus(st)); err != nil {
	// 	return err
	// }
	if err := resourceData.Set("status", nil); err != nil {
		return err
	}
	return nil
}

// ToResourceDataFromVMTemplateRead refreshes metadata and spec from an API template during Read.
// Status is set to an empty flattened phase/progress, matching the previous FromHapiVolume+ToResourceData
// path (DataVolumeTemplate from the VM spec does not carry CDI status).
func ToResourceDataFromVMTemplateRead(t *models.V1VMDataVolumeTemplateSpec, resourceData *schema.ResourceData) error {
	if t == nil {
		return nil
	}
	if err := resourceData.Set("metadata", k8s.FlattenMetadataDataVolumeFromVM(t.Metadata)); err != nil {
		return err
	}
	if err := resourceData.Set("spec", FlattenDataVolumeSpecFromVM(t.Spec)); err != nil {
		return err
	}
	// HAPI does not carry status
	// if err := resourceData.Set("status", flattenDataVolumeStatus(cdiv1.DataVolumeStatus{})); err != nil {
	// 	return err
	// }
	return nil
}

// ToResourceDataFromVMAddVolumeEntity writes Terraform state from models.V1VMAddVolumeEntity
// (same shape as create/read against the Palette VM API) without CDI types.
func ToResourceDataFromVMAddVolumeEntity(v *models.V1VMAddVolumeEntity, resourceData *schema.ResourceData) error {
	if v == nil {
		return nil
	}
	return ToResourceDataFromVMTemplateRead(v.DataVolumeTemplate, resourceData)
}

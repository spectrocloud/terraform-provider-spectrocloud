package datavolume

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	cdiv1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"

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

func FromResourceData(resourceData *schema.ResourceData) (*cdiv1.DataVolume, error) {
	result := &cdiv1.DataVolume{}

	result.ObjectMeta = k8s.ExpandMetadata(resourceData.Get("metadata").([]interface{}))
	spec, err := ExpandDataVolumeSpec(resourceData.Get("spec").([]interface{}))
	if err != nil {
		return result, err
	}
	result.Spec = spec
	result.Status = expandDataVolumeStatus(resourceData.Get("status").([]interface{}))

	return result, nil
}

func ToResourceData(dv cdiv1.DataVolume, resourceData *schema.ResourceData) error {

	if err := resourceData.Set("metadata", k8s.FlattenMetadataDataVolume(dv.ObjectMeta)); err != nil {
		return err
	}
	if err := resourceData.Set("spec", FlattenDataVolumeSpec(dv.Spec)); err != nil {
		return err
	}
	if err := resourceData.Set("status", flattenDataVolumeStatus(dv.Status)); err != nil {
		return err
	}

	return nil
}

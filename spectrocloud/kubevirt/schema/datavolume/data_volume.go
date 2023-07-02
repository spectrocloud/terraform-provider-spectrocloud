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

func ExpandDataVolumeTemplates(dataVolumes []interface{}) ([]cdiv1.DataVolume, error) {
	result := make([]cdiv1.DataVolume, len(dataVolumes))

	if len(dataVolumes) == 0 || dataVolumes[0] == nil {
		return result, nil
	}

	for i, dataVolume := range dataVolumes {
		in := dataVolume.(map[string]interface{})

		if v, ok := in["metadata"].([]interface{}); ok {
			result[i].ObjectMeta = k8s.ExpandMetadata(v)
		}
		if v, ok := in["spec"].([]interface{}); ok {
			spec, err := ExpandDataVolumeSpec(v)
			if err != nil {
				return result, err
			}
			result[i].Spec = spec
		}
		if v, ok := in["status"].([]interface{}); ok {
			result[i].Status = expandDataVolumeStatus(v)
		}
	}

	return result, nil
}

func FlattenDataVolumeTemplates(in []cdiv1.DataVolume) []interface{} {
	att := make([]interface{}, len(in))

	for i, v := range in {
		c := make(map[string]interface{})
		c["metadata"] = k8s.FlattenMetadata(v.ObjectMeta)
		c["spec"] = FlattenDataVolumeSpec(v.Spec)
		c["status"] = flattenDataVolumeStatus(v.Status)
		att[i] = c
	}

	return att
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
	if err := resourceData.Set("metadata", k8s.FlattenMetadata(dv.ObjectMeta)); err != nil {
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

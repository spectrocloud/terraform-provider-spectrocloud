package convert

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/utils"
)

// SchemaToHapiVmMetadata converts Terraform schema metadata fields to HAPI VM ObjectMeta
func SchemaToHapiVmMetadata(d *schema.ResourceData) (*models.V1VMObjectMeta, error) {
	metadata := &models.V1VMObjectMeta{}

	// Name - Required field
	if v, ok := d.GetOk("name"); ok {
		metadata.Name = v.(string)
	} else {
		return nil, fmt.Errorf("name is required")
	}

	// Namespace - Optional, defaults to "default"
	if v, ok := d.GetOk("namespace"); ok {
		metadata.Namespace = v.(string)
	} else {
		metadata.Namespace = "default"
	}

	// GenerateName - Optional
	if v, ok := d.GetOk("generate_name"); ok {
		metadata.GenerateName = v.(string)
	}

	// Labels - Optional
	if v, ok := d.GetOk("labels"); ok {
		labelsMap := v.(map[string]interface{})
		if len(labelsMap) > 0 {
			metadata.Labels = utils.ExpandStringMap(labelsMap)
		}
	}

	// Annotations - Optional, Computed
	if v, ok := d.GetOk("annotations"); ok {
		annotationsMap := v.(map[string]interface{})
		if len(annotationsMap) > 0 {
			metadata.Annotations = utils.ExpandStringMap(annotationsMap)
		}
	}

	// Generation - Computed (read-only)
	if v, ok := d.GetOk("generation"); ok {
		metadata.Generation = int64(v.(int))
	}

	// ResourceVersion - Computed (read-only)
	if v, ok := d.GetOk("resource_version"); ok {
		metadata.ResourceVersion = v.(string)
	}

	// UID - Computed (read-only)
	if v, ok := d.GetOk("uid"); ok {
		metadata.UID = v.(string)
	}

	// Note: DeletionGracePeriodSeconds, Finalizers, OwnerReferences, ManagedFields
	// are not exposed in Terraform schema, so we leave them as nil/empty

	return metadata, nil
}

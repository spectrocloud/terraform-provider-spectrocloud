package convert

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

// HapiDataVolumeTemplatesToSchema converts HAPI DataVolumeTemplateSpec to Terraform schema
func HapiDataVolumeTemplatesToSchema(templates []*models.V1VMDataVolumeTemplateSpec, d *schema.ResourceData) error {
	if len(templates) == 0 {
		return nil
	}

	result := make([]interface{}, len(templates))

	for i, template := range templates {
		if template == nil {
			continue
		}

		templateMap := make(map[string]interface{})

		// Convert metadata
		if template.Metadata != nil {
			// Convert HAPI metadata to k8s metadata via JSON, then flatten
			metaJSON, err := json.Marshal(template.Metadata)
			if err != nil {
				return fmt.Errorf("failed to marshal metadata to JSON: %w", err)
			}

			var k8sMeta map[string]interface{}
			if err := json.Unmarshal(metaJSON, &k8sMeta); err != nil {
				return fmt.Errorf("failed to unmarshal metadata JSON: %w", err)
			}

			// Use k8s.FlattenMetadataDataVolume pattern
			// For now, convert directly
			metadataMap := make(map[string]interface{})
			if template.Metadata.Name != "" {
				metadataMap["name"] = template.Metadata.Name
			}
			if template.Metadata.Namespace != "" {
				metadataMap["namespace"] = template.Metadata.Namespace
			}
			if template.Metadata.GenerateName != "" {
				metadataMap["generate_name"] = template.Metadata.GenerateName
			}
			if template.Metadata.Labels != nil {
				metadataMap["labels"] = template.Metadata.Labels
			}
			if template.Metadata.Annotations != nil {
				metadataMap["annotations"] = template.Metadata.Annotations
			}
			if template.Metadata.ResourceVersion != "" {
				metadataMap["resource_version"] = template.Metadata.ResourceVersion
			}
			if template.Metadata.UID != "" {
				metadataMap["uid"] = template.Metadata.UID
			}
			if template.Metadata.Generation != 0 {
				metadataMap["generation"] = int(template.Metadata.Generation)
			}

			templateMap["metadata"] = []interface{}{metadataMap}
		}

		// Convert spec
		if template.Spec != nil {
			specJSON, err := json.Marshal(template.Spec)
			if err != nil {
				return fmt.Errorf("failed to marshal spec to JSON: %w", err)
			}

			var specMap map[string]interface{}
			if err := json.Unmarshal(specJSON, &specMap); err != nil {
				return fmt.Errorf("failed to unmarshal spec JSON: %w", err)
			}

			// Convert JSON keys to Terraform schema format
			specTerraform := make(map[string]interface{})

			if contentType, ok := specMap["contentType"].(string); ok {
				specTerraform["content_type"] = contentType
			}

			if source, ok := specMap["source"].(map[string]interface{}); ok {
				specTerraform["source"] = []interface{}{source}
			}

			if pvc, ok := specMap["pvc"].(map[string]interface{}); ok {
				// Convert PVC spec to Terraform format
				specTerraform["pvc"] = []interface{}{pvc}
			}

			if storage, ok := specMap["storage"].(map[string]interface{}); ok {
				// Convert storage spec to Terraform format
				storageTerraform := make(map[string]interface{})

				if accessModes, ok := storage["accessModes"].([]interface{}); ok {
					storageTerraform["access_modes"] = accessModes
				}

				if resources, ok := storage["resources"].(map[string]interface{}); ok {
					storageTerraform["resources"] = []interface{}{resources}
				}

				if selector, ok := storage["selector"].(map[string]interface{}); ok {
					storageTerraform["selector"] = []interface{}{selector}
				}

				if volumeName, ok := storage["volumeName"].(string); ok {
					storageTerraform["volume_name"] = volumeName
				}

				if storageClassName, ok := storage["storageClassName"].(string); ok {
					storageTerraform["storage_class_name"] = storageClassName
				}

				if volumeMode, ok := storage["volumeMode"].(string); ok {
					storageTerraform["volume_mode"] = volumeMode
				}

				specTerraform["storage"] = []interface{}{storageTerraform}
			}

			templateMap["spec"] = []interface{}{specTerraform}
		}

		result[i] = templateMap
	}

	return d.Set("data_volume_templates", result)
}

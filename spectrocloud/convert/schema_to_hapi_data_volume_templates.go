package convert

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/schema/k8s"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/utils"
)

// SchemaToHapiDataVolumeTemplates converts Terraform schema data volume templates to HAPI DataVolumeTemplateSpec
func SchemaToHapiDataVolumeTemplates(d *schema.ResourceData) ([]*models.V1VMDataVolumeTemplateSpec, error) {
	if v, ok := d.GetOk("data_volume_templates"); !ok {
		return nil, nil
	} else {
		dataVolumes := v.([]interface{})
		if len(dataVolumes) == 0 {
			return nil, nil
		}

		result := make([]*models.V1VMDataVolumeTemplateSpec, len(dataVolumes))

		for i, dataVolume := range dataVolumes {
			if dataVolume == nil {
				continue
			}

			in := dataVolume.(map[string]interface{})

			// Convert metadata
			var metadata *models.V1VMObjectMeta
			if v, ok := in["metadata"].([]interface{}); ok && len(v) > 0 {
				// Use existing k8s.ExpandMetadata but convert to HAPI via JSON
				k8sMeta := k8s.ExpandMetadata(v)
				metaJSON, err := json.Marshal(k8sMeta)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal metadata to JSON: %w", err)
				}
				if err := json.Unmarshal(metaJSON, &metadata); err != nil {
					return nil, fmt.Errorf("failed to unmarshal JSON to HAPI metadata: %w", err)
				}
			}

			// Convert spec
			var spec *models.V1VMDataVolumeSpec
			if v, ok := in["spec"].([]interface{}); ok && len(v) > 0 {
				specJSON, err := SchemaToHapiDataVolumeSpecJSON(v)
				if err != nil {
					return nil, fmt.Errorf("failed to convert data volume spec: %w", err)
				}

				// Unmarshal to HAPI spec
				if err := json.Unmarshal(specJSON, &spec); err != nil {
					return nil, fmt.Errorf("failed to unmarshal JSON to HAPI data volume spec: %w", err)
				}
			}

			result[i] = &models.V1VMDataVolumeTemplateSpec{
				Metadata: metadata,
				Spec:     spec,
			}
		}

		return result, nil
	}
}

// SchemaToHapiDataVolumeSpecJSON converts Terraform schema data volume spec to JSON bytes
// This creates a JSON-compatible structure from Terraform schema
func SchemaToHapiDataVolumeSpecJSON(spec []interface{}) ([]byte, error) {
	if len(spec) == 0 || spec[0] == nil {
		return json.Marshal(map[string]interface{}{})
	}

	specMap := spec[0].(map[string]interface{})
	specJSON := make(map[string]interface{})

	// ContentType
	if v, ok := specMap["content_type"].(string); ok {
		specJSON["contentType"] = v
	}

	// Source
	if v, ok := specMap["source"].([]interface{}); ok && len(v) > 0 {
		sourceJSON, err := expandDataVolumeSourceJSON(v)
		if err != nil {
			return nil, fmt.Errorf("failed to expand data volume source: %w", err)
		}
		specJSON["source"] = sourceJSON
	}

	// PVC
	if v, ok := specMap["pvc"].([]interface{}); ok && len(v) > 0 {
		pvcJSON, err := expandPVCSpecJSON(v)
		if err != nil {
			return nil, fmt.Errorf("failed to expand PVC spec: %w", err)
		}
		specJSON["pvc"] = pvcJSON
	}

	// Storage
	if v, ok := specMap["storage"].([]interface{}); ok && len(v) > 0 {
		storageJSON, err := expandStorageSpecJSON(v)
		if err != nil {
			return nil, fmt.Errorf("failed to expand storage spec: %w", err)
		}
		specJSON["storage"] = storageJSON
	}

	return json.Marshal(specJSON)
}

// expandDataVolumeSourceJSON converts Terraform schema data volume source to JSON-compatible structure
func expandDataVolumeSourceJSON(source []interface{}) (map[string]interface{}, error) {
	if len(source) == 0 || source[0] == nil {
		return nil, nil
	}

	sourceMap := source[0].(map[string]interface{})
	sourceJSON := make(map[string]interface{})

	// Handle different source types (HTTP, S3, Registry, PVC, Blank, etc.)
	if v, ok := sourceMap["http"].([]interface{}); ok && len(v) > 0 {
		httpMap := v[0].(map[string]interface{})
		httpJSON := make(map[string]interface{})
		if url, ok := httpMap["url"].(string); ok {
			httpJSON["url"] = url
		}
		if secretRef, ok := httpMap["secret_ref"].(string); ok {
			httpJSON["secretRef"] = secretRef
		}
		if certConfigMap, ok := httpMap["cert_config_map"].(string); ok {
			httpJSON["certConfigMap"] = certConfigMap
		}
		sourceJSON["http"] = httpJSON
	}

	if v, ok := sourceMap["s3"].([]interface{}); ok && len(v) > 0 {
		s3Map := v[0].(map[string]interface{})
		s3JSON := make(map[string]interface{})
		if url, ok := s3Map["url"].(string); ok {
			s3JSON["url"] = url
		}
		if secretRef, ok := s3Map["secret_ref"].(string); ok {
			s3JSON["secretRef"] = secretRef
		}
		sourceJSON["s3"] = s3JSON
	}

	if v, ok := sourceMap["registry"].([]interface{}); ok && len(v) > 0 {
		registryMap := v[0].(map[string]interface{})
		registryJSON := make(map[string]interface{})
		if url, ok := registryMap["url"].(string); ok {
			registryJSON["url"] = url
		}
		if secretRef, ok := registryMap["secret_ref"].(string); ok {
			registryJSON["secretRef"] = secretRef
		}
		if pullMethod, ok := registryMap["pull_method"].(string); ok {
			registryJSON["pullMethod"] = pullMethod
		}
		sourceJSON["registry"] = registryJSON
	}

	if v, ok := sourceMap["pvc"].([]interface{}); ok && len(v) > 0 {
		pvcSourceMap := v[0].(map[string]interface{})
		pvcSourceJSON := make(map[string]interface{})
		if namespace, ok := pvcSourceMap["namespace"].(string); ok {
			pvcSourceJSON["namespace"] = namespace
		}
		if name, ok := pvcSourceMap["name"].(string); ok {
			pvcSourceJSON["name"] = name
		}
		sourceJSON["pvc"] = pvcSourceJSON
	}

	if v, ok := sourceMap["blank"].([]interface{}); ok && len(v) > 0 {
		sourceJSON["blank"] = map[string]interface{}{}
	}

	return sourceJSON, nil
}

// expandPVCSpecJSON converts Terraform schema PVC spec to JSON-compatible structure
func expandPVCSpecJSON(pvc []interface{}) (map[string]interface{}, error) {
	if len(pvc) == 0 || pvc[0] == nil {
		return nil, nil
	}

	// Use existing k8s.ExpandPersistentVolumeClaimSpec but convert via JSON
	k8sPVC, err := k8s.ExpandPersistentVolumeClaimSpec(pvc)
	if err != nil {
		return nil, fmt.Errorf("failed to expand PVC spec: %w", err)
	}
	pvcJSON, err := json.Marshal(k8sPVC)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal PVC spec to JSON: %w", err)
	}

	var pvcSpecJSON map[string]interface{}
	if err := json.Unmarshal(pvcJSON, &pvcSpecJSON); err != nil {
		return nil, fmt.Errorf("failed to unmarshal PVC spec JSON: %w", err)
	}

	return pvcSpecJSON, nil
}

// expandStorageSpecJSON converts Terraform schema storage spec to JSON-compatible structure
func expandStorageSpecJSON(storage []interface{}) (map[string]interface{}, error) {
	if len(storage) == 0 || storage[0] == nil {
		return nil, nil
	}

	storageMap := storage[0].(map[string]interface{})
	storageJSON := make(map[string]interface{})

	// AccessModes
	if v, ok := storageMap["access_modes"].(*schema.Set); ok {
		accessModes := make([]string, v.Len())
		for i, mode := range v.List() {
			accessModes[i] = mode.(string)
		}
		storageJSON["accessModes"] = accessModes
	}

	// Resources
	if v, ok := storageMap["resources"].([]interface{}); ok && len(v) > 0 {
		resourcesMap := v[0].(map[string]interface{})
		resourcesJSON := make(map[string]interface{})

		if requests, ok := resourcesMap["requests"].(map[string]interface{}); ok {
			requestsMap := make(map[string]string)
			for k, v := range requests {
				requestsMap[k] = v.(string)
			}
			resourcesJSON["requests"] = requestsMap
		}

		if limits, ok := resourcesMap["limits"].(map[string]interface{}); ok {
			limitsMap := make(map[string]string)
			for k, v := range limits {
				limitsMap[k] = v.(string)
			}
			resourcesJSON["limits"] = limitsMap
		}

		storageJSON["resources"] = resourcesJSON
	}

	// Selector
	if v, ok := storageMap["selector"].([]interface{}); ok && len(v) > 0 {
		selectorJSON, err := expandLabelSelectorJSON(v)
		if err != nil {
			return nil, fmt.Errorf("failed to expand selector: %w", err)
		}
		storageJSON["selector"] = selectorJSON
	}

	// VolumeName
	if v, ok := storageMap["volume_name"].(string); ok {
		storageJSON["volumeName"] = v
	}

	// StorageClassName
	if v, ok := storageMap["storage_class_name"].(string); ok {
		storageJSON["storageClassName"] = v
	}

	// VolumeMode
	if v, ok := storageMap["volume_mode"].(string); ok {
		storageJSON["volumeMode"] = v
	}

	return storageJSON, nil
}

// expandLabelSelectorJSON converts Terraform schema label selector to JSON-compatible structure
func expandLabelSelectorJSON(selector []interface{}) (map[string]interface{}, error) {
	if len(selector) == 0 || selector[0] == nil {
		return nil, nil
	}

	selectorMap := selector[0].(map[string]interface{})
	selectorJSON := make(map[string]interface{})

	// MatchLabels
	if v, ok := selectorMap["match_labels"].(map[string]interface{}); ok {
		selectorJSON["matchLabels"] = utils.ExpandStringMap(v)
	}

	// MatchExpressions
	if v, ok := selectorMap["match_expressions"].([]interface{}); ok {
		matchExprs := make([]map[string]interface{}, len(v))
		for i, expr := range v {
			exprMap := expr.(map[string]interface{})
			exprJSON := make(map[string]interface{})
			if key, ok := exprMap["key"].(string); ok {
				exprJSON["key"] = key
			}
			if operator, ok := exprMap["operator"].(string); ok {
				exprJSON["operator"] = operator
			}
			if values, ok := exprMap["values"].([]interface{}); ok {
				valueStrs := make([]string, len(values))
				for j, val := range values {
					valueStrs[j] = val.(string)
				}
				exprJSON["values"] = valueStrs
			}
			matchExprs[i] = exprJSON
		}
		selectorJSON["matchExpressions"] = matchExprs
	}

	return selectorJSON, nil
}

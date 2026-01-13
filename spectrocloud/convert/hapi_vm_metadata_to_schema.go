package convert

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/utils"
)

// HapiVmMetadataToSchema converts HAPI VM ObjectMeta to Terraform schema fields
func HapiVmMetadataToSchema(metadata *models.V1VMObjectMeta, d *schema.ResourceData) error {
	if metadata == nil {
		return fmt.Errorf("metadata is nil")
	}

	var err error

	// Name - Required
	if err = d.Set("name", metadata.Name); err != nil {
		return fmt.Errorf("failed to set name: %w", err)
	}

	// Namespace - Optional, defaults to "default"
	namespace := metadata.Namespace
	if namespace == "" {
		namespace = "default"
	}
	if err = d.Set("namespace", namespace); err != nil {
		return fmt.Errorf("failed to set namespace: %w", err)
	}

	// GenerateName - Optional
	if metadata.GenerateName != "" {
		if err = d.Set("generate_name", metadata.GenerateName); err != nil {
			return fmt.Errorf("failed to set generate_name: %w", err)
		}
	}

	// Labels - Optional
	if metadata.Labels != nil && len(metadata.Labels) > 0 {
		if err = d.Set("labels", utils.FlattenStringMap(metadata.Labels)); err != nil {
			return fmt.Errorf("failed to set labels: %w", err)
		}
	}

	// Annotations - Optional, Computed
	// Filter out system-managed annotations (similar to existing FlattenMetadata)
	if metadata.Annotations != nil && len(metadata.Annotations) > 0 {
		filteredAnnotations := filterSystemAnnotations(metadata.Annotations)
		if len(filteredAnnotations) > 0 {
			if err = d.Set("annotations", utils.FlattenStringMap(filteredAnnotations)); err != nil {
				return fmt.Errorf("failed to set annotations: %w", err)
			}
		}
	}

	// Generation - Computed
	if err = d.Set("generation", int(metadata.Generation)); err != nil {
		return fmt.Errorf("failed to set generation: %w", err)
	}

	// ResourceVersion - Computed
	if err = d.Set("resource_version", metadata.ResourceVersion); err != nil {
		return fmt.Errorf("failed to set resource_version: %w", err)
	}

	// UID - Computed
	if err = d.Set("uid", metadata.UID); err != nil {
		return fmt.Errorf("failed to set uid: %w", err)
	}

	return nil
}

// filterSystemAnnotations removes system-managed annotations that should not be managed by Terraform
func filterSystemAnnotations(annotations map[string]string) map[string]string {
	if annotations == nil {
		return nil
	}

	filtered := make(map[string]string)
	for key, value := range annotations {
		// Filter out all kubevirt.io/ system annotations
		if !strings.HasPrefix(key, "kubevirt.io/") {
			filtered[key] = value
		}
	}

	return filtered
}

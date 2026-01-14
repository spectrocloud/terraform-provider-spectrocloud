package convert

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

// HapiVmToSchema converts HAPI VM model directly to Terraform schema
// This is the main entry point for HAPI VM → Terraform Schema conversion
func HapiVmToSchema(hapiVM *models.V1ClusterVirtualMachine, d *schema.ResourceData) error {
	if hapiVM == nil {
		return fmt.Errorf("hapiVM is nil")
	}

	// Convert metadata
	if err := HapiVmMetadataToSchema(hapiVM.Metadata, d); err != nil {
		return fmt.Errorf("failed to convert metadata: %w", err)
	}

	// Convert spec
	if err := HapiVmSpecToSchema(hapiVM.Spec, d); err != nil {
		return fmt.Errorf("failed to convert spec: %w", err)
	}

	// Convert status
	if err := HapiVmStatusToSchema(hapiVM.Status, d); err != nil {
		return fmt.Errorf("failed to convert status: %w", err)
	}

	return nil
}

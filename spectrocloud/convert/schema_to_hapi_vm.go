package convert

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

// SchemaToHapiVm converts Terraform schema directly to HAPI VM model
// This is the main entry point for Terraform Schema → HAPI VM conversion
func SchemaToHapiVm(d *schema.ResourceData) (*models.V1ClusterVirtualMachine, error) {
	hapiVM := &models.V1ClusterVirtualMachine{}

	// Set Kind and APIVersion
	hapiVM.Kind = "VirtualMachine"
	hapiVM.APIVersion = "kubevirt.io/v1"

	// Convert metadata
	metadata, err := SchemaToHapiVmMetadata(d)
	if err != nil {
		return nil, fmt.Errorf("failed to convert metadata: %w", err)
	}
	hapiVM.Metadata = metadata

	// Convert spec
	spec, err := SchemaToHapiVmSpec(d)
	if err != nil {
		return nil, fmt.Errorf("failed to convert spec: %w", err)
	}
	hapiVM.Spec = spec

	// Convert status (optional, usually read-only)
	status, err := SchemaToHapiVmStatus(d)
	if err != nil {
		return nil, fmt.Errorf("failed to convert status: %w", err)
	}
	if status != nil {
		hapiVM.Status = status
	}

	return hapiVM, nil
}

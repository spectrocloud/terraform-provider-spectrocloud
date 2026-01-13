package convert

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

// SchemaToHapiVmSpec converts Terraform schema spec fields to HAPI VM Spec
func SchemaToHapiVmSpec(d *schema.ResourceData) (*models.V1ClusterVirtualMachineSpec, error) {
	spec := &models.V1ClusterVirtualMachineSpec{}

	// Handle run_strategy (mutually exclusive with run_on_launch)
	if v, ok := d.GetOk("run_strategy"); ok {
		if v.(string) != "" {
			spec.RunStrategy = v.(string)
		}
	}

	// Handle run_on_launch (mutually exclusive with run_strategy)
	// This is handled in the resource file, but we need to ensure it's set correctly
	// If run_on_launch is false, RunStrategy should be "Manual"
	// If run_on_launch is true, Running should be true
	// Note: This logic is also handled in the resource file, so we'll set it here for completeness
	if _, ok := d.GetOk("run_on_launch"); ok {
		// The actual conversion is done in the resource file
		// We just ensure the spec is ready
	}

	// Convert template spec
	templateSpec, err := SchemaToHapiTemplateSpec(d)
	if err != nil {
		return nil, fmt.Errorf("failed to convert template spec: %w", err)
	}
	if templateSpec != nil {
		spec.Template = &models.V1VMVirtualMachineInstanceTemplateSpec{
			Spec: templateSpec,
		}
	}

	// Convert data volume templates
	dataVolumeTemplates, err := SchemaToHapiDataVolumeTemplates(d)
	if err != nil {
		return nil, fmt.Errorf("failed to convert data volume templates: %w", err)
	}
	if len(dataVolumeTemplates) > 0 {
		spec.DataVolumeTemplates = dataVolumeTemplates
	}

	return spec, nil
}

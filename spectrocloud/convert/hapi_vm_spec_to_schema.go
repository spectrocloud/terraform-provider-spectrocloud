package convert

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

// HapiVmSpecToSchema converts HAPI VM Spec to Terraform schema fields
func HapiVmSpecToSchema(spec *models.V1ClusterVirtualMachineSpec, d *schema.ResourceData) error {
	if spec == nil {
		return nil
	}

	var err error

	// Handle run_strategy
	if spec.RunStrategy != "" {
		if err = d.Set("run_strategy", spec.RunStrategy); err != nil {
			return fmt.Errorf("failed to set run_strategy: %w", err)
		}
	}

	// Handle run_on_launch based on Running field
	// If Running is true, set run_on_launch to true
	// If RunStrategy is "Manual", set run_on_launch to false
	if spec.Running {
		if err = d.Set("run_on_launch", true); err != nil {
			return fmt.Errorf("failed to set run_on_launch: %w", err)
		}
	} else if spec.RunStrategy == "Manual" {
		if err = d.Set("run_on_launch", false); err != nil {
			return fmt.Errorf("failed to set run_on_launch: %w", err)
		}
	}

	// Convert template spec
	if spec.Template != nil && spec.Template.Spec != nil {
		if err = HapiTemplateSpecToSchema(spec.Template.Spec, d); err != nil {
			return fmt.Errorf("failed to convert template spec: %w", err)
		}
	}

	// Convert data volume templates
	if spec.DataVolumeTemplates != nil && len(spec.DataVolumeTemplates) > 0 {
		if err = HapiDataVolumeTemplatesToSchema(spec.DataVolumeTemplates, d); err != nil {
			return fmt.Errorf("failed to convert data volume templates: %w", err)
		}
	}

	return nil
}

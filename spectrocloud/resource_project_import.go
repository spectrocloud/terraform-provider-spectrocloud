package spectrocloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/client/herr"
)

func resourceProjectImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c := getV1ClientWithResourceContext(m, "")

	importID := d.Id()
	if importID == "" {
		return nil, fmt.Errorf("project import ID or name is required")
	}

	// Try to get by UID first
	project, err := c.GetProject(importID)
	if err != nil {
		if !herr.IsNotFound(err) {
			return nil, fmt.Errorf("unable to retrieve project '%s': %w", importID, err)
		}
		// Not found by UID — try by name
		projectUID, nameErr := c.GetProjectUID(importID)
		if nameErr != nil {
			return nil, fmt.Errorf("unable to retrieve project by name or id '%s': %w", importID, nameErr)
		}
		if projectUID == "" {
			return nil, fmt.Errorf("project '%s' not found", importID)
		}
		d.SetId(projectUID)
	} else if project != nil {
		d.SetId(importID)
	} else {
		return nil, fmt.Errorf("project '%s' not found", importID)
	}

	// Read all project data to populate the state
	diags := resourceProjectRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read project for import: %v", diags)
	}

	return []*schema.ResourceData{d}, nil
}

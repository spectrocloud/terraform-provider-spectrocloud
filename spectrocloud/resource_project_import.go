package spectrocloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceProjectImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c := getV1ClientWithResourceContext(m, "")

	// The import ID should be the project UID
	projectUID := d.Id()

	// Validate that the project exists and we can access it
	project, err := c.GetProject(projectUID)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve project for import: %s", err)
	}
	if project == nil {
		return nil, fmt.Errorf("project with ID %s not found", projectUID)
	}

	// Set the project name from the retrieved project
	if err := d.Set("name", project.Metadata.Name); err != nil {
		return nil, err
	}

	// Read all project data to populate the state
	diags := resourceProjectRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read project for import: %v", diags)
	}

	return []*schema.ResourceData{d}, nil
}

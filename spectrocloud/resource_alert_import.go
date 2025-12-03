package spectrocloud

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAlertImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	// Expected ID format: projectUID:component OR projectName:component
	idParts := strings.Split(d.Id(), ":")

	if len(idParts) != 2 {
		return nil, fmt.Errorf("invalid import ID format. Expected: 'projectUID:component' or 'projectName:component', got: %s", d.Id())
	}

	projectIdentifier := idParts[0]
	component := idParts[1]

	// Validate component
	if component != "ClusterHealth" {
		return nil, fmt.Errorf("invalid component: %s. Only 'ClusterHealth' is supported", component)
	}

	c := getV1ClientWithResourceContext(m, "")

	// Try to get project by UID first, then by name
	var projectName string
	var projectUID string

	// Check if projectIdentifier is a UID or name by trying to get the project
	pjt, err := c.GetProject(projectIdentifier)
	if err != nil {
		// If failed, try to get project UID from name
		projectUID, err = c.GetProjectUID(projectIdentifier)
		if err != nil {
			return nil, fmt.Errorf("could not find project with identifier '%s': %v", projectIdentifier, err)
		}
		// Get project details using the UID
		pjt, err = c.GetProject(projectUID)
		if err != nil {
			return nil, fmt.Errorf("could not get project details for UID '%s': %v", projectUID, err)
		}
		projectName = pjt.Metadata.Name
	} else {
		projectName = pjt.Metadata.Name
		projectUID = pjt.Metadata.UID
	}

	// Set the project and component in state
	if err := d.Set("project", projectName); err != nil {
		return nil, fmt.Errorf("error setting project: %v", err)
	}
	if err := d.Set("component", component); err != nil {
		return nil, fmt.Errorf("error setting component: %v", err)
	}

	// Set the canonical ID format
	d.SetId(fmt.Sprintf("%s:%s", projectUID, component))

	// Read all alert data to populate the state
	diags := resourceAlertRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read alert for import: %v", diags)
	}

	return []*schema.ResourceData{d}, nil
}

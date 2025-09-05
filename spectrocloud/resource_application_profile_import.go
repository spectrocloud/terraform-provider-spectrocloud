package spectrocloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceApplicationProfileImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	// Application profiles have a context, default to "project"
	c := getV1ClientWithResourceContext(m, "project")

	// The import ID should be the application profile UID
	profileUID := d.Id()

	// Validate that the application profile exists and we can access it
	appProfile, err := c.GetApplicationProfile(profileUID)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve application profile for import: %s", err)
	}
	if appProfile == nil {
		return nil, fmt.Errorf("application profile with ID %s not found", profileUID)
	}

	// Set the application profile name from the retrieved profile
	if err := d.Set("name", appProfile.Metadata.Name); err != nil {
		return nil, err
	}
	// Set the application profile version from the retrieved profile
	if err := d.Set("version", appProfile.Spec.Version); err != nil {
		return nil, err
	}
	// Set the cloud to all as default for import
	if err := d.Set("cloud", "all"); err != nil {
		return nil, err
	}
	// Set the context to project as default for import
	if err := d.Set("context", "project"); err != nil {
		return nil, err
	}

	// Read all application profile data to populate the state
	diags := resourceApplicationProfileRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read application profile for import: %v", diags)
	}

	return []*schema.ResourceData{d}, nil
}

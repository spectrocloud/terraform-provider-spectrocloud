package spectrocloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/client"
)

func resourceApplicationImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	_, err := GetCommonApplication(d, m)
	if err != nil {
		return nil, err
	}

	// Read all application data to populate the state
	diags := resourceApplicationRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read application for import: %v", diags)
	}

	return []*schema.ResourceData{d}, nil
}

func GetCommonApplication(d *schema.ResourceData, m interface{}) (*client.V1Client, error) {
	// Applications can work in both tenant and project context, so we'll try project first
	applicationID := d.Id()
	if applicationID == "" {
		return nil, fmt.Errorf("application ID is required for import")
	}

	// Try project context first
	c := getV1ClientWithResourceContext(m, "project")

	// Use the ID to retrieve the application data from the API
	appDeployment, err := c.GetApplication(applicationID)
	if err != nil {
		// If not found in project context, try tenant context
		c = getV1ClientWithResourceContext(m, "tenant")
		appDeployment, err = c.GetApplication(applicationID)
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve application data in either project or tenant context: %s", err)
		}
	}

	if appDeployment == nil {
		return nil, fmt.Errorf("application with ID %s not found", applicationID)
	}

	// Set the application name from the retrieved application
	if err := d.Set("name", appDeployment.Metadata.Name); err != nil {
		return nil, err
	}

	// Set placeholder values for required fields to prevent validation errors during import
	// These will be properly populated by the read function
	if appDeployment.Spec != nil && appDeployment.Spec.Profile != nil && appDeployment.Spec.Profile.Metadata != nil {
		if err := d.Set("application_profile_uid", appDeployment.Spec.Profile.Metadata.UID); err != nil {
			return nil, err
		}
	}

	// Set placeholder config with required cluster_context
	// The resource context will be determined and set properly in the read function
	if appDeployment.Spec != nil && appDeployment.Spec.Config != nil && appDeployment.Spec.Config.Target != nil {
		config := make(map[string]interface{})
		// Default to project context, will be adjusted in read function if needed
		config["cluster_context"] = "project"

		if err := d.Set("config", []interface{}{config}); err != nil {
			return nil, err
		}
	}

	// Set the ID of the resource in the state. This ID is used to track the
	// resource and must be set in the state during the import.
	d.SetId(appDeployment.Metadata.UID)

	return c, nil
}

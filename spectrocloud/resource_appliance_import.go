package spectrocloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/client"
)

func resourceApplianceImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	_, err := GetCommonAppliance(d, m)
	if err != nil {
		return nil, err
	}

	// Read all appliance data to populate the state
	diags := resourceApplianceRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read appliance for import: %v", diags)
	}

	return []*schema.ResourceData{d}, nil
}

func GetCommonAppliance(d *schema.ResourceData, m interface{}) (*client.V1Client, error) {
	// Appliances are project-level resources
	c := getV1ClientWithResourceContext(m, "project")

	// The import ID should be the appliance UID
	applianceUID := d.Id()
	if applianceUID == "" {
		return nil, fmt.Errorf("appliance import ID is required")
	}

	// Validate that the appliance exists and we can access it
	appliance, err := c.GetAppliance(applianceUID)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve appliance: %s", err)
	}
	if appliance == nil {
		return nil, fmt.Errorf("appliance with ID %s not found", applianceUID)
	}

	// Set the required uid field (this is what the resource uses internally)
	if err := d.Set("uid", appliance.Metadata.UID); err != nil {
		return nil, err
	}

	// Set optional fields if they exist
	if len(appliance.Metadata.Labels) > 0 {
		if err := d.Set("tags", appliance.Metadata.Labels); err != nil {
			return nil, err
		}
	}

	// Set other optional fields with default values to prevent validation errors
	if appliance.Spec != nil {
		// Set wait to false as default (this is likely what users expect for import)
		if err := d.Set("wait", false); err != nil {
			return nil, err
		}

		// Set remote shell access if configured
		if appliance.Spec.TunnelConfig != nil {
			if err := d.Set("remote_shell", appliance.Spec.TunnelConfig.RemoteSSH); err != nil {
				return nil, err
			}
			if err := d.Set("temporary_shell_credentials", appliance.Spec.TunnelConfig.RemoteSSHTempUser); err != nil {
				return nil, err
			}
		}
	}

	// Set the ID to the appliance UID
	d.SetId(applianceUID)

	return c, nil
}

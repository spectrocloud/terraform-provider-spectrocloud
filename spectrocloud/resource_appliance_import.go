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
	c := getV1ClientWithResourceContext(m, "")

	// The import ID should be the appliance UID
	idOrName := d.Id()
	if idOrName == "" {
		return nil, fmt.Errorf("appliance import ID or name is required")
	}

	// Validate that the appliance exists and we can access it
	appliance, err := c.GetAppliance(idOrName)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve appliance: %s", err)
	}

	// Fall back to lookup by name
	if err != nil || appliance == nil {
		appliance, err = c.GetApplianceByName(idOrName, nil, "", "", "")
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve appliance by name '%s': %w", idOrName, err)
		}
		if appliance == nil {
			return nil, fmt.Errorf("appliance with id or name '%s' not found", idOrName)
		}
	}

	// Set required uid field
	if err := d.Set("uid", appliance.Metadata.UID); err != nil {
		return nil, err
	}

	// Set optional fields if they exist
	if len(appliance.Metadata.Labels) > 0 {
		if err := d.Set("tags", appliance.Metadata.Labels); err != nil {
			return nil, err
		}
	}

	if appliance.Spec != nil {
		if err := d.Set("wait", false); err != nil {
			return nil, err
		}
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
	d.SetId(appliance.Metadata.UID)

	return c, nil
}

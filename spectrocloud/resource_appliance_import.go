package spectrocloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceApplianceImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c := getV1ClientWithResourceContext(m, "")

	// The import ID should be the appliance UID
	applianceUID := d.Id()

	// Validate that the appliance exists and we can access it
	appliance, err := c.GetAppliance(applianceUID)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve appliance for import: %s", err)
	}
	if appliance == nil {
		return nil, fmt.Errorf("appliance with ID %s not found", applianceUID)
	}

	// Set the appliance name from the retrieved appliance
	if err := d.Set("name", appliance.Metadata.Name); err != nil {
		return nil, err
	}

	// Read all appliance data to populate the state
	diags := resourceApplianceRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read appliance for import: %v", diags)
	}

	return []*schema.ResourceData{d}, nil
}

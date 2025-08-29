package spectrocloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceApplicationImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	// For applications, we need to extract cluster and profile information from the ID
	// Application IDs are typically composite in the format cluster_uid:profile_uid
	// Note: Applications might need special handling as they're linked to clusters
	// Setting basic name for now - the read function will populate the rest
	if err := d.Set("name", "imported_application"); err != nil {
		return nil, err
	}

	// Read all application data to populate the state
	diags := resourceApplicationRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read application for import: %v", diags)
	}

	return []*schema.ResourceData{d}, nil
}

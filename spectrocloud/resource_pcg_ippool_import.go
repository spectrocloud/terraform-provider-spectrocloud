package spectrocloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourcePrivateCloudGatewayIpPoolImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	// Set a placeholder name for now - the read function will populate the correct data
	if err := d.Set("name", "imported_ippool"); err != nil {
		return nil, err
	}

	// Read all IP pool data to populate the state
	diags := resourceIpPoolRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read IP pool for import: %v", diags)
	}

	return []*schema.ResourceData{d}, nil
}

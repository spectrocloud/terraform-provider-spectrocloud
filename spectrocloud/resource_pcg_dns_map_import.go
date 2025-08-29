package spectrocloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourcePrivateCloudGatewayDNSMapImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	// Set a placeholder domain for now - the read function will populate the correct data
	if err := d.Set("domain", "imported.domain"); err != nil {
		return nil, err
	}

	// Read all DNS map data to populate the state
	diags := resourcePCGDNSMapRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read PCG DNS map for import: %v", diags)
	}

	return []*schema.ResourceData{d}, nil
}

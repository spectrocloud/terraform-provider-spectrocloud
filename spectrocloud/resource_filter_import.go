package spectrocloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceFilterImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c := getV1ClientWithResourceContext(m, "")

	// The import ID should be the filter UID
	filterUID := d.Id()

	// Validate that the filter exists and we can access it
	tagFilter, err := c.GetTagFilter(filterUID)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve filter for import: %s", err)
	}
	if tagFilter == nil {
		return nil, fmt.Errorf("filter with ID %s not found", filterUID)
	}

	// Set the filter name from the retrieved filter metadata
	if tagFilter.Metadata != nil && tagFilter.Metadata.Name != "" {
		metadata := []interface{}{
			map[string]interface{}{
				"name": tagFilter.Metadata.Name,
			},
		}
		if err := d.Set("metadata", metadata); err != nil {
			return nil, err
		}
	}

	// Read all filter data to populate the state
	diags := resourceFilterRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read filter for import: %v", diags)
	}

	return []*schema.ResourceData{d}, nil
}

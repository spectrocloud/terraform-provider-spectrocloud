package spectrocloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/client/herr"
)

func resourceFilterImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c := getV1ClientWithResourceContext(m, "")

	// Import ID can be either the filter UID or the filter name
	importID := d.Id()
	if importID == "" {
		return nil, fmt.Errorf("filter import ID or name is required")
	}

	// Try to get by UID first
	tagFilter, err := c.GetTagFilter(importID)
	if err != nil {
		if !herr.IsNotFound(err) {
			return nil, fmt.Errorf("unable to retrieve filter '%s': %w", importID, err)
		}
		// Not found by UID — try by name
		filterByName, nameErr := c.GetTagFilterByName(importID)
		if nameErr != nil {
			return nil, fmt.Errorf("unable to retrieve filter by name or id '%s': %w", importID, nameErr)
		}
		if filterByName == nil || filterByName.Metadata == nil || filterByName.Metadata.UID == "" {
			return nil, fmt.Errorf("filter '%s' not found", importID)
		}
		d.SetId(filterByName.Metadata.UID)
	} else if tagFilter != nil {
		d.SetId(importID)
	} else {
		return nil, fmt.Errorf("filter '%s' not found", importID)
	}

	// Read all filter data to populate the state
	diags := resourceFilterRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read filter for import: %v", diags)
	}

	return []*schema.ResourceData{d}, nil
}

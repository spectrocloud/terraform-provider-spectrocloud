package spectrocloud

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/client"
)

func resourceClusterGroupImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	_, err := GetCommonClusterGroup(d, m)
	if err != nil {
		return nil, err
	}

	// Read all cluster group data to populate the state
	diags := resourceClusterGroupRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read cluster group for import: %v", diags)
	}

	return []*schema.ResourceData{d}, nil
}

// GetCommonClusterGroup resolves a cluster group from the import ID and populates
// the ResourceData with name and context. The import ID supports:
//
//	UID:CONTEXT   — e.g. abc123:project
//	NAME:CONTEXT  — e.g. my-group:tenant
func GetCommonClusterGroup(d *schema.ResourceData, m interface{}) (*client.V1Client, error) {
	parts := strings.Split(d.Id(), ":")
	if len(parts) != 2 || (strings.TrimSpace(parts[1]) != "project" && strings.TrimSpace(parts[1]) != "tenant") {
		return nil, fmt.Errorf("invalid import ID format: expected NAME_or_UID:CONTEXT (context must be project or tenant), got %q", d.Id())
	}

	idOrName := strings.TrimSpace(parts[0])
	resourceContext := strings.TrimSpace(parts[1])

	c := getV1ClientWithResourceContext(m, resourceContext)

	// 1) Try as UID first (backward-compatible with existing imports)
	clusterGroup, err := c.GetClusterGroup(idOrName)
	if err == nil && clusterGroup != nil {
		if err := d.Set("name", clusterGroup.Metadata.Name); err != nil {
			return nil, err
		}
		if err := d.Set("context", clusterGroup.Metadata.Annotations["scope"]); err != nil {
			return nil, err
		}
		d.SetId(clusterGroup.Metadata.UID)
		return c, nil
	}

	// 2) Treat as NAME: look up by name via summary, then fetch full object
	summary, err := c.GetClusterGroupSummaryByName(idOrName)
	if err != nil {
		return nil, fmt.Errorf("unable to look up cluster group by name %q: %w", idOrName, err)
	}
	if summary == nil {
		return nil, fmt.Errorf("cluster group %q not found in context %q", idOrName, resourceContext)
	}

	// Fetch the full object to get Metadata.Annotations["scope"] needed for context
	clusterGroup, err = c.GetClusterGroupWithoutStatus(summary.Metadata.UID)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve cluster group data for %q: %w", idOrName, err)
	}
	if clusterGroup == nil {
		return nil, fmt.Errorf("cluster group %q resolved by name but could not be retrieved", idOrName)
	}

	if err := d.Set("name", clusterGroup.Metadata.Name); err != nil {
		return nil, err
	}
	if err := d.Set("context", clusterGroup.Metadata.Annotations["scope"]); err != nil {
		return nil, err
	}
	d.SetId(clusterGroup.Metadata.UID)

	return c, nil
}

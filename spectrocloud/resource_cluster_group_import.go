package spectrocloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceClusterGroupImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	// Cluster groups have a context, default to "project"
	c := getV1ClientWithResourceContext(m, "project")

	// The import ID should be the cluster group UID
	clusterGroupUID := d.Id()

	// Validate that the cluster group exists and we can access it
	clusterGroup, err := c.GetClusterGroup(clusterGroupUID)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve cluster group for import: %s", err)
	}
	if clusterGroup == nil {
		return nil, fmt.Errorf("cluster group with ID %s not found", clusterGroupUID)
	}

	// Set the cluster group name from the retrieved cluster group
	if err := d.Set("name", clusterGroup.Metadata.Name); err != nil {
		return nil, err
	}

	// Set the context to project as default for import
	if err := d.Set("context", "project"); err != nil {
		return nil, err
	}

	// Read all cluster group data to populate the state
	diags := resourceClusterGroupRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read cluster group for import: %v", diags)
	}

	return []*schema.ResourceData{d}, nil
}

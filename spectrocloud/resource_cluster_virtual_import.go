package spectrocloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceClusterVirtualImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	// Virtual clusters have a context, default to "project"
	c := getV1ClientWithResourceContext(m, "project")

	// The import ID should be the cluster UID
	clusterUID := d.Id()

	// Validate that the cluster exists and we can access it
	cluster, err := c.GetCluster(clusterUID)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve virtual cluster for import: %s", err)
	}
	if cluster == nil {
		return nil, fmt.Errorf("virtual cluster with ID %s not found", clusterUID)
	}

	// Set the cluster name from the retrieved cluster
	if err := d.Set("name", cluster.Metadata.Name); err != nil {
		return nil, err
	}

	// Set the context to project as default for import
	if err := d.Set("context", "project"); err != nil {
		return nil, err
	}

	// Read all cluster data to populate the state
	diags := resourceClusterVirtualRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read virtual cluster for import: %v", diags)
	}

	return []*schema.ResourceData{d}, nil
}

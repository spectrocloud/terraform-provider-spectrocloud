package spectrocloud

import (
	"context"
	"fmt"

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

func GetCommonClusterGroup(d *schema.ResourceData, m interface{}) (*client.V1Client, error) {
	// Parse resource ID and scope
	resourceContext, clusterGroupID, err := ParseResourceID(d)
	if err != nil {
		return nil, err
	}
	c := getV1ClientWithResourceContext(m, resourceContext)

	// Use the ID to retrieve the cluster group data from the API
	clusterGroup, err := c.GetClusterGroup(clusterGroupID)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve cluster group data: %s", err)
	}
	if clusterGroup == nil {
		return nil, fmt.Errorf("cluster group with ID %s not found", clusterGroupID)
	}

	// Set the cluster group name from the retrieved cluster group
	if err := d.Set("name", clusterGroup.Metadata.Name); err != nil {
		return nil, err
	}

	// Set the context from the cluster group metadata
	if err := d.Set("context", clusterGroup.Metadata.Annotations["scope"]); err != nil {
		return nil, err
	}

	// Set the ID of the resource in the state. This ID is used to track the
	// resource and must be set in the state during the import.
	d.SetId(clusterGroup.Metadata.UID)

	return c, nil
}

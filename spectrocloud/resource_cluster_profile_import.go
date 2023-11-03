package spectrocloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/client"
)

func resourceClusterProfileImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c := m.(*client.V1Client)

	err := GetCommonClusterProfile(d, c)
	if err != nil {
		return nil, err
	}

	diags := resourceClusterProfileRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read cluster for import: %v", diags)
	}

	// Return the resource data. In most cases, this method is only used to
	// import one resource at a time, so you should return the resource data
	// in a slice with a single element.
	return []*schema.ResourceData{d}, nil
}

func GetCommonClusterProfile(d *schema.ResourceData, c *client.V1Client) error {
	// Use the IDs to retrieve the cluster data from the API
	clusterC, err := c.GetClusterClient()
	profile, err := c.GetClusterProfile(clusterC, d.Id())
	if err != nil {
		return fmt.Errorf("unable to retrieve cluster data: %s", err)
	}

	err = d.Set("name", profile.Metadata.Name)
	if err != nil {
		return err
	}
	err = d.Set("context", profile.Metadata.Annotations["scope"])
	if err != nil {
		return err
	}

	// Set the ID of the resource in the state. This ID is used to track the
	// resource and must be set in the state during the import.
	d.SetId(d.Id())
	return nil
}

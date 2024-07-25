package spectrocloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/client"
)

func resourceAccountGcpImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c := m.(*client.V1Client)

	err := GetCommonAccount(d, c)
	if err != nil {
		return nil, err
	}

	diags := resourceCloudAccountGcpRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read cluster for import: %v", diags)
	}

	// Return the resource data. In most cases, this method is only used to
	// import one resource at a time, so you should return the resource data
	// in a slice with a single element.
	return []*schema.ResourceData{d}, nil
}

func GetCommonAccount(d *schema.ResourceData, c *client.V1Client) error {
	// parse resource ID and scope
	scope, accountID, err := ParseResourceID(d)
	if err != nil {
		return err
	}

	// Use the IDs to retrieve the cluster data from the API
	cluster, err := c.GetCloudAccount(accountID)
	if err != nil {
		return fmt.Errorf("unable to retrieve cluster data: %s", err)
	}

	err = d.Set("name", cluster.Metadata.Name)
	if err != nil {
		return err
	}
	if cluster.Metadata.Annotations != nil {
		if scope != cluster.Metadata.Annotations["scope"] {
			return fmt.Errorf("CloudAccount scope mismatch: %s != %s", scope, cluster.Metadata.Annotations["scope"])
		}
		err = d.Set("context", cluster.Metadata.Annotations["scope"])
		if err != nil {
			return err
		}
	}
	// Set the ID of the resource in the state. This ID is used to track the
	// resource and must be set in the state during the import.
	d.SetId(accountID)
	return nil
}

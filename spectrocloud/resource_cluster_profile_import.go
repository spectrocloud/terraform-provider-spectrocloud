package spectrocloud

import (
	"context"
	"fmt"
	"github.com/spectrocloud/palette-sdk-go/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceClusterProfileImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {

	_, err := GetCommonClusterProfile(d, m)
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

func GetCommonClusterProfile(d *schema.ResourceData, m interface{}) (*client.V1Client, error) {

	resourceContext, profileID, err := ParseResourceID(d)
	if err != nil {
		return nil, err
	}
	c := getV1ClientWithResourceContext(m, resourceContext)
	profile, err := c.GetClusterProfile(profileID)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve cluster data: %s", err)
	}
	if profile == nil {
		return nil, fmt.Errorf("cluster profile id: %s not found", d.Id())
	}

	err = d.Set("name", profile.Metadata.Name)
	if err != nil {
		return nil, err
	}
	err = d.Set("context", profile.Metadata.Annotations["scope"])
	if err != nil {
		return nil, err
	}

	// Set the ID of the resource in the state. This ID is used to track the
	// resource and must be set in the state during the import.
	d.SetId(profile.Metadata.UID)

	return c, nil
}

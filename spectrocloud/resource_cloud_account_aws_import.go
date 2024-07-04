package spectrocloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAccountAwsImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {

	scope, _, err := ParseResourceID(d)
	if err != nil {
		return nil, err
	}
	c := GetResourceLevelV1Client(m, scope)

	err = GetCommonAccount(d, c)
	if err != nil {
		return nil, err
	}

	diags := resourceCloudAccountAwsRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read cluster for import: %v", diags)
	}

	// Return the resource data. In most cases, this method is only used to
	// import one resource at a time, so you should return the resource data
	// in a slice with a single element.
	return []*schema.ResourceData{d}, nil
}

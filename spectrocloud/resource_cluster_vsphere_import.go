package spectrocloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceClusterVsphereImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c, err := GetCommonCluster(d, m)
	if err != nil {
		return nil, err
	}

	diags := resourceClusterVsphereRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read cluster for import: %v", diags)
	}

	// cluster profile and common default cluster attribute is get set here
	err = flattenCommonAttributeForClusterImport(c, d)
	if err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

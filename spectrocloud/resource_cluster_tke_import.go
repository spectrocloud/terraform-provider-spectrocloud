package spectrocloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceClusterTkeImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	// m is the client, which can be used to make API requests to the infrastructure
	scope, _, err := ParseResourceID(d)
	if err != nil {
		return nil, err
	}
	c := GetResourceLevelV1Client(m, scope)

	err = GetCommonCluster(d, c)
	if err != nil {
		return nil, err
	}

	diags := resourceClusterTkeRead(ctx, d, m)
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

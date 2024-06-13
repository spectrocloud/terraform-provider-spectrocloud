package spectrocloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/client"
)

func resourceClusterTkeImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	// m is the client, which can be used to make API requests to the infrastructure
	c := m.(*client.V1Client)

	err := GetCommonCluster(d, c)
	if err != nil {
		return nil, err
	}

	diags := resourceClusterTkeRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read cluster for import: %v", diags)
	}

	clusterProfiles, err := flattenClusterProfileForImport(c, d)
	if err != nil {
		return nil, err
	}
	if err := d.Set("cluster_profile", clusterProfiles); err != nil {
		return nil, fmt.Errorf("could not read cluster for import: %v", diags)
	}

	err = setTFDefaultValueForClusterImport(d)
	if err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

package spectrocloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/client"
)

func resourceClusterVsphereImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c := m.(*client.V1Client)
	err := GetCommonCluster(d, c)
	if err != nil {
		return nil, err
	}

	diags := resourceClusterVsphereRead(ctx, d, m)
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
	err = setTerraformDefaultValueForClusterImport(d)
	if err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

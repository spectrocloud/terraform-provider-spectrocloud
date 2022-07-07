package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceCluster() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceClusterRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func dataSourceClusterRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	var diags diag.Diagnostics
	if name, okName := d.GetOk("name"); okName {
		cluster, err := c.GetClusterByName(name.(string))
		if err != nil {
			return diag.FromErr(err)
		}
		if cluster != nil {
			d.SetId(cluster.Metadata.UID)
			if err := d.Set("name", cluster.Metadata.Name); err != nil {
				return diag.FromErr(err)
			}
		}
	}
	return diags
}

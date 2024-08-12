package spectrocloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourcePrivateCloudGatewayIpPool() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceIpPoolRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the IP pool.",
			},
			"private_cloud_gateway_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the private cloud gateway.",
			},
		},
	}
}

func dataSourceIpPoolRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := getV1ClientWithResourceContext(m, "")
	pcgUID := d.Get("private_cloud_gateway_id").(string)
	name := d.Get("name").(string)

	pool, err := c.GetIPPoolByName(pcgUID, name)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(pool.Metadata.UID)
	err = d.Set("private_cloud_gateway_id", pcgUID)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("name", pool.Metadata.Name)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

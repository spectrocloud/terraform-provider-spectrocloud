package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"
)

func dataSourcePCG() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePCGRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:          schema.TypeString,
				Computed:      true,
				Optional:      true,
				ConflictsWith: []string{"name"},
				Description:   "The ID of Private Cloud Gateways.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "The Name of Private Cloud Gateways.",
			},
		},
	}
}

func dataSourcePCGRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	var diags diag.Diagnostics
	if v, ok := d.GetOk("name"); ok {
		name := v.(string)
		var namePointer *string
		namePointer = &name
		uid, err := c.GetPrivateCloudGatewayID(namePointer)
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(uid)
		d.Set("name", v.(string))
	}
	return diags
}

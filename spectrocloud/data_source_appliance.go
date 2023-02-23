package spectrocloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/spectrocloud/hapi/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAppliance() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceApplianceRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tags": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceApplianceRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	var diags diag.Diagnostics
	if name, okName := d.GetOk("name"); okName {
		appliance, err := c.GetApplianceByName(name.(string))
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(appliance.Metadata.UID)
		err = d.Set("name", appliance.Metadata.Name)
		if err != nil {
			return diag.FromErr(err)
		}
		err = d.Set("tags", appliance.Metadata.Labels)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	return diags
}

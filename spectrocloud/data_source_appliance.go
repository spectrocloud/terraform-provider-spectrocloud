package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"

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
			"labels": {
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
		d.Set("name", appliance.Metadata.Name)
		d.Set("labels", appliance.Metadata.Labels)
	}
	return diags
}

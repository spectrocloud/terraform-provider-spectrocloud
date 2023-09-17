package spectrocloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/spectrocloud/palette-sdk-go/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceRegistryPack() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRegistryPackRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func dataSourceRegistryPackRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	var diags diag.Diagnostics
	if v, ok := d.GetOk("name"); ok {
		registry, err := c.GetPackRegistryCommonByName(v.(string))
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(registry.UID)
		if err := d.Set("name", registry.Name); err != nil {
			return diag.FromErr(err)
		}
	}
	return diags
}

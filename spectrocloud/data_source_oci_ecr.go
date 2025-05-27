package spectrocloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceRegistryOci() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRegistryOciRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "The unique identifier of the OCI registry.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the OCI registry.",
			},
		},
	}
}

func dataSourceRegistryOciRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "")
	var diags diag.Diagnostics
	if v, ok := d.GetOk("name"); ok {
		registry, err := c.GetOciRegistryByName(v.(string))
		if err != nil {
			return handleReadError(d, err, diags)
		}
		d.SetId(registry.Metadata.UID)
		err = d.Set("name", registry.Metadata.Name)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	return diags
}

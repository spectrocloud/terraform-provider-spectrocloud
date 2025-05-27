package spectrocloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourcePCG() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePCGRead,
		Description: "A data resource to get the ID or name of Private Cloud Gateway.",

		Schema: map[string]*schema.Schema{
			"id": {
				Type:          schema.TypeString,
				Computed:      true,
				Optional:      true,
				ConflictsWith: []string{"name"},
				Description:   "The ID of Private Cloud Gateway.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "The name of Private Cloud Gateway.",
			},
		},
	}
}

func dataSourcePCGRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "")
	var diags diag.Diagnostics
	if v, ok := d.GetOk("name"); ok {
		name := v.(string)
		namePointer := &name
		uid, err := c.GetPCGId(namePointer)
		if err != nil {
			return handleReadError(d, err, diags)
		}
		d.SetId(uid)
		if err := d.Set("name", v.(string)); err != nil {
			return diag.FromErr(err)
		}
	}
	return diags
}

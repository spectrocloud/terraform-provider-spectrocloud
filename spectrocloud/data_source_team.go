package spectrocloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTeam() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTeamRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:          schema.TypeString,
				Computed:      true,
				Optional:      true,
				ConflictsWith: []string{"name"},
				Description:   "The unique ID of the team. If provided, `name` cannot be used.",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name of the team. If provided, `id` cannot be used.",
			},
		},
	}
}

func dataSourceTeamRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "")
	var diags diag.Diagnostics

	if v, ok := d.GetOk("name"); ok {
		team, err := c.GetTeam(v.(string))
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(team.Metadata.UID)
		if err := d.Set("email", team.Metadata.Name); err != nil {
			return diag.FromErr(err)
		}
	} else {
		if val, okay := d.GetOk("id"); okay && val != "" {
			team, err := c.GetTeam(val.(string))
			if err != nil {
				return diag.FromErr(err)
			}
			d.SetId(team.Metadata.UID)
			if err := d.Set("email", team.Metadata.Name); err != nil {
				return diag.FromErr(err)
			}
		}

	}
	return diags
}

package spectrocloud

import (
	"context"
	"github.com/spectrocloud/palette-sdk-go/api/models"

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
			"role_ids": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The roles id's assigned to the team.",
			},
		},
	}
}

func dataSourceTeamRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "")
	var diags diag.Diagnostics
	var team *models.V1Team
	var err error
	if v, ok := d.GetOk("name"); ok {
		team, err = c.GetTeamWithName(v.(string))
		if err != nil {
			return diag.FromErr(err)
		}
	} else {
		if val, okay := d.GetOk("id"); okay && val != "" {
			team, err = c.GetTeam(val.(string))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}
	if team != nil {
		d.SetId(team.Metadata.UID)
		if err := d.Set("name", team.Metadata.Name); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("role_ids", team.Spec.Roles); err != nil {
			return diag.FromErr(err)
		}
	}
	return diags
}

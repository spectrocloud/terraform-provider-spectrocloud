package spectrocloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceProject() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceProjectRead,
		Description: "Data source for looking up a Spectro Cloud project by name or id.",

		Schema: map[string]*schema.Schema{
			"id": {
				Type:          schema.TypeString,
				Computed:      true,
				Optional:      true,
				ConflictsWith: []string{"name"},
				Description:   "ID of the project to look up.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "Name of the project to look up.",
			},
		},
	}
}

func dataSourceProjectRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "")
	var diags diag.Diagnostics
	if v, ok := d.GetOk("name"); ok {
		uid, err := c.GetProjectUID(v.(string))
		if err != nil {
			return handleReadError(d, err, diags)
		}
		d.SetId(uid)
		if err := d.Set("name", v.(string)); err != nil {
			return diag.FromErr(err)
		}
	} else if v, ok := d.GetOk("id"); ok {
		project, err := c.GetProject(v.(string))
		if err != nil {
			return handleReadError(d, err, diags)
		}
		d.SetId(project.Metadata.UID)
		if err := d.Set("name", project.Metadata.Name); err != nil {
			return diag.FromErr(err)
		}
	}
	return diags
}

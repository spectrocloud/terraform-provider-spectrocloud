package spectrocloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceWorkspace() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceWorkspaceRead,
		Description: "Data source for looking up a Spectro Cloud workspace by name.",

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the workspace to look up.",
			},
		},
	}
}

func dataSourceWorkspaceRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "")
	var diags diag.Diagnostics
	if name, okName := d.GetOk("name"); okName {
		workspace, err := c.GetWorkspaceByName(name.(string))
		if err != nil {
			return handleReadError(d, err, diags)
		}
		if workspace != nil {
			d.SetId(workspace.Metadata.UID)
			if err := d.Set("name", workspace.Metadata.Name); err != nil {
				return diag.FromErr(err)
			}
		}
	}
	return diags
}

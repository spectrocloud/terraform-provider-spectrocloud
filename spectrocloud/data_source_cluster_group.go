package spectrocloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceClusterGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceClusterGroupRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the cluster group.",
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "tenant",
				ValidateFunc: validation.StringInSlice([]string{"", "tenant", "system", "project"}, false),
				Description:  "The context of where the cluster group is located. Allowed values  are `system` or `tenant`. Defaults to 'tenant'.",
			},
		},
	}
}

func dataSourceClusterGroupRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	var diags diag.Diagnostics
	if name, okName := d.GetOk("name"); okName {
		GroupContext := d.Get("context").(string)
		switch GroupContext {
		case "system", "tenant":
			group, err := c.GetClusterGroupByName(name.(string), GroupContext)
			if err != nil {
				return diag.FromErr(err)
			}
			if group != nil {
				d.SetId(group.UID)
				if err := d.Set("name", group.Name); err != nil {
					return diag.FromErr(err)
				}
			}
		case "project":
			group, err := c.GetClusterGroupByNameForProject(name.(string), GroupContext)
			if err != nil {
				return diag.FromErr(err)
			}
			if group != nil {
				d.SetId(group.Metadata.UID)
				if err := d.Set("name", group.Metadata.Name); err != nil {
					return diag.FromErr(err)
				}
			}
		}
	}
	return diags
}

package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceClusterGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceClusterGroupRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "tenant",
				ValidateFunc: validation.StringInSlice([]string{"", "tenant", "system"}, false),
			},
		},
	}
}

func dataSourceClusterGroupRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	var diags diag.Diagnostics
	if name, okName := d.GetOk("name"); okName {
		GroupContext := d.Get("context").(string)
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
	}
	return diags
}

package spectrocloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/spectrocloud/palette-sdk-go/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUserRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:          schema.TypeString,
				Computed:      true,
				Optional:      true,
				ConflictsWith: []string{"email"},
			},
			"email": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceUserRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	var diags diag.Diagnostics

	if v, ok := d.GetOk("email"); ok {
		user, err := c.GetUserByEmail(v.(string))
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(user.Metadata.UID)
		if err := d.Set("email", user.Spec.EmailID); err != nil {
			return diag.FromErr(err)
		}
	}
	return diags
}

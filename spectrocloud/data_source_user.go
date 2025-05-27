package spectrocloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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
				Description:   "The unique ID of the user. If provided, `email` cannot be used.",
			},
			"email": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The email address of the user. If provided, `id` cannot be used.",
			},
		},
	}
}

func dataSourceUserRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "")
	var diags diag.Diagnostics

	if v, ok := d.GetOk("email"); ok {
		user, err := c.GetUserByEmail(v.(string))
		if err != nil {
			return handleReadError(d, err, diags)
		}
		d.SetId(user.Metadata.UID)
		if err := d.Set("email", user.Spec.EmailID); err != nil {
			return diag.FromErr(err)
		}
	}
	return diags
}

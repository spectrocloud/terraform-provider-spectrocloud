package spectrocloud

import (
	"context"
	"github.com/spectrocloud/palette-sdk-go/api/models"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceRole() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRoleRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:          schema.TypeString,
				Computed:      true,
				Optional:      true,
				ConflictsWith: []string{"name"},
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"permissions": {
				Type:     schema.TypeSet,
				Computed: true,
				Set:      schema.HashString,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "List of permissions associated with the role. ",
			},
		},
	}
}

func dataSourceRoleRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "")
	var diags diag.Diagnostics
	var role *models.V1Role
	var err error
	if i, ok := d.GetOk("id"); ok {
		role, err = c.GetRoleByID(i.(string))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if v, ok := d.GetOk("name"); ok {
		role, err = c.GetRole(v.(string))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if role != nil {
		d.SetId(role.Metadata.UID)
		if err := d.Set("name", role.Metadata.Name); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("permissions", role.Spec.Permissions); err != nil {
			return diag.FromErr(err)
		}
	}
	return diags
}

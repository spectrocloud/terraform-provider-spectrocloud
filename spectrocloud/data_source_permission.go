package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"strings"
)

func dataSourcePermission() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePermissionRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"scope": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"project", "tenant", "resource"}, false),
				Description: "Permission scope. Allowed permission levels are `project` or `tenant` or `resource` . " +
					"Defaults to `project`.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the permissions. eg: `App Deployment`.",
			},
			"permissions": {
				Type:     schema.TypeSet,
				Computed: true,
				Set:      schema.HashString,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "List of permissions associated with the permission name. ",
			},
		},
	}
}

func dataSourcePermissionRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "")
	var diags diag.Diagnostics

	scope := d.Get("scope").(string)
	if v, ok := d.GetOk("name"); ok {
		permission, err := c.GetPermissionByName(v.(string), scope)
		if err != nil {
			return diag.FromErr(err)
		}
		if permission != nil && len(permission.Permissions) > 0 {
			d.SetId(strings.Trim(permission.Name, " "))
			if err := d.Set("name", permission.Name); err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set("permissions", permission.Permissions); err != nil {
				return diag.FromErr(err)
			}
		}
	}
	return diags
}

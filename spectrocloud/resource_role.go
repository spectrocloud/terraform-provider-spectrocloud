package spectrocloud

import (
	"context"
	"fmt"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceRole() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRoleCreate,
		ReadContext:   resourceRoleRead,
		UpdateContext: resourceRoleUpdate,
		DeleteContext: resourceRoleDelete,
		Description:   "The role resource allows you to manage roles in Palette.",

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		SchemaVersion: 2,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the role.",
			},
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"project", "tenant", "resource"}, false),
				Description:  "The role type. Allowed values are `project` or `tenant` or `project`",
			},
			"permissions": {
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The permission's assigned to the role.",
			},
		},
	}
}

func convertInterfaceSliceToStringSlice(input []interface{}) ([]string, error) {
	var output []string
	for _, item := range input {
		str, ok := item.(string)
		if !ok {
			return nil, fmt.Errorf("item %v is not a string", item)
		}
		output = append(output, str)
	}
	return output, nil
}

func toRole(d *schema.ResourceData) *models.V1Role {
	name := d.Get("name").(string)
	roleType := d.Get("type").(string)
	permission, _ := convertInterfaceSliceToStringSlice(d.Get("permissions").(*schema.Set).List())
	return &models.V1Role{
		Metadata: &models.V1ObjectMeta{
			Annotations: map[string]string{
				"scope": roleType,
			},
			LastModifiedTimestamp: models.V1Time{},
			Name:                  name,
		},
		Spec: &models.V1RoleSpec{
			Permissions: permission,
			Scope:       models.V1Scope(roleType),
			Type:        "user",
		},
		Status: &models.V1RoleStatus{
			IsEnabled: true,
		},
	}
}

func flattenRole(d *schema.ResourceData, role *models.V1Role) error {
	var err error
	err = d.Set("name", role.Metadata.Name)
	if err != nil {
		return err
	}
	err = d.Set("type", role.Spec.Scope)
	if err != nil {
		return err
	}
	err = d.Set("permissions", role.Spec.Permissions)
	if err != nil {
		return err
	}
	return nil
}

func resourceRoleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "tenant")
	var diags diag.Diagnostics
	role := toRole(d)
	uid, err := c.CreateRole(role)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(uid)
	return diags
}

func resourceRoleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "tenant")
	var diags diag.Diagnostics
	role, err := c.GetRoleByID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	err = flattenRole(d, role)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceRoleUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "tenant")
	var diags diag.Diagnostics
	role := toRole(d)
	err := c.UpdateRole(role, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceRoleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "tenant")
	var diags diag.Diagnostics
	err := c.DeleteRole(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

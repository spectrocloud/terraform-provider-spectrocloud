package spectrocloud

import (
	"context"
	"time"

	"github.com/spectrocloud/hapi/models"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"
)

func resourceMacro() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMacrosCreate,
		ReadContext:   resourceMacrosRead,
		UpdateContext: resourceMacrosUpdate,
		DeleteContext: resourceMacrosDelete,
		Description:   "A resource for creating and managing service output variables and macros.",

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
				ForceNew:    true,
				Description: "The name of the macro or service variable output.",
			},
			"value": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The value that the macro or service output variable will contain.",
			},
			"project": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The Spectro Cloud project name.",
			},
		},
	}
}

func resourceMacrosCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	var diags diag.Diagnostics
	uid := ""
	var err error
	if v, ok := d.GetOk("project"); ok && v.(string) != "" { //if project name is set it's a project scope
		uid, err = c.GetProjectUID(v.(string))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	err = c.CreateMacros(uid, toMacros(d))
	if err != nil {
		return diag.FromErr(err)
	}
	name := d.Get("name").(string)
	d.SetId(c.GetMacroId(uid, name))
	return diags
}

func resourceMacrosRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	var diags diag.Diagnostics
	var macro *models.V1Macro
	var err error
	uid := ""

	if v, ok := d.GetOk("project"); ok && v.(string) != "" { //if project name is set it's a project scope
		uid, err = c.GetProjectUID(v.(string))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	macro, err = c.GetMacro(d.Get("name").(string), uid)
	if err != nil {
		return diag.FromErr(err)
	} else if macro == nil {
		// Deleted - Terraform will recreate it
		d.SetId("")
		return diags
	}

	d.SetId(c.GetMacroId(uid, d.Get("name").(string)))

	if err := d.Set("name", macro.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("value", macro.Value); err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceMacrosUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	var diags diag.Diagnostics
	var err error
	uid := ""
	if v, ok := d.GetOk("project"); ok && v.(string) != "" { //if project name is set it's a project scope
		uid, err = c.GetProjectUID(v.(string))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if d.HasChange("value") && !d.HasChange("name") {
		err = c.UpdateMacros(uid, toMacros(d))
		if err != nil {
			return diag.FromErr(err)
		}

	}
	return diags
}

func resourceMacrosDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	var diags diag.Diagnostics
	var err error
	uid := ""

	if v, ok := d.GetOk("project"); ok && v.(string) != "" { //if project name is set it's a project scope
		uid, err = c.GetProjectUID(v.(string))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	err = c.DeleteMacros(uid, toMacros(d))
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func toMacros(d *schema.ResourceData) *models.V1Macros {

	var macro []*models.V1Macro
	macro = append(macro, &models.V1Macro{
		Name:  d.Get("name").(string),
		Value: d.Get("value").(string),
	})
	retMacros := &models.V1Macros{
		Macros: macro,
	}
	return retMacros
}

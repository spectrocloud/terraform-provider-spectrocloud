package spectrocloud

import (
	"context"
	"github.com/spectrocloud/palette-sdk-go/client/apiutil"
	"time"

	"github.com/spectrocloud/palette-api-go/models"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceMacro() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMacroCreate,
		ReadContext:   resourceMacroRead,
		UpdateContext: resourceMacroUpdate,
		DeleteContext: resourceMacroDelete,
		Description:   "A resource for creating and managing service output variables and macro.",

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

func resourceMacroCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	c := getV1ClientWithResourceContext(m, "")
	var diags diag.Diagnostics
	uid := ""
	var err error
	if v, ok := d.GetOk("project"); ok && v.(string) != "" { //if project name is set it's a project scope
		uid, err = c.GetProjectUID(v.(string))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	err = c.CreateMacro(uid, toMacro(d))
	if err != nil {
		return diag.FromErr(err)
	}
	name := d.Get("name").(string)
	d.SetId(getMacroId(uid, name))
	return diags
}

func resourceMacroRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	c := getV1ClientWithResourceContext(m, "")
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

	d.SetId(getMacroId(uid, d.Get("name").(string)))

	if err := d.Set("name", macro.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("value", macro.Value); err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceMacroUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	c := getV1ClientWithResourceContext(m, "")
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
		err = c.UpdateMacro(uid, toMacro(d))
		if err != nil {
			return diag.FromErr(err)
		}

	}
	return diags
}

func resourceMacroDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	c := getV1ClientWithResourceContext(m, "")
	var diags diag.Diagnostics
	var err error
	uid := ""

	if v, ok := d.GetOk("project"); ok && v.(string) != "" { //if project name is set it's a project scope
		uid, err = c.GetProjectUID(v.(string))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	err = c.DeleteMacro(uid, toMacro(d))
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func toMacro(d *schema.ResourceData) *models.V1Macros {

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

func getMacroId(uid, name string) string {
	var hash string
	if uid != "" {
		hash = apiutil.StringHash(name + uid)
	} else {
		hash = apiutil.StringHash(name + "%tenant")
	}
	return hash
}

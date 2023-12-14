package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"time"

	"github.com/spectrocloud/hapi/models"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/spectrocloud/palette-sdk-go/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceMacros() *schema.Resource {
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
			"macros": {
				Type:     schema.TypeMap,
				Required: true,
				Set:      schema.HashString,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "A list of macros to be created. macros must be in the form of `macro_name:value`.",
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"", "project", "tenant"}, false),
				Description: "The context of the macro. Allowed values are `\"\"`, `project` or `tenant`. " +
					"Default value is `project`. " + PROJECT_NAME_NUANCE + ". If context is `\"\"` then context set to `tenant` ",
			},
		},
	}
}

func resourceMacrosCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	var diags diag.Diagnostics
	uid := ""
	var err error
	var macroContext string
	//projectUID := ""
	if v, ok := d.GetOk("context"); ok { //if project name is set it's a project scope
		if v == "" {
			macroContext = "tenant"
		} else {
			macroContext = "project"
			//projectUID, err = c.GetProjectUID(v.(string))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}
	err = c.CreateMacrosNew(macroContext, toMacros(d))
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
	macroRD := d.Get("macros").(map[string]interface{})
	for i, j := range macroRD {
		macro = append(macro, &models.V1Macro{
			Name:  macroRD[i].(string),
			Value: j.(string),
		})
	}

	retMacros := &models.V1Macros{
		Macros: macro,
	}
	return retMacros
}

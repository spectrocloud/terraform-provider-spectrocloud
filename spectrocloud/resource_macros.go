package spectrocloud

import (
	"context"
	"time"

	"github.com/spectrocloud/hapi/models"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceMacro() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMacrosCreate,
		ReadContext:   resourceMacrosRead,
		UpdateContext: resourceMacrosUpdate,
		DeleteContext: resourceMacrosDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 2,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"value": {
				Type:     schema.TypeString,
				Required: true,
			},
			"project": {
				Type:     schema.TypeString,
				Optional: true,
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
	//resourceMacrosRead(ctx, d, m)
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
	if macro != nil {
		d.SetId(c.GetMacroId(uid, d.Get("name").(string)))

		if err := d.Set("name", macro.Name); err != nil {
			return diag.FromErr(err)
		}

		if err := d.Set("value", macro.Value); err != nil {
			return diag.FromErr(err)
		}
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
		err = c.PatchMacros(uid, toMacros(d))
		if err != nil {
			return diag.FromErr(err)
		}

	}
	if d.HasChange("name") {
		oldName, _ := d.GetChange("name")
		oldValue, _ := d.GetChange("name")
		deleteMacro := resourceMacro().TestResourceData()
		deleteMacro.Set("name", oldName)
		deleteMacro.Set("value", oldValue)
		err = c.DeleteMacros(uid, toMacros(deleteMacro))
		if err != nil {
			return diag.FromErr(err)
		}
		c.CreateMacros(uid, toMacros(d))
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

// Below is unused code will remove before merge to master
//func toMacrosUpdate(macros []*models.V1Macro, d *schema.ResourceData, old string) *models.V1Macros {
//
//	ret := &models.V1Macro{
//		Name:  d.Get("name").(string),
//		Value: d.Get("value").(string),
//	}
//	put_macros := make([]*models.V1Macro, 0)
//	for _, m := range macros {
//		if m.Name != old {
//			put_macros = append(put_macros, m)
//		}
//	}
//	put_macros = append(put_macros, ret)
//	return &models.V1Macros{
//		Macros: put_macros,
//	}
//}
//
//func macrosExists(macros []*models.V1Macro, name string) bool {
//	for _, macros := range macros {
//		if macros.Name == name {
//			return true
//		}
//	}
//
//	return false
//}

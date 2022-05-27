package spectrocloud

import (
	"context"
	"time"

	"github.com/spectrocloud/hapi/models"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceMacros() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMacrosUpdate,
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
				ForceNew: true,
			},
			"value": {
				Type:     schema.TypeString,
				Optional: true,
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

	macros, err := c.GetMacros(uid)
	if err != nil {
		return diag.FromErr(err)
	}
	err = c.CreateMacros(uid, toMacrosUpdate(macros, d))
	if err != nil {
		return diag.FromErr(err)
	}

	name := d.Get("name").(string)
	d.SetId(c.MacrosHash(name))

	return diags
}

func resourceMacrosRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	var diags diag.Diagnostics
	var macros *models.V1Macro
	var err error
	uid := ""

	if v, ok := d.GetOk("project"); ok && v.(string) != "" { //if project name is set it's a project scope
		uid, err = c.GetProjectUID(v.(string))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	macros, err = c.GetMacro(d.Id(), uid)

	if err != nil {
		return diag.FromErr(err)
	} else if macros == nil {
		// Deleted - Terraform will recreate it
		d.SetId("")
		return diags
	}

	return diags
}

func resourceMacrosUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	var diags diag.Diagnostics
	var err error
	uid := ""

	if d.HasChanges("name") || d.HasChanges("value") {

		if v, ok := d.GetOk("project"); ok && v.(string) != "" { //if project name is set it's a project scope
			uid, err = c.GetProjectUID(v.(string))
			if err != nil {
				return diag.FromErr(err)
			}
		}

		macros, err := c.GetMacros(uid)
		if err != nil {
			return diag.FromErr(err)
		}

		err = c.UpdateMacros(toMacrosUpdate(macros, d).Macros, uid)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	resourceMacrosRead(ctx, d, m)

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

	err = c.DeleteMacros(d.Get("name").(string), uid)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

// toMacros() appends macro to existing list of macros if it is not already there
func toMacros(macros []*models.V1Macro, d *schema.ResourceData) *models.V1Macros {

	ret := &models.V1Macro{
		Name:  d.Get("name").(string),
		Value: d.Get("value").(string),
	}

	if !macrosExists(macros, ret.Name) {
		macros = append(macros, ret)
	}

	return &models.V1Macros{
		Macros: macros,
	}
}

func toMacrosUpdate(macros []*models.V1Macro, d *schema.ResourceData) *models.V1Macros {

	ret := &models.V1Macro{
		Name:  d.Get("name").(string),
		Value: d.Get("value").(string),
	}

	if macrosExists(macros, ret.Name) {
		new_macros := make([]*models.V1Macro, 0)
		for _, macro := range macros {
			if macro.Name != ret.Name {
				new_macros = append(new_macros, macro)
			} else {
				new_macros = append(new_macros, ret)
			}
		}
		return &models.V1Macros{
			Macros: new_macros,
		}
	} else {
		macros = append(macros, ret)
		return &models.V1Macros{
			Macros: macros,
		}
	}

}

func macrosExists(macros []*models.V1Macro, name string) bool {
	for _, macros := range macros {
		if macros.Name == name {
			return true
		}
	}

	return false
}

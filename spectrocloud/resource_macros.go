package spectrocloud

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/apiutil/transport"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
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
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"macros": {
				Type:     schema.TypeMap,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "The key-value mapping includes the macro name and its corresponding value, representing either a macro or a service variable output.",
			},
			"project": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "",
				Description: "The Spectro Cloud project name.",
			},
		},
	}
}

func resourceMacrosCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

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
	macroUID, err := c.CreateMacros(uid, toMacros(d))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(macroUID)
	return diags
}

func resourceMacrosRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	c := getV1ClientWithResourceContext(m, "")
	var diags diag.Diagnostics
	var macros []*models.V1Macro
	var err error
	uid := ""

	if v, ok := d.GetOk("project"); ok && v.(string) != "" { //if project name is set it's a project scope
		uid, err = c.GetProjectUID(v.(string))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	macros, err = c.GetTFMacrosV2(d.Get("macros").(map[string]interface{}), uid)
	if err != nil {
		return diag.FromErr(err)
	} else if len(macros) == 0 {
		// Deleted - Terraform will recreate it
		d.SetId("")
		return diags
	}
	macrosId, err := GetMacrosId(c, uid)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(macrosId)

	retMacros := map[string]interface{}{}

	for _, v := range macros {
		retMacros[v.Name] = v.Value
	}

	if err := d.Set("macros", retMacros); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceMacrosUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
	if d.HasChange("macros") {
		oldMacros, _ := d.GetChange("macros")
		existMacros, _ := c.GetExistMacros(oldMacros.(map[string]interface{}), uid)
		err = c.UpdateMacros(uid, mergeExistingMacros(d, existMacros))
		if err != nil {
			var e *transport.TransportError
			if errors.As(err, &e) && e.HttpCode == 422 {
				if err := d.Set("macros", oldMacros); err != nil {
					return diag.FromErr(err)
				}
				e.Payload.Message = e.Payload.Message + "\n Kindly verify if any of the specified macro names already exist in the system."
				return diag.FromErr(e)
			}
			return diag.FromErr(err)
		}
	}
	return diags
}

func resourceMacrosDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
	err = c.DeleteMacros(uid, toMacros(d))
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func toMacros(d *schema.ResourceData) *models.V1Macros {
	var macro []*models.V1Macro
	dMacros := d.Get("macros").(map[string]interface{})
	for k, v := range dMacros {
		macro = append(macro, &models.V1Macro{
			Name:  k,
			Value: v.(string),
		})
	}
	retMacros := &models.V1Macros{
		Macros: macro,
	}
	return retMacros
}

func mergeExistingMacros(d *schema.ResourceData, existMacros []*models.V1Macro) *models.V1Macros {
	var macro []*models.V1Macro
	dMacros := d.Get("macros").(map[string]interface{})
	for k, v := range dMacros {
		macro = append(macro, &models.V1Macro{
			Name:  k,
			Value: v.(string),
		})
	}
	for _, em := range existMacros {
		macro = append(macro, &models.V1Macro{
			Name:  em.Name,
			Value: em.Value,
		})
	}
	retMacros := &models.V1Macros{
		Macros: macro,
	}
	return retMacros
}

func GetMacrosId(c *client.V1Client, uid string) (string, error) {

	hashId := ""
	if uid != "" {
		hashId = fmt.Sprintf("%s-%s-%s", "project", "macros", uid)
	} else {
		tenantID, err := c.GetTenantUID()
		if err != nil {
			return "", err
		}
		hashId = fmt.Sprintf("%s-%s-%s", "tenant", "macros", tenantID)
	}
	return hashId, nil
}

package spectrocloud

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
		Importer: &schema.ResourceImporter{
			StateContext: resourceMacrosImport,
		},
		Description: "A resource for creating and managing service output variables and macros.",

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
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "tenant",
				ValidateFunc: validation.StringInSlice([]string{"project", "tenant"}, false),
				Description: "The context of the cluster profile. Allowed values are `project` or `tenant`. " +
					"Default value is `project`. " + PROJECT_NAME_NUANCE,
			},
		},
	}
}

func resourceMacrosCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	macrosContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, "")
	var diags diag.Diagnostics
	contextUid := ""
	var err error
	if macrosContext == "project" {
		contextUid = ProviderInitProjectUid
	}
	macroUID, err := c.CreateMacros(contextUid, toMacros(d))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(macroUID)
	return diags
}

func resourceMacrosRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	macrosContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, "")
	var diags diag.Diagnostics
	var macros []*models.V1Macro
	var err error
	contextUid := ""
	if macrosContext == "project" {
		contextUid = ProviderInitProjectUid
	}
	macros, err = c.GetTFMacrosV2(d.Get("macros").(map[string]interface{}), contextUid)
	if err != nil {
		return diag.FromErr(err)
	} else if len(macros) == 0 {
		// Deleted - Terraform will recreate it
		d.SetId("")
		return diags
	}
	macrosId, err := GetMacrosId(c, contextUid)
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
	macrosContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, "")
	var diags diag.Diagnostics
	var err error
	contextUid := ""
	if macrosContext == "project" {
		contextUid = ProviderInitProjectUid
	}
	if d.HasChange("macros") {
		oldMacros, _ := d.GetChange("macros")
		existMacros, _ := c.GetExistMacros(oldMacros.(map[string]interface{}), contextUid)
		err = c.UpdateMacros(contextUid, mergeExistingMacros(d, existMacros))
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
	macrosContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, "")
	var diags diag.Diagnostics
	var err error
	contextUid := ""
	if macrosContext == "project" {
		contextUid = ProviderInitProjectUid
	}
	err = c.DeleteMacros(contextUid, toMacros(d))
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

func resourceMacrosImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {

	rawID := d.Id()
	parts := strings.Split(rawID, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("import ID must be in the format 'id:context' where context is either 'project' or 'tenant'")
	}

	actualID := parts[0]
	context := parts[1]

	if context != "project" && context != "tenant" {
		return nil, fmt.Errorf("context must be either 'project' or 'tenant', got: %s", context)
	}

	if err := d.Set("context", context); err != nil {
		return nil, fmt.Errorf("failed to set context: %w", err)
	}

	d.SetId(actualID)

	c := getV1ClientWithResourceContext(m, "")
	contextUid := ""
	if context == "project" {
		contextUid = ProviderInitProjectUid
	}
	macros, err := c.GetMacrosV2(contextUid)
	if err != nil {
		return nil, fmt.Errorf("could not get macros: %w", err)
	}

	if len(macros) == 0 {
		return nil, fmt.Errorf("no macros found to import for context '%s' with ID '%s'", context, actualID)
	}

	retMacros := map[string]interface{}{}
	for _, v := range macros {
		retMacros[v.Name] = v.Value
	}

	if err := d.Set("macros", retMacros); err != nil {
		return nil, fmt.Errorf("failed to set macros: %w", err)
	}
	return []*schema.ResourceData{d}, nil
}

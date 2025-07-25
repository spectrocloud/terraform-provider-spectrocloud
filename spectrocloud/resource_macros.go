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
		return handleReadError(d, err, diags)
	} else if len(macros) == 0 {
		// Deleted - Terraform will recreate it
		d.SetId("")
		return diags
	}
	macrosId, err := GetMacrosId(c, contextUid)
	if err != nil {
		return handleReadError(d, err, diags)
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
			_ = d.Set("macros", oldMacros)
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

	var diags diag.Diagnostics

	rawIDContext := d.Id()
	parts := strings.Split(rawIDContext, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("import ID must be in the format '{projectUID/tenanatUID}:{project/tenant}'")
	}

	contextID := parts[0]
	macrosContext := parts[1]
	err := ValidateContext(macrosContext)
	if err != nil {
		return nil, err
	}
	err = d.Set("context", macrosContext)
	if err != nil {
		return nil, err
	}

	c := getV1ClientWithResourceContext(m, macrosContext)
	var macros []*models.V1Macro

	if macrosContext == "project" {
		if contextID != ProviderInitProjectUid {
			return nil, fmt.Errorf("invalid import: given project UID {%s} and provider project UID {%s} are different — cross-project resource imports are not allowed; project UID must match the provider configuration", contextID, ProviderInitProjectUid)
		}
		macros, err = c.GetMacros(ProviderInitProjectUid)
		if err != nil {
			return nil, err
		}

	} else {
		actualTenantId, _ := c.GetTenantUID()
		if contextID != actualTenantId {
			return nil, fmt.Errorf("invalid import: tenant UID {%s} does not match your authorized tenant UID {%s}", contextID, actualTenantId)
		}
		macros, err = c.GetMacros("")
		if err != nil {
			return nil, err
		}
	}

	existingMacros := map[string]interface{}{}
	for _, v := range macros {
		existingMacros[v.Name] = v.Value
	}
	err = d.Set("macros", existingMacros)
	if err != nil {
		return nil, err
	}
	macrosId, err := GetMacrosId(c, contextID)
	if err != nil {
		return nil, err
	}
	d.SetId(macrosId)

	if diags.HasError() {
		return nil, fmt.Errorf("could not read password policy for import: %v", diags)
	}
	return []*schema.ResourceData{d}, nil
}

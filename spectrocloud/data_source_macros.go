package spectrocloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

func dataSourceMacros() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMacrosRead,
		Description: "Use this data source to get the ID of a macros resource for use with terraform import.",

		Schema: map[string]*schema.Schema{
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "tenant",
				Description:  "The context to retrieve macros from. Valid values are `project` or `tenant`. Defaults to `tenant`.",
				ValidateFunc: validation.StringInSlice([]string{"project", "tenant"}, false),
			},
			"macros_map": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Map of macros where the key is the macro name and the value is the macro value. ",
			},
			"macro_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name of the macros resource. If specified, the data source will return the macros with this name.",
			},
			"macro_value": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The value of the macros resource. This will be set if `macro_name` is specified.",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "unique identifier for the macros resource, which is the UID of the project or tenant.",
			},
		},
	}
}

func dataSourceMacrosRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	macroContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, "")
	var diags diag.Diagnostics
	var err error

	var uid string
	var macros []*models.V1Macro

	if macroContext == "project" {
		uid = ProviderInitProjectUid
		macros, err = c.GetMacros(uid)
		if err != nil {
			return diag.FromErr(err)
		}
	} else {
		uid, _ = c.GetTenantUID()
		macros, err = c.GetMacros(uid)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if len(macros) == 0 {
		return diag.FromErr(fmt.Errorf("no macros found for context '%s'", macroContext))
	}

	// Convert macros to map
	macroName := d.Get("macro_name").(string)
	macroMap := make(map[string]interface{})
	for _, macro := range macros {
		macroMap[macro.Name] = macro.Value
		if macroName != "" && macro.Name == macroName {
			err := d.Set("macro_value", macro.Value)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}
	if macroName != "" && macroMap[macroName] == nil {
		return diag.FromErr(fmt.Errorf("macro with name '%s' not found", macroName))
	}

	if err := d.Set("macros_map", macroMap); err != nil {
		return diag.FromErr(err)
	}
	// Use the UID as the ID
	uid, _ = c.GetMacrosID(uid)
	d.SetId(uid)

	return diags
}

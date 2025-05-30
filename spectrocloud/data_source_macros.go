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
			"macros": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Map of macro names to their values",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The UID of the macros resource.",
			},
		},
	}
}

func dataSourceMacrosRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "")
	var diags diag.Diagnostics
	var err error

	var uid string
	var macros []*models.V1Macro

	context := d.Get("context").(string)
	if context == "project" {
		uid = ProviderInitProjectUid
		if uid == "" {
			return diag.FromErr(fmt.Errorf("no project context found in provider configuration"))
		}

		macros, err = c.GetMacrosV2(uid)
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to get macros for project: %w", err))
		}
	} else {
		ProviderInitProjectUid = ""
		uid, err = c.GetTenantUID()
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to get tenant UID: %w", err))
		}

		macros, err = c.GetMacrosV2("")
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to get tenant macros: %w", err))
		}
	}

	if len(macros) == 0 {
		return diag.FromErr(fmt.Errorf("no macros found for context '%s'", context))
	}

	// Convert macros to map
	macroMap := make(map[string]interface{})
	for _, macro := range macros {
		macroMap[macro.Name] = macro.Value
	}

	if err := d.Set("macros", macroMap); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set macros: %w", err))
	}

	// Use the UID as the ID
	d.SetId(uid)

	return diags
}

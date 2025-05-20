package spectrocloud

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

func dataSourceMacros() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMacrosRead,
		Description: "Use this data source to get the ID of a macros resource for use with terraform import.",

		Schema: map[string]*schema.Schema{
			"project": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The Spectro Cloud project name. If not specified, macros will be looked up at tenant level.",
			},
			"macros": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "The key-value mapping of macros to look up.",
			},
			"macro_ids": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "A map of macro names to their import IDs (which are just the macro names).",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the macros resource that can be used with terraform import.",
			},
		},
	}
}

func dataSourceMacrosRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[DEBUG] ===== Starting dataSourceMacrosRead =====")

	c := getV1ClientWithResourceContext(m, "")
	log.Printf("[DEBUG] Got client with context")

	var diags diag.Diagnostics
	var err error

	var projectName string
	var uid string
	var importID string
	var macros []*models.V1Macro

	if v, ok := d.GetOk("project"); ok && v.(string) != "" {
		projectName = v.(string)
		log.Printf("[DEBUG] Project name specified: %s", projectName)

		// Get project UID
		log.Printf("[DEBUG] Getting project UID for project: %s", projectName)
		uid, err = c.GetProjectUID(projectName)
		if err != nil {
			log.Printf("[ERROR] Failed to get project UID: %v", err)
			return diag.FromErr(err)
		}
		log.Printf("[DEBUG] Got project UID: %s", uid)
		ProviderInitProjectUid = uid
		log.Printf("[DEBUG] Set ProviderInitProjectUid to: %s", uid)
		importID = "project-macros-" + projectName
		log.Printf("[DEBUG] Set importID to: %s", importID)
		macros, err = c.GetMacrosV2(uid)
		if err != nil {
			log.Printf("[ERROR] Failed to get project macros: %v", err)
			return diag.FromErr(err)
		}
		log.Printf("[DEBUG] Retrieved %d project macros", len(macros))
	} else {
		log.Printf("[DEBUG] No project specified, getting tenant macros")
		ProviderInitProjectUid = ""
		log.Printf("[DEBUG] Reset ProviderInitProjectUid to empty string")

		// Tenant
		log.Printf("[DEBUG] Getting tenant UID")
		uid, err = c.GetTenantUID()
		if err != nil {
			log.Printf("[ERROR] Failed to get tenant UID: %v", err)
			return diag.FromErr(err)
		}
		log.Printf("[DEBUG] Got tenant UID: %s", uid)
		importID = "tenant-macros"
		log.Printf("[DEBUG] Set importID to: %s", importID)
		macros, err = c.GetMacrosV2("")
		if err != nil {
			log.Printf("[ERROR] Failed to get tenant macros: %v", err)
			return diag.FromErr(err)
		}
		log.Printf("[DEBUG] Retrieved %d tenant macros", len(macros))
	}

	// ✅ Build map[string]interface{} to return
	out := make(map[string]interface{}, len(macros))
	macroIDs := make(map[string]interface{}, len(macros))
	for _, macro := range macros {
		out[macro.Name] = macro.Value
		macroIDs[macro.Name] = macro.Name
		log.Printf("[DEBUG] Processing macro: %s = %s", macro.Name, macro.Value)
	}

	log.Printf("[DEBUG] Setting macros in state: %+v", out)
	_ = d.Set("macros", out)

	log.Printf("[DEBUG] Setting macro_ids in state: %+v", macroIDs)
	_ = d.Set("macro_ids", macroIDs)

	log.Printf("[DEBUG] Setting resource ID to: %s", importID)
	d.SetId(importID)

	log.Printf("[DEBUG] ===== dataSourceMacrosRead completed successfully =====")
	return diags
}

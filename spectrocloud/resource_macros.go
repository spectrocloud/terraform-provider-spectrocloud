package spectrocloud

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/api/apiutil/transport"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
)

const (
	projectPrefix = "project-macros-" // keep the trailing dash!
	tenantPrefix  = "tenant-macros"   // no dash because the whole string is the ID
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
	log.Printf("[DEBUG] ===== Starting resourceMacrosImport =====")
	log.Printf("[DEBUG] Input ID: %s", d.Id())

	c := getV1ClientWithResourceContext(m, "")
	log.Printf("[DEBUG] Got client with context: %+v", c)

	rawID := d.Id()
	log.Printf("[DEBUG] Raw ID to process: %s", rawID)

	var contextUID string
	var err error

	switch {
	// -------------------  Project macros  -------------------
	case strings.HasPrefix(rawID, projectPrefix):
		log.Printf("[DEBUG] Processing project macros import")
		// Everything after the prefix is treated as the *literal* project name.
		projectName := rawID[len(projectPrefix):] // safe even if the name contains dashes
		log.Printf("[DEBUG] Extracted project name: %s", projectName)

		if projectName == "" {
			log.Printf("[ERROR] Project name is empty")
			return nil, fmt.Errorf("project name cannot be empty in %q", rawID)
		}

		if err = d.Set("context", "project"); err != nil {
			log.Printf("[ERROR] Failed to set context to project: %v", err)
			return nil, err
		}
		log.Printf("[DEBUG] Set context to project")

		// Translate project name → UID
		log.Printf("[DEBUG] Converting project name to UID: %s", projectName)
		contextUID, err = c.GetProjectUID(projectName)
		if err != nil {
			log.Printf("[ERROR] Failed to get project UID: %v", err)
			return nil, fmt.Errorf("project %q not found: %w", projectName, err)
		}
		log.Printf("[DEBUG] Got project UID: %s", contextUID)

		ProviderInitProjectUid = contextUID
		log.Printf("[DEBUG] Set ProviderInitProjectUid to: %s", contextUID)

		// Get the macros using the project UID
		log.Printf("[DEBUG] Fetching macros for project UID: %s", contextUID)
		macros, err := c.GetMacrosV2(contextUID)
		if err != nil {
			log.Printf("[ERROR] Failed to get macros: %v", err)
			return nil, fmt.Errorf("error getting macros: %w", err)
		}
		log.Printf("[DEBUG] Retrieved %d macros", len(macros))

		// Set the ID using the project UID
		resourceID := fmt.Sprintf("%s-%s-%s", "project", "macros", contextUID)
		log.Printf("[DEBUG] Setting resource ID to: %s", resourceID)
		d.SetId(resourceID)

		// Set the macros in the resource data
		retMacros := make(map[string]interface{}, len(macros))
		for _, v := range macros {
			retMacros[v.Name] = v.Value
			log.Printf("[DEBUG] Processing macro: %s = %s", v.Name, v.Value)
		}
		log.Printf("[DEBUG] Setting macros in resource data: %+v", retMacros)
		if err := d.Set("macros", retMacros); err != nil {
			log.Printf("[ERROR] Failed to set macros: %v", err)
			return nil, fmt.Errorf("error setting macros: %w", err)
		}

	// -------------------  Tenant macros  --------------------
	case rawID == tenantPrefix:
		log.Printf("[DEBUG] Processing tenant macros import")
		if err = d.Set("context", "tenant"); err != nil {
			log.Printf("[ERROR] Failed to set context to tenant: %v", err)
			return nil, err
		}
		log.Printf("[DEBUG] Set context to tenant")

		log.Printf("[DEBUG] Getting tenant UID")
		contextUID, err = c.GetTenantUID()
		if err != nil {
			log.Printf("[ERROR] Failed to get tenant UID: %v", err)
			return nil, fmt.Errorf("cannot determine tenant UID: %w", err)
		}
		log.Printf("[DEBUG] Got tenant UID: %s", contextUID)

		// Get the macros using the tenant UID
		log.Printf("[DEBUG] Fetching macros for tenant UID: %s", contextUID)
		macros, err := c.GetMacrosV2("")
		if err != nil {
			log.Printf("[ERROR] Failed to get macros: %v", err)
			return nil, fmt.Errorf("error getting macros: %w", err)
		}
		log.Printf("[DEBUG] Retrieved %d macros", len(macros))

		// Set the ID using the tenant UID
		resourceID := fmt.Sprintf("%s-%s-%s", "tenant", "macros", contextUID)
		log.Printf("[DEBUG] Setting resource ID to: %s", resourceID)
		d.SetId(resourceID)

		// Set the macros in the resource data
		retMacros := make(map[string]interface{}, len(macros))
		for _, v := range macros {
			retMacros[v.Name] = v.Value
			log.Printf("[DEBUG] Processing macro: %s = %s", v.Name, v.Value)
		}
		log.Printf("[DEBUG] Setting macros in resource data: %+v", retMacros)
		if err := d.Set("macros", retMacros); err != nil {
			log.Printf("[ERROR] Failed to set macros: %v", err)
			return nil, fmt.Errorf("error setting macros: %w", err)
		}

	// -------------------  Invalid format  -------------------
	default:
		log.Printf("[ERROR] Invalid import ID format: %s", rawID)
		return nil, fmt.Errorf(
			`import ID must be either %q<PROJECT-NAME> or exactly %q`,
			projectPrefix, tenantPrefix,
		)
	}

	log.Printf("[DEBUG] ===== Import completed successfully =====")
	return []*schema.ResourceData{d}, nil
}

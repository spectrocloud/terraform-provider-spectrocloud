package spectrocloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceAppliances() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcesApplianceRead,

		Description: "Provides details about a set of appliances used for Edge Native cluster provisioning. " +
			"Various attributes could be used to search for appliances like `tags`, `status`, `health`, and `architecture`.",

		Schema: map[string]*schema.Schema{
			"ids": {
				Type:        schema.TypeList,
				Description: "The unique ids of the appliances. This is a computed field and is not required to be set.",
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"project", "tenant"}, false),
				Description: "The context of the appliances. Allowed values are `project` or `tenant`. " +
					"Defaults to `project`." + PROJECT_NAME_NUANCE,
			},
			"tags": {
				Type:        schema.TypeMap,
				Description: "A list of tags to filter the appliances.",
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"status": {
				Type:        schema.TypeString,
				Description: APPLIANCE_STATUS_DESC + " If not specified, all appliances are returned.",
				Optional:    true,
			},
			"health": {
				Type:        schema.TypeString,
				Description: APPLIANCE_HEALTH_DESC + " If not specified, all appliances are returned.",
				Optional:    true,
			},
			"architecture": {
				Type:        schema.TypeString,
				Description: ARCH_DESC + " If not specified, all appliances are returned.",
				Optional:    true,
			},
		},
	}
}

func dataSourcesApplianceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	// Initialize tags if present
	var tags map[string]string
	if v, ok := d.Get("tags").(map[string]interface{}); ok && len(v) > 0 {
		tags = expandStringMap(v)
	}

	status := d.Get("status").(string)
	health := d.Get("health").(string)
	architecture := d.Get("architecture").(string)

	// Read appliances using the new GetAppliances method
	appliances, err := c.GetAppliances(tags, status, health, architecture)
	if err != nil {
		return diag.FromErr(err)
	}

	// Extract appliance IDs
	var applianceIDs []string
	for _, appliance := range appliances {
		applianceIDs = append(applianceIDs, appliance.Metadata.UID)
	}

	// Set the resource ID and appliance IDs in the schema
	id := toDatasourcesId("appliance", tags)
	d.SetId(id) //need to set some id
	if err := d.Set("ids", applianceIDs); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

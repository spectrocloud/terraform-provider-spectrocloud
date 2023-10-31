package spectrocloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/palette-sdk-go/client"
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

func dataSourcesApplianceRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	var tags map[string]string
	if d.Get("tags") != nil {
		tags = expandStringMap(d.Get("tags").(map[string]interface{}))
	}

	// read all appliances
	appliances, err := c.GetAppliances(nil)
	if err != nil {
		return diag.FromErr(err)
	}

	// prepare filter
	check := func(edgeHostDevice *models.V1EdgeHostsMetadata) bool {
		is := false

		is = is || IsMapSubset(edgeHostDevice.Metadata.Labels, tags)

		if d.Get("status") != nil && d.Get("status").(string) != "" && edgeHostDevice.Status != nil {
			state := string(edgeHostDevice.Status.State)
			is = is && state == d.Get("status").(string)
		}

		if d.Get("health") != nil && d.Get("health").(string) != "" && edgeHostDevice.Status != nil {
			if edgeHostDevice.Status.Health == nil || edgeHostDevice.Status.Health.State == "" {
				return false // if health is not set, it's not a match
			} else {
				is = is && edgeHostDevice.Status.Health.State == d.Get("health").(string)
			}
		}

		if d.Get("architecture") != nil && d.Get("architecture").(string) != "" {
			if edgeHostDevice.Spec == nil || edgeHostDevice.Spec.Device == nil || edgeHostDevice.Spec.Device.ArchType == nil {
				return false
			} else {
				is = is && *edgeHostDevice.Spec.Device.ArchType == d.Get("architecture").(string)
			}
		}
		return is
	}

	// apply filter
	output := filter(appliances.Payload.Items, check)

	//read back all ids
	var applianceIDs []string
	for _, appliance := range output {
		applianceIDs = append(applianceIDs, appliance.Metadata.UID)
	}

	id := toDatasourcesId("appliance", tags)

	d.SetId(id) //need to set some id
	err = d.Set("ids", applianceIDs)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

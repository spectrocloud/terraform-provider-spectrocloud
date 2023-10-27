package spectrocloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/spectrocloud/palette-sdk-go/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const APPLIANCE_STATUS_DESC = "The status of the appliance. Allowed values are: 'ready', 'in-use', and 'unpaired'. "
const APPLIANCE_HEALTH_DESC = "The health of the appliance. Allowed values are: 'healthy', and 'unhealthy'. "
const ARCH_DESC = "The architecture of the appliance. Allowed values are: 'amd64', and  'arm64'. "

func dataSourceAppliance() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceApplianceRead,

		Description: "Provides details about a single appliance.",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the appliance. ",
			},
			"tags": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "The tags of the appliance.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: APPLIANCE_STATUS_DESC,
			},
			"health": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: APPLIANCE_HEALTH_DESC,
			},
			"architecture": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: ARCH_DESC,
			},
		},
	}
}

func dataSourceApplianceRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	var diags diag.Diagnostics
	if name, okName := d.GetOk("name"); okName {
		appliance, err := c.GetApplianceByName(name.(string))
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(appliance.Metadata.UID)
		err = d.Set("name", appliance.Metadata.Name)
		if err != nil {
			return diag.FromErr(err)
		}
		err = d.Set("tags", appliance.Metadata.Labels)
		if err != nil {
			return diag.FromErr(err)
		}

		if appliance.Status != nil {
			err = d.Set("status", appliance.Status.State)
			if err != nil {
				return diag.FromErr(err)
			}
		}

		if appliance.Status != nil && appliance.Status.Health != nil {
			err = d.Set("health", appliance.Status.Health.State)
			if err != nil {
				return diag.FromErr(err)
			}
		}

		if appliance.Spec != nil && appliance.Spec.Device != nil && appliance.Spec.Device.ArchType != nil {
			err = d.Set("architecture", *appliance.Spec.Device.ArchType)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}
	return diags
}

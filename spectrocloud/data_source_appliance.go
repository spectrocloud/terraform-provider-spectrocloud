package spectrocloud

import (
	"context"
	"github.com/spectrocloud/palette-sdk-go/api/models"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const APPLIANCE_STATUS_DESC = "The status of the appliance. Supported values are: 'ready', 'in-use', and 'unpaired'. "
const APPLIANCE_HEALTH_DESC = "The health of the appliance. Supported values are: 'healthy', and 'unhealthy'. "
const ARCH_DESC = "The architecture of the appliance. Supported values are: 'amd64', and  'arm64'. "

func dataSourceAppliance() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceApplianceRead,

		Description: "Provides details about a single appliance used for Edge Native cluster provisioning.",
		Schema: map[string]*schema.Schema{
			"id": {
				Type:         schema.TypeString,
				Description:  "ID of the appliance registered in Palette.",
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"id", "name"},
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
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
	c := getV1ClientWithResourceContext(m, "")
	var diags diag.Diagnostics
	var err error
	var appliance *models.V1EdgeHostDevice
	if id, okId := d.GetOk("id"); okId {
		appliance, err = c.GetAppliance(id.(string))
		if err != nil {
			return handleReadError(d, err, diags)
		}
	}
	if name, okName := d.GetOk("name"); okName {
		appliance, err = c.GetApplianceByName(name.(string), nil, "", "", "")
		if err != nil {
			return handleReadError(d, err, diags)
		}
	}
	if appliance != nil {
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

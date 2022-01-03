package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAppliance() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceApplianceRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceApplianceRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	var diags diag.Diagnostics
	if name, okName := d.GetOk("name"); okName {
		if projectId, okProject := d.GetOk("project_id"); okProject {
			appliance, err := c.GetApplianceByName(projectId.(string), name.(string))
			if err != nil {
				return diag.FromErr(err)
			}
			d.SetId(appliance.Metadata.UID)
			d.Set("name", appliance.Metadata.Name)
			d.Set("project_id", appliance.Aclmeta.ProjectUID)
		}
	}
	return diags
}

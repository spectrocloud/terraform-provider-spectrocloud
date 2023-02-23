package spectrocloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/spectrocloud/palette-sdk-go/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceApplicationProfile() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceApplicationProfileRead,
		Description: "Use this data source to get the details of an existing application profile.",

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the application profile",
			},
			"version": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "The version of the app profile. Default value is '1.0.0'.",
			},
		},
	}
}

func dataSourceApplicationProfileRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	var diags diag.Diagnostics
	if name, okName := d.GetOk("name"); okName {
		version, okVersion := d.GetOk("version")
		if !okVersion || version == "" {
			version = "1.0.0"
		}
		applicationProfile, appUID, getVersion, err := c.GetApplicationProfileByNameAndVersion(name.(string), version.(string))
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(appUID)
		if err := d.Set("name", applicationProfile.Metadata.Name); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("version", getVersion); err != nil {
			return diag.FromErr(err)
		}
	}
	return diags
}

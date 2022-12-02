package spectrocloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"

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
		},
	}
}

func dataSourceApplicationProfileRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	var diags diag.Diagnostics
	if name, okName := d.GetOk("name"); okName {
		application_profile, err := c.GetApplicationProfileByName(name.(string))
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(application_profile.Metadata.UID)
		d.Set("name", application_profile.Metadata.Name)
	}
	return diags
}

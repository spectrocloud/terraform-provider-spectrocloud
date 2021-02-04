package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceClusterProfile() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceClusterProfileRead,

		Schema: map[string]*schema.Schema{
			// TODO packs
			id: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{id, name},
			},
			name: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{id, name},
			},
		},
	}
}

func dataSourceClusterProfileRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1alpha1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	profiles, err := c.GetClusterProfiles()
	if err != nil {
		return diag.FromErr(err)
	}

	var profile *models.V1alpha1ClusterProfile
	for _, p := range profiles {

		if v, ok := d.GetOk(id); ok && v.(string) == p.Metadata.UID {
			profile = p
			break
		} else if v, ok := d.GetOk(name); ok && v.(string) == p.Metadata.Name {
			profile = p
			break
		}
	}

	if profile == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to find cluster profile",
			Detail:   "Unable to find the specified cluster profile",
		})
		return diags
	}

	d.SetId(profile.Metadata.UID)
	d.Set(name, profile.Metadata.Name)

	return diags
}
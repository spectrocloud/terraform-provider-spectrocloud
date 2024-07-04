package spectrocloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-api-go/models"
)

func dataSourceCloudAccountMaas() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCloudAccountMaasRead,

		Schema: map[string]*schema.Schema{
			"maas_api_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"maas_api_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"id", "name"},
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"id", "name"},
			},
		},
	}
}

func dataSourceCloudAccountMaasRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// sdk cut over context handling
	c := GetResourceLevelV1Client(m, "")

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	accounts, err := c.GetCloudAccountsMaas()
	if err != nil {
		return diag.FromErr(err)
	}

	var account *models.V1MaasAccount
	for _, a := range accounts {

		if v, ok := d.GetOk("id"); ok && v.(string) == a.Metadata.UID {
			account = a
			break
		} else if v, ok := d.GetOk("name"); ok && v.(string) == a.Metadata.Name {
			account = a
			break
		}
	}

	if account == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to find maas cloud account",
			Detail:   "Unable to find the specified maas cloud account",
		})
		return diags
	}

	d.SetId(account.Metadata.UID)
	err = d.Set("name", account.Metadata.Name)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("maas_api_endpoint", account.Spec.APIEndpoint)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("maas_api_key", account.Spec.APIKey)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

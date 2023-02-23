package spectrocloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/spectrocloud/hapi/client"
	"github.com/spectrocloud/hapi/models"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceCloudAccountAzure() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCloudAccountAzureRead,

		Schema: map[string]*schema.Schema{
			"azure_tenant_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"azure_client_id": {
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

func dataSourceCloudAccountAzureRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	accounts, err := c.GetCloudAccountsAzure()
	if err != nil {
		return diag.FromErr(err)
	}

	var account *models.V1AzureAccount
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
			Summary:  "Unable to find azure cloud account",
			Detail:   "Unable to find the specified azure cloud account",
		})
		return diags
	}

	d.SetId(account.Metadata.UID)
	d.Set("name", account.Metadata.Name)
	d.Set("azure_tenant_id", *account.Spec.TenantID)
	d.Set("azure_client_id", *account.Spec.ClientID)

	return diags
}

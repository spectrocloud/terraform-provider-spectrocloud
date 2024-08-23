package spectrocloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

func dataSourceCloudAccountAzure() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCloudAccountAzureRead,
		Description: "A data source for retrieving information about an Azure cloud account registered in Palette.",

		Schema: map[string]*schema.Schema{
			"azure_tenant_id": {
				Type:        schema.TypeString,
				Description: "The tenant ID of the Azure cloud account registered in Palette.",
				Computed:    true,
			},
			"azure_client_id": {
				Type:        schema.TypeString,
				Description: "The unique client ID from Azure Management Portal.",
				Computed:    true,
			},
			"id": {
				Type:         schema.TypeString,
				Description:  "ID of the Azure cloud account registered in Palette.",
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"id", "name"},
			},
			"name": {
				Type:         schema.TypeString,
				Description:  "Name of the Azure cloud account registered in Palette.",
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"id", "name"},
			},
			"tenant_name": {
				Description: "The name of the Azure tenant.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"disable_properties_request": {
				Type:        schema.TypeBool,
				Description: "The status of the disable properties option.",
				Computed:    true,
			},
		},
	}
}

func dataSourceCloudAccountAzureRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "")

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
	if err := d.Set("name", account.Metadata.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("azure_tenant_id", *account.Spec.TenantID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("azure_client_id", *account.Spec.ClientID); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

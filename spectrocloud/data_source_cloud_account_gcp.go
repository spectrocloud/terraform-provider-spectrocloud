package spectrocloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

func dataSourceCloudAccountGcp() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCloudAccountGcpRead,
		Description: "A data source for retrieving information about a GCP cloud account registered in Palette.",

		Schema: map[string]*schema.Schema{
			"id": {
				Type:         schema.TypeString,
				Description:  "ID of the GCP cloud account registered in Palette.",
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"id", "name"},
			},
			"name": {
				Type:         schema.TypeString,
				Description:  "Name of the GCP cloud account registered in Palette.",
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"id", "name"},
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "",
				ValidateFunc: validation.StringInSlice([]string{"", "project", "tenant"}, false),
				Description:  "The context of the cluster. Allowed values are `project` or `tenant` or ``. ",
			},
		},
	}
}

func dataSourceCloudAccountGcpRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "")

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	accounts, err := c.GetCloudAccountsGcp()
	if err != nil {
		return handleReadError(d, err, diags)
	}

	var account *models.V1GcpAccount
	filteredAccounts := make([]*models.V1GcpAccount, 0)
	for _, a := range accounts {

		if v, ok := d.GetOk("id"); ok && v.(string) == a.Metadata.UID {
			account = a
			break
		} else if v, ok := d.GetOk("name"); ok && v.(string) == a.Metadata.Name {
			filteredAccounts = append(filteredAccounts, a)
		}
	}

	if len(filteredAccounts) > 1 {
		if accContext, ok := d.GetOk("context"); ok && accContext != "" {
			for _, ac := range filteredAccounts {
				if ac.Metadata.Annotations["scope"] == accContext {
					account = ac
					break
				}
			}
		} else {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Found multiple cloud accounts",
				Detail:   fmt.Sprintf("more than 1 account found for name - '%s'. Kindly re-try with `context` set, Allowed value `project` or `tenant`", d.Get("name").(string)),
			})
			return diags
		}
	} else if len(filteredAccounts) == 1 {
		account = filteredAccounts[0]
	}

	if account == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to find gcp cloud account",
			Detail:   "Unable to find the specified gcp cloud account",
		})
		return diags
	}

	d.SetId(account.Metadata.UID)
	if err := d.Set("name", account.Metadata.Name); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

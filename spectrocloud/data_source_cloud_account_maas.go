package spectrocloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

func dataSourceCloudAccountMaas() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCloudAccountMaasRead,

		Schema: map[string]*schema.Schema{
			"maas_api_endpoint": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The API endpoint of the MaaS account. This value is computed based on the cloud account's configuration and is used for interaction with the MaaS service.",
			},
			"maas_api_key": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The API key associated with the MaaS account. This is used to authenticate API requests to the MaaS service and is computed from the cloud account's credentials.",
			},
			"id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"id", "name"},
				Description:  "The unique ID of the cloud account. Either `id` or `name` must be provided, but not both.",
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"id", "name"},
				Description:  "The name of the cloud account. This can be used instead of `id` to retrieve the account details. Only one of `id` or `name` can be specified.",
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

func dataSourceCloudAccountMaasRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "")

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	accounts, err := c.GetCloudAccountsMaas()
	if err != nil {
		return handleReadError(d, err, diags)
	}

	var account *models.V1MaasAccount
	filteredAccounts := make([]*models.V1MaasAccount, 0)
	for _, a := range accounts {
		if v, ok := d.GetOk("id"); ok && v.(string) == a.Metadata.UID {
			account = a
			break
		} else if v, ok := d.GetOk("name"); ok && v.(string) == a.Metadata.Name {
			account = a
			break
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
		account = filteredAccounts[:1][0]
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

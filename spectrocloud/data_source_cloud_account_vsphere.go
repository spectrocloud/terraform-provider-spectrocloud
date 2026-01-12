package spectrocloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

func dataSourceCloudAccountVsphere() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCloudAccountVsphereRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"id", "name"},
				Description:  "The unique ID of the vSphere cloud account. Either `id` or `name` must be provided, but not both.",
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"id", "name"},
				Description:  "The name of the vSphere cloud account. Either `id` or `name` must be provided, but not both.",
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

func dataSourceCloudAccountVsphereRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "")

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	accounts, err := c.GetCloudAccountsVsphere()
	if err != nil {
		return handleReadError(d, err, diags)
	}

	var account *models.V1VsphereAccount
	filteredAccounts := make([]*models.V1VsphereAccount, 0)
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
		account = filteredAccounts[0]
	}

	if account == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to find vsphere cloud account",
			Detail:   "Unable to find the specified vsphere cloud account",
		})
		return diags
	}

	d.SetId(account.Metadata.UID)
	if err := d.Set("name", account.Metadata.Name); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

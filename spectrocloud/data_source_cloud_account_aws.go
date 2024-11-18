package spectrocloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

func dataSourceCloudAccountAws() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCloudAccountAwsRead,
		Description: "A data source for retrieving information about an AWS cloud account registered in Palette.",

		Schema: map[string]*schema.Schema{
			"id": {
				Type:         schema.TypeString,
				Description:  "ID of the AWS cloud account registered in Palette.",
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"id", "name"},
			},
			"name": {
				Type:         schema.TypeString,
				Description:  "Name of the AWS cloud account registered in Palette.",
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
			"depends": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"id", "name"},
			},
		},
	}
}

func dataSourceCloudAccountAwsRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "")

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	accounts, err := c.GetCloudAccountsAws()
	if err != nil {
		return diag.FromErr(err)
	}
	var fAccount *models.V1AwsAccount
	filteredAccounts := make([]*models.V1AwsAccount, 0)
	for _, acc := range accounts {
		if v, ok := d.GetOk("id"); ok && v.(string) == acc.Metadata.UID {
			fAccount = acc
			break
		}
		if v, ok := d.GetOk("name"); ok && v.(string) == acc.Metadata.Name {
			filteredAccounts = append(filteredAccounts, acc)
		}
	}
	if len(filteredAccounts) > 1 {
		if accContext, ok := d.GetOk("context"); ok && accContext != "" {
			for _, ac := range filteredAccounts {
				if ac.Metadata.Annotations["scope"] == accContext {
					fAccount = ac
					break
				}
			}
		} else {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Found more than 1 AWS account",
				Detail:   fmt.Sprintf("more than 1 aws account found for name - '%s'. Kindly re-try with `context` set, Allowed value `project` or `tenant`", d.Get("name").(string)),
			})
			return diags
		}
	} else if len(filteredAccounts) == 1 {
		fAccount = filteredAccounts[0]
	}

	if fAccount == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to find aws cloud account",
			Detail:   "Unable to find the specified aws cloud account",
		})
		return diags
	}

	d.SetId(fAccount.Metadata.UID)
	err = d.Set("name", fAccount.Metadata.Name)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

//func dataSourceCloudAccountAwsRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
//	c := getV1ClientWithResourceContext(m, "")
//
//	// Warning or errors can be collected in a slice type
//	var diags diag.Diagnostics
//
//	accounts, err := c.GetCloudAccountsAws()
//	if err != nil {
//		return diag.FromErr(err)
//	}
//
//	var account *models.V1AwsAccount
//	for _, a := range accounts {
//
//		if v, ok := d.GetOk("id"); ok && v.(string) == a.Metadata.UID {
//			account = a
//			break
//		} else if v, ok := d.GetOk("name"); ok && v.(string) == a.Metadata.Name {
//			if v, ok := d.GetOk("context"); ok && v.(string) == a.Metadata.Annotations["scope"] {
//				account = a
//				break
//			}
//		}
//	}
//
//	if account == nil {
//		diags = append(diags, diag.Diagnostic{
//			Severity: diag.Error,
//			Summary:  "Unable to find aws cloud account",
//			Detail:   "Unable to find the specified aws cloud account",
//		})
//		return diags
//	}
//
//	d.SetId(account.Metadata.UID)
//	err = d.Set("name", account.Metadata.Name)
//	if err != nil {
//		return diag.FromErr(err)
//	}
//	err = d.Set("context", account.Metadata.Annotations["scope"])
//	if err != nil {
//		return diag.FromErr(err)
//	}
//
//	return diags
//}

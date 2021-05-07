package spectrocloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceCloudAccountAws() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCloudAccountAwsRead,

		Schema: map[string]*schema.Schema{
			"aws_access_key": {
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

func dataSourceCloudAccountAwsRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1alpha1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	accounts, err := c.GetCloudAccountsAws()
	if err != nil {
		return diag.FromErr(err)
	}

	var account *models.V1alpha1AwsAccount
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
			Summary:  "Unable to find aws cloud account",
			Detail:   "Unable to find the specified aws cloud account",
		})
		return diags
	}

	d.SetId(account.Metadata.UID)
	d.Set("name", account.Metadata.Name)
	d.Set("aws_access_key", account.Spec.AccessKey)

	return diags
}

package spectrocloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-api-go/models"
)

func dataSourceCloudAccountTencent() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCloudAccountTencentRead,

		Schema: map[string]*schema.Schema{
			"tencent_secret_id": {
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

func dataSourceCloudAccountTencentRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// sdk cut over context handling
	c := GetResourceLevelV1Client(m, "")

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	accounts, err := c.GetCloudAccountsTke()
	if err != nil {
		return diag.FromErr(err)
	}

	var account *models.V1TencentAccount
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
			Summary:  "Unable to find tencent cloud account",
			Detail:   "Unable to find the specified tencent cloud account",
		})
		return diags
	}

	d.SetId(account.Metadata.UID)
	if err := d.Set("name", account.Metadata.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("tencent_secret_id", *account.Spec.SecretID); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

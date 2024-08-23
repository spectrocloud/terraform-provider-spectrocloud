package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

func dataSourceCloudAccountCustom() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCloudAccountCustomRead,

		Schema: map[string]*schema.Schema{
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
			"cloud": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The cloud provider name.",
			},
		},
	}
}

func dataSourceCloudAccountCustomRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "")
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	cloudType := d.Get("cloud").(string)

	accounts, err := c.GetCustomCloudAccountList(cloudType)
	if err != nil {
		return diag.FromErr(err)
	}
	var account *models.V1CustomAccount
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
	err = d.Set("name", account.Metadata.Name)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

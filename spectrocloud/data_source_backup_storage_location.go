package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

func dataSourceBackupStorageLocation() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceBackupStorageLocationRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"id", "name"},
				Description:  "The unique ID of the backup storage location. This is an optional field, but if provided, it will be used to retrieve the specific backup storage location.",
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"id", "name"},
				Description:  "The name of the backup storage location. This is an optional field, but if provided, it will be used to retrieve the specific backup storage location.",
			},
		},
	}
}

func dataSourceBackupStorageLocationRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "project")

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	//projectScope := true

	bsls, err := c.ListBackupStorageLocation()
	if err != nil {
		return handleReadError(d, err, diags)
	}

	var bsl *models.V1UserAssetsLocation
	for _, a := range bsls {
		if v, ok := d.GetOk("id"); ok && v.(string) == a.Metadata.UID {
			bsl = a
			break
		} else if v, ok := d.GetOk("name"); ok && v.(string) == a.Metadata.Name {
			bsl = a
			break
		}
	}

	if bsl == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to find backup storage location",
			Detail:   "Unable to find the specified backup storage location",
		})
		return diags
	}

	d.SetId(bsl.Metadata.UID)
	err = d.Set("name", bsl.Metadata.Name)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

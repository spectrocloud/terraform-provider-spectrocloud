package spectrocloud

import (
	"context"
	"fmt"
	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"time"
)

func dataSourceRegistrationToken() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRegistrationTokenRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:          schema.TypeString,
				Computed:      true,
				Optional:      true,
				Description:   "The UID of the registration token.",
				ConflictsWith: []string{"name"},
			},
			"name": {
				Type:         schema.TypeString,
				Description:  "The name of the registration token.",
				Optional:     true,
				AtLeastOneOf: []string{"name", "id"},
			},
			"project_uid": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The unique identifier of the project associated with the registration token.",
			},
			"expiry_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The expiration date of the registration token in `YYYY-MM-DD` format.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the registration token. Allowed values are `active` or `inactive`. Default is `active`.",
			},
			"token": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The registration token",
			},
		},
	}
}

func dataSourceRegistrationTokenRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, tenantString)
	var diags diag.Diagnostics
	var err error
	var tokenEntity *models.V1EdgeToken
	if name, okName := d.GetOk("name"); okName {
		tokenEntity, err = c.GetRegistrationTokenByName(name.(string))
		if err != nil {
			return handleReadError(d, err, diags)
		}
	} else if id, okId := d.GetOk("id"); okId {
		tokenEntity, err = c.GetRegistrationTokenByUID(id.(string))
		if err != nil {
			return handleReadError(d, err, diags)
		}
	}
	if tokenEntity != nil {
		d.SetId(tokenEntity.Metadata.UID)
		if err := d.Set("name", tokenEntity.Metadata.Name); err != nil {
			return diag.FromErr(err)
		}

		if tokenEntity.Spec.DefaultProject != nil {
			if err := d.Set("project_uid", tokenEntity.Spec.DefaultProject.UID); err != nil {
				return diag.FromErr(err)
			}
		}

		dt := strfmt.DateTime(tokenEntity.Spec.Expiry)
		expDate := time.Time(dt).Format("2006-01-02")
		if err := d.Set("expiry_date", expDate); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("token", tokenEntity.Spec.Token); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("status", StateConvertBool(tokenEntity.Status.IsActive)); err != nil {
			return diag.FromErr(err)
		}
	} else {
		return diag.FromErr(fmt.Errorf("could not find registration token: %v", diags))
	}

	return diags
}

package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/gomi/pkg/ptr"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"
)

func resourceCloudAccountTencent() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCloudAccountTencentCreate,
		ReadContext:   resourceCloudAccountTencentRead,
		UpdateContext: resourceCloudAccountTencentUpdate,
		DeleteContext: resourceCloudAccountTencentDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"tencent_secret_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tencent_secret_key": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceCloudAccountTencentCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	var diags diag.Diagnostics

	account := toTencentAccount(d)

	uid, err := c.CreateCloudAccountTke(account)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(uid)

	resourceCloudAccountTencentRead(ctx, d, m)

	return diags
}

func resourceCloudAccountTencentRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	var diags diag.Diagnostics

	uid := d.Id()

	account, err := c.GetCloudAccountTke(uid)
	if err != nil {
		return diag.FromErr(err)
	} else if account == nil {
		d.SetId("")
		return diags
	}

	if err := d.Set("name", account.Metadata.Name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("tencent_secret_id", account.Spec.SecretID); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

//
func resourceCloudAccountTencentUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	account := toTencentAccount(d)

	err := c.UpdateCloudAccountTencent(account)
	if err != nil {
		return diag.FromErr(err)
	}

	resourceCloudAccountTencentRead(ctx, d, m)

	return diags
}

func resourceCloudAccountTencentDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	var diags diag.Diagnostics

	cloudAccountID := d.Id()

	err := c.DeleteCloudAccountTke(cloudAccountID)
	if err != nil {
		return diag.FromErr(err)
	}

	// d.SetId("") is automatically called assuming delete returns no errors

	return diags
}

func toTencentAccount(d *schema.ResourceData) *models.V1TencentAccount {
	account := &models.V1TencentAccount{
		Metadata: &models.V1ObjectMeta{
			Name: d.Get("name").(string),
			UID:  d.Id(),
		},
		Spec: &models.V1TencentCloudAccount{
			SecretID:  ptr.StringPtr(d.Get("tencent_secret_id").(string)),
			SecretKey: ptr.StringPtr(d.Get("tencent_secret_key").(string)),
		},
	}

	return account
}

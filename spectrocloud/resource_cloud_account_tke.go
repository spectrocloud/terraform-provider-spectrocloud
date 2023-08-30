package spectrocloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/palette-sdk-go/client"

	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
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
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"", "project", "tenant"}, false),
				Description:  "The context of the Tencent configuration. Can be `project` or `tenant`.",
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
	AccountContext := d.Get("context").(string)
	uid, err := c.CreateCloudAccountTke(account, AccountContext)
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
	AccountContext := d.Get("context").(string)
	account, err := c.GetCloudAccountTke(uid, AccountContext)
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
	AccountContext := d.Get("context").(string)
	err := c.DeleteCloudAccountTke(cloudAccountID, AccountContext)
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
			SecretID:  types.Ptr(d.Get("tencent_secret_id").(string)),
			SecretKey: types.Ptr(d.Get("tencent_secret_key").(string)),
		},
	}

	return account
}

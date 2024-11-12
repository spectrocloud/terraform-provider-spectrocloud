package spectrocloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/api/models"

	"github.com/spectrocloud/terraform-provider-spectrocloud/util/ptr"
)

func resourceCloudAccountTencent() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCloudAccountTencentCreate,
		ReadContext:   resourceCloudAccountTencentRead,
		UpdateContext: resourceCloudAccountTencentUpdate,
		DeleteContext: resourceCloudAccountTencentDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the Tencent account to be managed.",
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"", "project", "tenant"}, false),
				Description: "The context of the Tencent account. Allowed values are `project` or `tenant`. " +
					"Default value is `project`. " + PROJECT_NAME_NUANCE,
			},
			"tencent_secret_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The secret ID associated with the Tencent account for authentication.",
			},
			"tencent_secret_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "The secret key associated with the Tencent account for authentication.",
			},
		},
	}
}

func resourceCloudAccountTencentCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

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
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

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

func resourceCloudAccountTencentUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	account := toTencentAccount(d)

	err := c.UpdateCloudAccountTke(account)
	if err != nil {
		return diag.FromErr(err)
	}

	resourceCloudAccountTencentRead(ctx, d, m)

	return diags
}

func resourceCloudAccountTencentDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

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
			SecretID:  ptr.To(d.Get("tencent_secret_id").(string)),
			SecretKey: ptr.To(d.Get("tencent_secret_key").(string)),
		},
	}

	return account
}

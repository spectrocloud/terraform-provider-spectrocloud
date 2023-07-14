package spectrocloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/palette-sdk-go/client"
)

func resourceCloudAccountMaas() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCloudAccountMaasCreate,
		ReadContext:   resourceCloudAccountMaasRead,
		UpdateContext: resourceCloudAccountMaasUpdate,
		DeleteContext: resourceCloudAccountMaasDelete,
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
				Description:  "The context of the MAAS configuration. Can be `project` or `tenant`.",
			},
			"private_cloud_gateway_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"maas_api_endpoint": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"maas_api_key": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceCloudAccountMaasCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	account := toMaasAccount(d)
	AccountContext := d.Get("context").(string)
	uid, err := c.CreateCloudAccountMaas(account, AccountContext)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(uid)

	resourceCloudAccountMaasRead(ctx, d, m)

	return diags
}

func resourceCloudAccountMaasRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	var diags diag.Diagnostics

	uid := d.Id()
	AccountContext := d.Get("context").(string)
	account, err := c.GetCloudAccountMaas(uid, AccountContext)
	if err != nil {
		return diag.FromErr(err)
	} else if account == nil {
		// Deleted - Terraform will recreate it
		d.SetId("")
		return diags
	}

	if err := d.Set("name", account.Metadata.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("private_cloud_gateway_id", account.Metadata.Annotations[OverlordUID]); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceCloudAccountMaasUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	account := toMaasAccount(d)
	err := c.UpdateCloudAccountMaas(account)
	if err != nil {
		return diag.FromErr(err)
	}

	resourceCloudAccountMaasRead(ctx, d, m)

	return diags
}

func resourceCloudAccountMaasDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	var diags diag.Diagnostics

	cloudAccountID := d.Id()
	AccountContext := d.Get("context").(string)
	err := c.DeleteCloudAccountMaas(cloudAccountID, AccountContext)
	if err != nil {
		return diag.FromErr(err)
	}

	// d.SetId("") is automatically called assuming delete returns no errors

	return diags
}

func toMaasAccount(d *schema.ResourceData) *models.V1MaasAccount {
	EndpointVal := d.Get("maas_api_endpoint").(string)
	KeyVal := d.Get("maas_api_key").(string)
	account := &models.V1MaasAccount{
		Metadata: &models.V1ObjectMeta{
			Name: d.Get("name").(string),
			UID:  d.Id(),
		},
		Spec: &models.V1MaasCloudAccount{
			APIEndpoint: &EndpointVal,
			APIKey:      &KeyVal,
		},
	}

	return account
}

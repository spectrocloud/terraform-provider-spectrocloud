package spectrocloud

import (
	"context"
	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

func resourceCloudAccountAzure() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCloudAccountAzureCreate,
		ReadContext:   resourceCloudAccountAzureRead,
		UpdateContext: resourceCloudAccountAzureUpdate,
		DeleteContext: resourceCloudAccountAzureDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"azure_tenant_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"azure_client_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"azure_client_secret": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
				//DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
				//	return false
				//},
				//StateFunc: func(val interface{}) string {
				//	return strings.ToLower(val.(string))
				//},
			},
		},
	}
}

func resourceCloudAccountAzureCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	account := toAzureAccount(d)

	uid, err := c.CreateCloudAccountAzure(account)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(uid)

	resourceCloudAccountAzureRead(ctx, d, m)

	return diags
}

func resourceCloudAccountAzureRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	var diags diag.Diagnostics

	uid := d.Id()

	account, err := c.GetCloudAccountAzure(uid)
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
	if err := d.Set("azure_tenant_id", *account.Spec.TenantID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("azure_client_id", *account.Spec.ClientID); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceCloudAccountAzureUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	account := toAzureAccount(d)

	err := c.UpdateCloudAccountAzure(account)
	if err != nil {
		return diag.FromErr(err)
	}

	resourceCloudAccountAzureRead(ctx, d, m)

	return diags
}

func resourceCloudAccountAzureDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	var diags diag.Diagnostics

	cloudAccountID := d.Id()

	err := c.DeleteCloudAccountAzure(cloudAccountID)
	if err != nil {
		return diag.FromErr(err)
	}

	// d.SetId("") is automatically called assuming delete returns no errors

	return diags
}

func toAzureAccount(d *schema.ResourceData) *models.V1AzureAccount {
	clientSecret := strfmt.Password(d.Get("azure_client_secret").(string)).String()
	account := &models.V1AzureAccount{
		Metadata: &models.V1ObjectMeta{
			Name: d.Get("name").(string),
			UID:  d.Id(),
		},
		Spec: &models.V1AzureCloudAccount{
			ClientID:     types.Ptr(d.Get("azure_client_id").(string)),
			ClientSecret: &clientSecret,
			TenantID:     types.Ptr(d.Get("azure_tenant_id").(string)),
		},
	}
	return account
}

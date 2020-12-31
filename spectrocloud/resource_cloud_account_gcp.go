package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"
)

func resourceCloudAccountGcp() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCloudAccountGcpCreate,
		ReadContext:   resourceCloudAccountGcpRead,
		UpdateContext: resourceCloudAccountGcpUpdate,
		DeleteContext: resourceCloudAccountGcpDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"gcp_json_credentials": {
				Type:     schema.TypeString,
				Required: true,
				Sensitive: true,
			},
		},
	}
}

func resourceCloudAccountGcpCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1alpha1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	account := toGcpAccount(d)

	uid, err := c.CreateCloudAccountGcp(account)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(uid)

	resourceCloudAccountGcpRead(ctx, d, m)

	return diags
}

func resourceCloudAccountGcpRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1alpha1Client)

	var diags diag.Diagnostics

	uid := d.Id()

	account, err := c.GetCloudAccountGcp(uid)
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

	return diags
}

//
func resourceCloudAccountGcpUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1alpha1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	//time.Sleep(20 * time.Second)
	account := toGcpAccount(d)

	err := c.UpdateCloudAccountGcp(account)
	if err != nil {
		return diag.FromErr(err)
	}

	resourceCloudAccountGcpRead(ctx, d, m)

	return diags
}

func resourceCloudAccountGcpDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1alpha1Client)

	var diags diag.Diagnostics

	cloudAccountID := d.Id()

	err := c.DeleteCloudAccountGcp(cloudAccountID)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func toGcpAccount(d *schema.ResourceData) *models.V1alpha1GcpAccountEntity {
	account := &models.V1alpha1GcpAccountEntity{
		Metadata: &models.V1ObjectMeta{
			Name: d.Get("name").(string),
			UID : d.Id(),
		},
		Spec: &models.V1alpha1GcpAccountEntitySpec{
			JSONCredentials:        d.Get("gcp_json_credentials").(string),
		},
	}
	return account
}

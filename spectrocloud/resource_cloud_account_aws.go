package spectrocloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"
)

func resourceCloudAccountAws() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCloudAccountAwsCreate,
		ReadContext:   resourceCloudAccountAwsRead,
		UpdateContext: resourceCloudAccountAwsUpdate,
		DeleteContext: resourceCloudAccountAwsDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"aws_access_key": {
				Type:     schema.TypeString,
				Required: true,
			},
			"aws_secret_key": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceCloudAccountAwsCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1alpha1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	account := toAwsAccount(d)

	uid, err := c.CreateCloudAccountAws(account)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(uid)

	resourceCloudAccountAwsRead(ctx, d, m)

	return diags
}

func resourceCloudAccountAwsRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1alpha1Client)

	var diags diag.Diagnostics

	uid := d.Id()

	account, err := c.GetCloudAccountAws(uid)
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
	if err := d.Set("aws_access_key", account.Spec.AccessKey); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

//
func resourceCloudAccountAwsUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1alpha1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	account := toAwsAccount(d)

	err := c.UpdateCloudAccountAws(account)
	if err != nil {
		return diag.FromErr(err)
	}

	resourceCloudAccountAwsRead(ctx, d, m)

	return diags
}

func resourceCloudAccountAwsDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1alpha1Client)

	var diags diag.Diagnostics

	cloudAccountID := d.Id()

	err := c.DeleteCloudAccountAws(cloudAccountID)
	if err != nil {
		return diag.FromErr(err)
	}

	// d.SetId("") is automatically called assuming delete returns no errors

	return diags
}

func toAwsAccount(d *schema.ResourceData) *models.V1alpha1AwsAccount {
	account := &models.V1alpha1AwsAccount{
		Metadata: &models.V1ObjectMeta{
			Name: d.Get("name").(string),
			UID:  d.Id(),
		},
		Spec: &models.V1alpha1AwsCloudAccount{
			AccessKey: d.Get("aws_access_key").(string),
			SecretKey: d.Get("aws_secret_key").(string),
		},
	}
	return account
}

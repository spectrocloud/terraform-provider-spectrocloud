package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/gomi/pkg/ptr"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"
)

const OverlordUID = "overlordUid"

func resourceCloudAccountVsphere() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCloudAccountVsphereCreate,
		ReadContext:   resourceCloudAccountVsphereRead,
		UpdateContext: resourceCloudAccountVsphereUpdate,
		DeleteContext: resourceCloudAccountVsphereDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"private_cloud_gateway_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"vsphere_vcenter": {
				Type:     schema.TypeString,
				Required: true,
			},
			"vsphere_username": {
				Type:     schema.TypeString,
				Required: true,
			},
			"vsphere_password": {
				Type:     schema.TypeString,
				Required: true,
				Sensitive: true,
			},
			"vsphere_ignore_insecure_error": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func resourceCloudAccountVsphereCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	account := toVsphereAccount(d)

	uid, err := c.CreateCloudAccountVsphere(account)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(uid)

	resourceCloudAccountVsphereRead(ctx, d, m)

	return diags
}

func resourceCloudAccountVsphereRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	var diags diag.Diagnostics

	uid := d.Id()

	account, err := c.GetCloudAccountVsphere(uid)
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
	if err := d.Set("vsphere_vcenter", *account.Spec.VcenterServer); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("vsphere_username", *account.Spec.Username); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("vsphere_ignore_insecure_error", account.Spec.Insecure); err != nil {
		return diag.FromErr(err)
	}

	// Don't read the password!!
	//d.Set("vsphere_password", *account.Spec.Password)

	return diags
}

//
func resourceCloudAccountVsphereUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	account := toVsphereAccount(d)

	err := c.UpdateCloudAccountVsphere(account)
	if err != nil {
		return diag.FromErr(err)
	}

	resourceCloudAccountVsphereRead(ctx, d, m)

	return diags
}

func resourceCloudAccountVsphereDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	var diags diag.Diagnostics

	cloudAccountID := d.Id()

	err := c.DeleteCloudAccountVsphere(cloudAccountID)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func toVsphereAccount(d *schema.ResourceData) *models.V1VsphereAccount {
	account := &models.V1VsphereAccount{
		Metadata: &models.V1ObjectMeta{
			Name: d.Get("name").(string),
			UID:  d.Id(),
		},
		Spec: &models.V1VsphereCloudAccount{
			VcenterServer: ptr.StringPtr(d.Get("vsphere_password").(string)),
			Username:      ptr.StringPtr(d.Get("vsphere_password").(string)),
			Password:      ptr.StringPtr(d.Get("vsphere_password").(string)),
			Insecure:      d.Get("vsphere_ignore_insecure_error").(bool),
		},
	}
	return account
}

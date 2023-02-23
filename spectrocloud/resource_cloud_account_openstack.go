package spectrocloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/hapi/client"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

func resourceCloudAccountOpenstack() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCloudAccountOpenStackCreate,
		ReadContext:   resourceCloudAccountOpenStackRead,
		UpdateContext: resourceCloudAccountOpenStackUpdate,
		DeleteContext: resourceCloudAccountOpenStackDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"private_cloud_gateway_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"openstack_username": {
				Type:     schema.TypeString,
				Required: true,
			},
			"openstack_password": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"identity_endpoint": {
				Type:     schema.TypeString,
				Required: true,
			},
			"openstack_allow_insecure": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"ca_certificate": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"parent_region": {
				Type:     schema.TypeString,
				Required: true,
			},
			"default_domain": {
				Type:     schema.TypeString,
				Required: true,
			},
			"default_project": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceCloudAccountOpenStackCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	account := toOpenStackAccount(d)

	uid, err := c.CreateCloudAccountOpenStack(account)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(uid)

	resourceCloudAccountOpenStackRead(ctx, d, m)

	return diags
}

func resourceCloudAccountOpenStackRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	var diags diag.Diagnostics

	uid := d.Id()

	account, err := c.GetCloudAccountOpenStack(uid)
	if err != nil {
		return diag.FromErr(err)
	} else if account == nil {
		d.SetId("")
		return diags
	}

	if err := d.Set("name", account.Metadata.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("private_cloud_gateway_id", account.Metadata.Annotations[OverlordUID]); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("identity_endpoint", *account.Spec.IdentityEndpoint); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("openstack_username", *account.Spec.Username); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("openstack_allow_insecure", account.Spec.Insecure); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("ca_certificate", account.Spec.CaCert); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("parent_region", account.Spec.ParentRegion); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("default_domain", account.Spec.DefaultDomain); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("default_project", account.Spec.DefaultProject); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceCloudAccountOpenStackUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	account := toOpenStackAccount(d)

	err := c.UpdateCloudAccountOpenStack(account)
	if err != nil {
		return diag.FromErr(err)
	}

	resourceCloudAccountOpenStackRead(ctx, d, m)

	return diags
}

func resourceCloudAccountOpenStackDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	var diags diag.Diagnostics

	cloudAccountID := d.Id()

	err := c.DeleteCloudAccountOpenStack(cloudAccountID)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func toOpenStackAccount(d *schema.ResourceData) *models.V1OpenStackAccount {

	account := &models.V1OpenStackAccount{
		Metadata: &models.V1ObjectMeta{
			Name: d.Get("name").(string),
			UID:  d.Id(),
		},

		Spec: &models.V1OpenStackCloudAccount{
			CaCert:           d.Get("ca_certificate").(string),
			DefaultDomain:    d.Get("default_domain").(string),
			DefaultProject:   d.Get("default_project").(string),
			IdentityEndpoint: types.Ptr(d.Get("identity_endpoint").(string)),
			Insecure:         d.Get("openstack_allow_insecure").(bool),
			ParentRegion:     d.Get("parent_region").(string),
			Password:         types.Ptr(d.Get("openstack_password").(string)),
			Username:         types.Ptr(d.Get("openstack_username").(string)),
		},
	}

	return account
}

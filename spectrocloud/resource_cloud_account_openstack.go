package spectrocloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/api/models"

	"github.com/spectrocloud/terraform-provider-spectrocloud/util/ptr"
)

func resourceCloudAccountOpenstack() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCloudAccountOpenStackCreate,
		ReadContext:   resourceCloudAccountOpenStackRead,
		UpdateContext: resourceCloudAccountOpenStackUpdate,
		DeleteContext: resourceCloudAccountOpenStackDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the OpenStack cloud account.",
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"", "project", "tenant"}, false),
				Description: "The context of the OpenStack configuration. " +
					"Allowed values are `project` or `tenant`. Default value is `project`. " + PROJECT_NAME_NUANCE,
			},
			"private_cloud_gateway_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the private cloud gateway that is used to connect to the OpenStack cloud.",
			},
			"openstack_username": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The username of the OpenStack cloud that is used to connect to the OpenStack cloud.",
			},
			"openstack_password": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "The password of the OpenStack cloud that is used to connect to the OpenStack cloud.",
			},
			"identity_endpoint": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The identity endpoint of the OpenStack cloud that is used to connect to the OpenStack cloud.",
			},
			"openstack_allow_insecure": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether to allow insecure connections to the OpenStack cloud. Default is `false`.",
			},
			"ca_certificate": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The CA certificate of the OpenStack cloud that is used to connect to the OpenStack cloud.",
			},
			"parent_region": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The parent region of the OpenStack cloud that is used to connect to the OpenStack cloud.",
			},
			"default_domain": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The default domain of the OpenStack cloud that is used to connect to the OpenStack cloud.",
			},
			"default_project": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The default project of the OpenStack cloud that is used to connect to the OpenStack cloud.",
			},
		},
	}
}

func resourceCloudAccountOpenStackCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

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
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

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
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

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
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

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
			Name:        d.Get("name").(string),
			Annotations: map[string]string{OverlordUID: d.Get("private_cloud_gateway_id").(string)},
			UID:         d.Id(),
		},

		Spec: &models.V1OpenStackCloudAccount{
			CaCert:           d.Get("ca_certificate").(string),
			DefaultDomain:    d.Get("default_domain").(string),
			DefaultProject:   d.Get("default_project").(string),
			IdentityEndpoint: ptr.To(d.Get("identity_endpoint").(string)),
			Insecure:         d.Get("openstack_allow_insecure").(bool),
			ParentRegion:     d.Get("parent_region").(string),
			Password:         ptr.To(d.Get("openstack_password").(string)),
			Username:         ptr.To(d.Get("openstack_username").(string)),
		},
	}

	return account
}

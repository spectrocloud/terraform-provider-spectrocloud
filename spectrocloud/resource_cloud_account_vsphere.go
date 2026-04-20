package spectrocloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

const OverlordUID = "overlordUid"

func resourceCloudAccountVsphere() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCloudAccountVsphereCreate,
		ReadContext:   resourceCloudAccountVsphereRead,
		UpdateContext: resourceCloudAccountVsphereUpdate,
		DeleteContext: resourceCloudAccountVsphereDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceAccountVsphereImport,
		},
		Description: "A resource to manage a vSphere cloud account in Palette.",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the cloud account. This name is used to identify the cloud account in the Spectro Cloud UI.",
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"", "project", "tenant"}, false),
				Description: "Context of the cloud account. Allowed values are `project` or `tenant`. " +
					"Default value is `project`. " + PROJECT_NAME_NUANCE,
			},
			"private_cloud_gateway_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the private cloud gateway. This is the ID of the private cloud gateway that is used to connect to the vSphere cloud.",
			},
			"vsphere_vcenter": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "vCenter server address. This is the address of the vCenter server that is used to connect to the vSphere cloud.",
			},
			"vsphere_username": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Username of the vSphere cloud. This is the username of the vSphere cloud that is used to connect to the vSphere cloud.",
			},
			"vsphere_password": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "Password of the vSphere cloud. This is the password of the vSphere cloud that is used to connect to the vSphere cloud.",
			},
			"vsphere_ignore_insecure_error": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Ignore insecure error. This is a boolean value that indicates whether to ignore the insecure error or not. If not specified, the default value is false.",
			},
		},
	}
}

func resourceCloudAccountVsphereCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

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
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	var diags diag.Diagnostics

	uid := d.Id()
	account, err := c.GetCloudAccountVsphere(uid)
	if err != nil {
		return handleReadError(d, err, diags)
	} else if account == nil {
		// Deleted - Terraform will recreate it
		d.SetId("")
		return diags
	}

	diagnostics, done := flattenVsphereCloudAccount(d, account)
	if done {
		return diagnostics
	}

	return diags
}

func flattenVsphereCloudAccount(d *schema.ResourceData, account *models.V1VsphereAccount) (diag.Diagnostics, bool) {
	if err := d.Set("name", account.Metadata.Name); err != nil {
		return diag.FromErr(err), true
	}
	if err := d.Set("context", account.Metadata.Annotations["scope"]); err != nil {
		return diag.FromErr(err), true
	}
	if err := d.Set("private_cloud_gateway_id", account.Metadata.Annotations[OverlordUID]); err != nil {
		return diag.FromErr(err), true
	}
	if err := d.Set("vsphere_vcenter", *account.Spec.VcenterServer); err != nil {
		return diag.FromErr(err), true
	}
	if err := d.Set("vsphere_username", *account.Spec.Username); err != nil {
		return diag.FromErr(err), true
	}
	if err := d.Set("vsphere_ignore_insecure_error", account.Spec.Insecure); err != nil {
		return diag.FromErr(err), true
	}
	return nil, false
}

func resourceCloudAccountVsphereUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

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
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

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
			Annotations: map[string]string{
				"scope":     d.Get("context").(string),
				OverlordUID: d.Get("private_cloud_gateway_id").(string),
			},
			UID: d.Id(),
		},
		Spec: &models.V1VsphereCloudAccount{
			VcenterServer: types.Ptr(d.Get("vsphere_vcenter").(string)),
			Username:      types.Ptr(d.Get("vsphere_username").(string)),
			Password:      types.Ptr(d.Get("vsphere_password").(string)),
			Insecure:      d.Get("vsphere_ignore_insecure_error").(bool),
		},
	}
	return account
}

func resourceAccountVsphereImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	err := GetCommonAccount(d, c, "vsphere")
	if err != nil {
		return nil, err
	}

	diags := resourceCloudAccountVsphereRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read cluster for import: %v", diags)
	}

	// Return the resource data. In most cases, this method is only used to
	// import one resource at a time, so you should return the resource data
	// in a slice with a single element.
	return []*schema.ResourceData{d}, nil
}

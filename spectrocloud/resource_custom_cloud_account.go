package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/palette-sdk-go/client"
)

func resourceCustomCloudAccount() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCustomCloudAccountCreate,
		ReadContext:   resourceCustomCloudAccountRead,
		UpdateContext: resourceCustomCloudAccountUpdate,
		DeleteContext: resourceCustomCloudAccountDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the cloud account.",
			},
			"cloud": {
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
				Description: "The cloud provider name.",
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"", "project", "tenant"}, false),
				Description: "The context of the AWS configuration. Allowed values are `project` or `tenant`. " +
					"Default value is `project`. " + PROJECT_NAME_NUANCE,
			},
			"private_cloud_gateway_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the private cloud gateway, which serves as the connection point to establish connectivity with the cloud infrastructure.",
			},
			"credentials": {
				Type:        schema.TypeMap,
				Required:    true,
				Sensitive:   true,
				Description: "The credentials required for accessing the cloud.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceCustomCloudAccountCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	var diags diag.Diagnostics

	accountContext := d.Get("context").(string)
	cloudType := d.Get("cloud").(string)

	// For custom cloud we need to validate cloud type id isCustom for all actions.
	err := c.ValidateCustomCloudType(d.Get("cloud").(string), accountContext)
	if err != nil {
		return diag.FromErr(err)
	}
	account, err := toCustomCloudAccount(d)
	if err != nil {
		return diag.FromErr(err)
	}
	uid, err := c.CreateCustomCloudAccount(account, cloudType, accountContext)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(uid)
	resourceCustomCloudAccountRead(ctx, d, m)

	return diags
}

func resourceCustomCloudAccountRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	var diags diag.Diagnostics
	accountContext := d.Get("context").(string)
	cloudType := d.Get("cloud").(string)

	account, err := c.GetCustomCloudAccount(d.Id(), cloudType, accountContext)
	if err != nil {
		return diag.FromErr(err)
	} else if account == nil {
		// Deleted - Terraform will recreate it
		d.SetId("")
		return diags
	}
	diagnostics, done := flattenCustomCloudAccount(d, account)
	if done {
		return diagnostics
	}

	return diags
}

func resourceCustomCloudAccountUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	accountContext := d.Get("context").(string)
	cloudType := d.Get("cloud").(string)
	account, err := toCustomCloudAccount(d)
	if err != nil {
		return diag.FromErr(err)
	}
	err = c.UpdateCustomCloudAccount(d.Id(), account, cloudType, accountContext)
	if err != nil {
		return diag.FromErr(err)
	}
	resourceCustomCloudAccountRead(ctx, d, m)

	return diags
}

func resourceCustomCloudAccountDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	var diags diag.Diagnostics
	customAccountID := d.Id()
	accountContext := d.Get("context").(string)
	cloudType := d.Get("cloud").(string)
	err := c.DeleteCustomCloudAccount(customAccountID, cloudType, accountContext)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func toCustomCloudAccount(d *schema.ResourceData) (*models.V1CustomAccountEntity, error) {
	var overlayID string
	credentials := make(map[string]string)
	overlayID = d.Get("private_cloud_gateway_id").(string)
	credInterface := d.Get("credentials").(map[string]interface{})
	if credInterface != nil {
		for k, v := range credInterface {
			credentials[k] = v.(string)
		}
	}
	account := &models.V1CustomAccountEntity{
		Metadata: &models.V1ObjectMetaInputEntity{
			Annotations: map[string]string{
				OverlordUID: overlayID,
			},
			Name: d.Get("name").(string),
		},
		Spec: &models.V1CustomCloudAccount{
			Credentials: credentials,
		},
	}
	return account, nil
}

func flattenCustomCloudAccount(d *schema.ResourceData, account *models.V1CustomAccount) (diag.Diagnostics, bool) {
	if err := d.Set("name", account.Metadata.Name); err != nil {
		return diag.FromErr(err), true
	}
	if err := d.Set("context", account.Metadata.Annotations["scope"]); err != nil {
		return diag.FromErr(err), true
	}
	if err := d.Set("private_cloud_gateway_id", account.Metadata.Annotations[OverlordUID]); err != nil {
		return diag.FromErr(err), true
	}
	if err := d.Set("cloud", account.Kind); err != nil {
		return diag.FromErr(err), true
	}
	// We are not setting credentials because they are masked and considered sensitive.

	return nil, false
}

package spectrocloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

func resourceCloudAccountMaas() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCloudAccountMaasCreate,
		ReadContext:   resourceCloudAccountMaasRead,
		UpdateContext: resourceCloudAccountMaasUpdate,
		DeleteContext: resourceCloudAccountMaasDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the MAAS cloud account.",
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"", "project", "tenant"}, false),
				Description: "The context of the MAAS configuration. " +
					"Allowed values are `project` or `tenant`. Default value is `project`. " + PROJECT_NAME_NUANCE,
			},
			"private_cloud_gateway_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the private cloud gateway that is used to connect to the MAAS cloud.",
			},
			"maas_api_endpoint": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Endpoint of the MAAS API that is used to connect to the MAAS cloud. I.e. http://maas:5240/MAAS",
			},
			"maas_api_key": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "API key that is used to connect to the MAAS cloud.",
			},
		},
	}
}

func resourceCloudAccountMaasCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	account := toMaasAccount(d)
	uid, err := c.CreateCloudAccountMaas(account)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(uid)

	resourceCloudAccountMaasRead(ctx, d, m)

	return diags
}

func resourceCloudAccountMaasRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	var diags diag.Diagnostics

	uid := d.Id()
	account, err := c.GetCloudAccountMaas(uid)
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
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

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
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	var diags diag.Diagnostics

	cloudAccountID := d.Id()
	err := c.DeleteCloudAccountMaas(cloudAccountID)
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
			Name:        d.Get("name").(string),
			Annotations: map[string]string{OverlordUID: d.Get("private_cloud_gateway_id").(string)},
			UID:         d.Id(),
		},
		Spec: &models.V1MaasCloudAccount{
			APIEndpoint: &EndpointVal,
			APIKey:      &KeyVal,
		},
	}

	return account
}

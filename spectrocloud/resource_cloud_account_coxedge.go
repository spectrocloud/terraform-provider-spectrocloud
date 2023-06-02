package spectrocloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/palette-sdk-go/client"

	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

func resourceCloudAccountCoxEdge() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCloudAccountCoxEdgeCreate,
		ReadContext:   resourceCloudAccountCoxEdgeRead,
		UpdateContext: resourceCloudAccountCoxEdgeUpdate,
		DeleteContext: resourceCloudAccountCoxEdgeDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the CoxEdge cloud account.",
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				Description:  "The context of the CoxEdge configuration.",
				ValidateFunc: validation.StringInSlice([]string{"", "project", "tenant"}, false),
			},
			"api_key": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The API key for CoxEdge authentication.",
			},
			"service": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The service for CoxEdge.",
			},
			"environment": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The environment for CoxEdge.",
			},
			"api_base_url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The CoxEdge API endpoint.",
			},
			"organization_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The organization ID for CoxEdge.",
			},
		},
	}
}

func resourceCloudAccountCoxEdgeCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	account := toCoxEdgeAccount(d)

	AccountContext := d.Get("context").(string)
	uid, err := c.CreateCloudAccountCoxEdge(account, AccountContext)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(uid)

	resourceCloudAccountCoxEdgeRead(ctx, d, m)

	return diags
}

func resourceCloudAccountCoxEdgeRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	var diags diag.Diagnostics

	uid := d.Id()

	account, err := c.GetCloudAccountCoxEdge(uid)
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
	if err := d.Set("api_base_url", account.Spec.APIBaseURL); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("environment", account.Spec.Environment); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("organization_id", account.Spec.OrganizationID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("service", account.Spec.Service); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceCloudAccountCoxEdgeUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	account := toCoxEdgeAccount(d)

	err := c.UpdateCloudAccountCoxEdge(account)
	if err != nil {
		return diag.FromErr(err)
	}

	resourceCloudAccountCoxEdgeRead(ctx, d, m)

	return diags
}

func resourceCloudAccountCoxEdgeDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	var diags diag.Diagnostics

	cloudAccountID := d.Id()

	err := c.DeleteCloudAccountCoxEdge(cloudAccountID)
	if err != nil {
		return diag.FromErr(err)
	}

	// d.SetId("") is automatically called assuming delete returns no errors

	return diags
}

func toCoxEdgeAccount(d *schema.ResourceData) *models.V1CoxEdgeAccount {
	account := &models.V1CoxEdgeAccount{
		Metadata: &models.V1ObjectMeta{
			Name: d.Get("name").(string),
			UID:  d.Id(),
		},
		Spec: &models.V1CoxEdgeCloudAccount{
			APIBaseURL:     types.Ptr(d.Get("api_base_url").(string)),
			APIKey:         types.Ptr(d.Get("api_key").(string)),
			Environment:    d.Get("environment").(string),
			OrganizationID: d.Get("organization_id").(string),
			Service:        d.Get("service").(string),
		},
	}

	return account
}

package spectrocloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

func resourceCloudAccountGcp() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCloudAccountGcpCreate,
		ReadContext:   resourceCloudAccountGcpRead,
		UpdateContext: resourceCloudAccountGcpUpdate,
		DeleteContext: resourceCloudAccountGcpDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceAccountGcpImport,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the GCP account.",
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"", "project", "tenant"}, false),
				Description: "The context of the GCP configuration. " +
					"Allowed values are `project` or `tenant`. Default value is `project`. " + PROJECT_NAME_NUANCE,
			},
			"gcp_json_credentials": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "The GCP credentials in JSON format. These credentials are required to authenticate and manage.",
			},
		},
	}
}

func resourceCloudAccountGcpCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

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
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

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
	if err := d.Set("context", account.Metadata.Annotations["scope"]); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceCloudAccountGcpUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	account := toGcpAccount(d)

	err := c.UpdateCloudAccountGcp(account)
	if err != nil {
		return diag.FromErr(err)
	}

	resourceCloudAccountGcpRead(ctx, d, m)

	return diags
}

func resourceCloudAccountGcpDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	var diags diag.Diagnostics

	cloudAccountID := d.Id()
	//AccountContext := d.Get("context").(string)
	err := c.DeleteCloudAccountGcp(cloudAccountID)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func toGcpAccount(d *schema.ResourceData) *models.V1GcpAccountEntity {
	account := &models.V1GcpAccountEntity{
		Metadata: &models.V1ObjectMeta{
			Name: d.Get("name").(string),
			UID:  d.Id(),
		},
		Spec: &models.V1GcpAccountEntitySpec{
			JSONCredentials: d.Get("gcp_json_credentials").(string),
		},
	}
	return account
}

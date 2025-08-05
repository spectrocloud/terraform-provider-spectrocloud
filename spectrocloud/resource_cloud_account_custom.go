package spectrocloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

func resourceCloudAccountCustom() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCloudAccountCustomCreate,
		ReadContext:   resourceCloudAccountCustomRead,
		UpdateContext: resourceCloudAccountCustomUpdate,
		DeleteContext: resourceCloudAccountCustomDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceAccountCustomImport,
		},
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
				Description: "The context of the custom cloud configuration. Allowed values are `project` or `tenant`. " +
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
				Optional:    true,
				Sensitive:   true,
				Description: "The credentials required for accessing the cloud.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceCloudAccountCustomCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)
	var diags diag.Diagnostics

	//accountContext := d.Get("context").(string)
	cloudType := d.Get("cloud").(string)

	// For custom cloud we need to validate cloud type id isCustom for all actions.
	err := c.ValidateCustomCloudType(d.Get("cloud").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	account, err := toCloudAccountCustom(d)
	if err != nil {
		return diag.FromErr(err)
	}
	uid, err := c.CreateAccountCustomCloud(account, cloudType)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(uid)
	resourceCloudAccountCustomRead(ctx, d, m)

	return diags
}

func resourceCloudAccountCustomRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	var diags diag.Diagnostics
	cloudType := d.Get("cloud").(string)

	account, err := c.GetCustomCloudAccount(d.Id(), cloudType)
	if err != nil {
		return handleReadError(d, err, diags)
	} else if account == nil {
		// Deleted - Terraform will recreate it
		d.SetId("")
		return diags
	}
	diagnostics, done := flattenCloudAccountCustom(d, account)
	if done {
		return diagnostics
	}

	return diags
}

func resourceCloudAccountCustomUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	cloudType := d.Get("cloud").(string)
	account, err := toCloudAccountCustom(d)
	if err != nil {
		return diag.FromErr(err)
	}
	err = c.UpdateAccountCustomCloud(d.Id(), account, cloudType)
	if err != nil {
		return diag.FromErr(err)
	}
	resourceCloudAccountCustomRead(ctx, d, m)

	return diags
}

func resourceCloudAccountCustomDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	var diags diag.Diagnostics
	customAccountID := d.Id()
	cloudType := d.Get("cloud").(string)
	err := c.DeleteCloudAccountCustomCloud(customAccountID, cloudType)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func toCloudAccountCustom(d *schema.ResourceData) (*models.V1CustomAccountEntity, error) {
	var overlayID string
	credentials := make(map[string]string)
	overlayID = d.Get("private_cloud_gateway_id").(string)

	// Validate that credentials are provided for create/update operations
	credInterface, ok := d.GetOk("credentials")
	if !ok || credInterface == nil {
		return nil, fmt.Errorf("credentials are required for custom cloud account operations")
	}

	credMap, ok := credInterface.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("credentials must be a map of string values")
	}

	if len(credMap) == 0 {
		return nil, fmt.Errorf("credentials cannot be empty - at least one credential key-value pair is required")
	}

	for k, v := range credMap {
		if vStr, ok := v.(string); ok {
			credentials[k] = vStr
		} else {
			return nil, fmt.Errorf("credential value for key '%s' must be a string", k)
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

func flattenCloudAccountCustom(d *schema.ResourceData, account *models.V1CustomAccount) (diag.Diagnostics, bool) {
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

	return nil, false
}

func resourceAccountCustomImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	cloudAccountID, scope, customCloudName, err := ParseResourceCustomCloudImportID(d)
	if err != nil {
		return nil, err
	}
	d.SetId(cloudAccountID + ":" + scope)
	_ = d.Set("context", scope)
	_ = d.Set("cloud", customCloudName)
	c := getV1ClientWithResourceContext(m, scope)

	err = GetCommonAccount(d, c)
	if err != nil {
		return nil, err
	}

	diags := resourceCloudAccountCustomRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read cluster for import: %v", diags)
	}

	return []*schema.ResourceData{d}, nil
}

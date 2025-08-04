package spectrocloud

import (
	"context"
	"fmt"

	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

func resourceCloudAccountAzure() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCloudAccountAzureCreate,
		ReadContext:   resourceCloudAccountAzureRead,
		UpdateContext: resourceCloudAccountAzureUpdate,
		DeleteContext: resourceCloudAccountAzureDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceAccountAzureImport,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the Azure cloud account.",
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"", "project", "tenant"}, false),
				Description: "The context of the Azure configuration. " +
					"Defaults to `project`. " + PROJECT_NAME_NUANCE,
			},
			"private_cloud_gateway_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID of the private cloud gateway. This is the ID of the private cloud gateway that is used to connect to the private cluster endpoint.",
			},
			"azure_tenant_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Unique tenant Id from Azure console.",
			},
			"azure_client_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Unique client Id from Azure console.",
			},
			"azure_client_secret": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "Azure secret for authentication.",
			},
			"tenant_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name of the tenant. This is the name of the tenant that is used to connect to the Azure cloud.",
			},
			"disable_properties_request": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Disable properties request. This is a boolean value that indicates whether to disable properties request or not. If not specified, the default value is `false`.",
			},
			"cloud": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "AzurePublicCloud",
				ValidateFunc: validation.StringInSlice([]string{"AzurePublicCloud", "AzureUSGovernmentCloud", "AzureUSSecretCloud"}, false),
				Description: `The Azure partition in which the cloud account is located. 
Can be 'AzurePublicCloud' for standard Azure regions or 'AzureUSGovernmentCloud' for Azure GovCloud (US) regions or 'AzureUSSecretCloud' for Azure Secret Cloud regions.
Default is 'AzurePublicCloud'.`,
			},
			"tls_cert": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "TLS certificate for authentication. This field is only allowed when cloud is set to 'AzureUSSecretCloud'.",
			},
		},
	}
}

func resourceCloudAccountAzureCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// Validate tls_cert is only used with AzureUSSecretCloud
	if err := validateTlsCertConfiguration(d); err != nil {
		return diag.FromErr(err)
	}

	account := toAzureAccount(d)

	uid, err := c.CreateCloudAccountAzure(account)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(uid)

	resourceCloudAccountAzureRead(ctx, d, m)

	return diags
}

func resourceCloudAccountAzureRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	var diags diag.Diagnostics

	uid := d.Id()

	account, err := c.GetCloudAccountAzure(uid)
	if err != nil {
		return handleReadError(d, err, diags)
	} else if account == nil {
		// Deleted - Terraform will recreate it
		d.SetId("")
		return diags
	}

	diagnostics, done := flattenCloudAccountAzure(d, account)
	if done {
		return diagnostics
	}

	return diags
}

func flattenCloudAccountAzure(d *schema.ResourceData, account *models.V1AzureAccount) (diag.Diagnostics, bool) {
	if err := d.Set("name", account.Metadata.Name); err != nil {
		return diag.FromErr(err), true
	}
	if err := d.Set("context", account.Metadata.Annotations["scope"]); err != nil {
		return diag.FromErr(err), true
	}
	if err := d.Set("private_cloud_gateway_id", account.Metadata.Annotations[OverlordUID]); err != nil {
		return diag.FromErr(err), true
	}
	if err := d.Set("azure_tenant_id", *account.Spec.TenantID); err != nil {
		return diag.FromErr(err), true
	}
	if err := d.Set("azure_client_id", *account.Spec.ClientID); err != nil {
		return diag.FromErr(err), true
	}
	if err := d.Set("tenant_name", account.Spec.TenantName); err != nil {
		return diag.FromErr(err), true
	}
	if err := d.Set("disable_properties_request", account.Spec.Settings.DisablePropertiesRequest); err != nil {
		return diag.FromErr(err), true
	}
	if account.Spec.AzureEnvironment != nil {
		if err := d.Set("cloud", account.Spec.AzureEnvironment); err != nil {
			return diag.FromErr(err), true
		}
	}
	if account.Spec.TLS != nil && account.Spec.TLS.Cert != "" {
		if err := d.Set("tls_cert", account.Spec.TLS.Cert); err != nil {
			return diag.FromErr(err), true
		}
	}
	return nil, false
}

func resourceCloudAccountAzureUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// Validate tls_cert is only used with AzureUSSecretCloud
	if err := validateTlsCertConfiguration(d); err != nil {
		return diag.FromErr(err)
	}

	account := toAzureAccount(d)

	err := c.UpdateCloudAccountAzure(account)
	if err != nil {
		return diag.FromErr(err)
	}

	resourceCloudAccountAzureRead(ctx, d, m)

	return diags
}

func resourceCloudAccountAzureDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	var diags diag.Diagnostics

	cloudAccountID := d.Id()
	//AccountContext := d.Get("context").(string)
	err := c.DeleteCloudAccountAzure(cloudAccountID)
	if err != nil {
		return diag.FromErr(err)
	}

	// d.SetId("") is automatically called assuming delete returns no errors

	return diags
}

func toAzureAccount(d *schema.ResourceData) *models.V1AzureAccount {
	clientSecret := strfmt.Password(d.Get("azure_client_secret").(string)).String()
	account := &models.V1AzureAccount{
		Metadata: &models.V1ObjectMeta{
			Name: d.Get("name").(string),
			Annotations: map[string]string{
				"scope":     d.Get("context").(string),
				OverlordUID: d.Get("private_cloud_gateway_id").(string),
			},
			UID: d.Id(),
		},
		Spec: &models.V1AzureCloudAccount{
			ClientID:     types.Ptr(d.Get("azure_client_id").(string)),
			ClientSecret: &clientSecret,
			TenantID:     types.Ptr(d.Get("azure_tenant_id").(string)),
			TenantName:   d.Get("tenant_name").(string),
			Settings: &models.V1CloudAccountSettings{
				DisablePropertiesRequest: d.Get("disable_properties_request").(bool),
			},
		},
	}

	// add partition to account
	if d.Get("cloud") != nil {
		account.Spec.AzureEnvironment = types.Ptr(d.Get("cloud").(string))
	}

	// add TLS configuration if tls_cert is provided
	if tlsCert, ok := d.GetOk("tls_cert"); ok && tlsCert.(string) != "" {
		account.Spec.TLS = &models.V1AzureSecretTLSConfig{
			Cert: tlsCert.(string),
		}
	}

	return account
}

func validateTlsCertConfiguration(d *schema.ResourceData) error {
	cloud := d.Get("cloud").(string)
	tlsCert := d.Get("tls_cert").(string)

	// If tls_cert is provided but cloud is not AzureUSSecretCloud, return an error
	if tlsCert != "" && cloud != "AzureUSSecretCloud" {
		return fmt.Errorf("tls_cert can only be set when cloud is 'AzureUSSecretCloud', but cloud is set to '%s'", cloud)
	}

	return nil
}

func resourceAccountAzureImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	err := GetCommonAccount(d, c)
	if err != nil {
		return nil, err
	}

	diags := resourceCloudAccountAzureRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read cluster for import: %v", diags)
	}

	// Return the resource data. In most cases, this method is only used to
	// import one resource at a time, so you should return the resource data
	// in a slice with a single element.
	return []*schema.ResourceData{d}, nil
}

package spectrocloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

func resourceCloudAccountApacheCloudStack() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCloudAccountApacheCloudStackCreate,
		ReadContext:   resourceCloudAccountApacheCloudStackRead,
		UpdateContext: resourceCloudAccountApacheCloudStackUpdate,
		DeleteContext: resourceCloudAccountApacheCloudStackDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceAccountApacheCloudStackImport,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the Apache CloudStack cloud account.",
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"", "project", "tenant"}, false),
				Description: "The context of the Apache CloudStack configuration. " +
					"Allowed values are `project` or `tenant`. Default value is `project`. " + PROJECT_NAME_NUANCE,
			},
			"private_cloud_gateway_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the private cloud gateway that is used to connect to the Apache CloudStack cloud.",
			},
			"api_url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The API URL of the Apache CloudStack management server. For example: https://cloudstack.example.com:8080/client/api",
			},
			"api_key": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "The API key for Apache CloudStack authentication.",
			},
			"secret_key": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "The secret key for Apache CloudStack authentication.",
			},
			"domain": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The domain for the Apache CloudStack account. Optional, for multi-domain CloudStack environments. Default is empty (ROOT domain).",
			},
			"insecure": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Skip SSL certificate verification. Default is `false`. Note: Apache CloudStack must have valid SSL certificates from a trusted CA if this is false.",
			},
		},
	}
}

func resourceCloudAccountApacheCloudStackCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	var diags diag.Diagnostics

	account := toApacheCloudStackAccount(d, c)
	uid, err := c.CreateCloudAccountCloudStack(account)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(uid)

	resourceCloudAccountApacheCloudStackRead(ctx, d, m)

	return diags
}

func resourceCloudAccountApacheCloudStackRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	var diags diag.Diagnostics

	uid := d.Id()
	account, err := c.GetCloudAccountCloudStack(uid)
	if err != nil {
		return handleReadError(d, err, diags)
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
	if account.Spec != nil {
		if account.Spec.APIURL != nil {
			if err := d.Set("api_url", *account.Spec.APIURL); err != nil {
				return diag.FromErr(err)
			}
		}
		if err := d.Set("domain", account.Spec.Domain); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("insecure", account.Spec.Insecure); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func resourceCloudAccountApacheCloudStackUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	var diags diag.Diagnostics

	account := toApacheCloudStackAccount(d, c)

	err := c.UpdateCloudAccountCloudStack(account)
	if err != nil {
		return diag.FromErr(err)
	}

	resourceCloudAccountApacheCloudStackRead(ctx, d, m)

	return diags
}

func resourceCloudAccountApacheCloudStackDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	var diags diag.Diagnostics

	cloudAccountID := d.Id()
	err := c.DeleteCloudAccountCloudStack(cloudAccountID)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func toApacheCloudStackAccount(d *schema.ResourceData, c *client.V1Client) *models.V1CloudStackAccount {

	account := &models.V1CloudStackAccount{
		Metadata: &models.V1ObjectMeta{
			Name:        d.Get("name").(string),
			Annotations: map[string]string{OverlordUID: d.Get("private_cloud_gateway_id").(string)},
			UID:         d.Id(),
		},
		Spec: &models.V1CloudStackCloudAccount{
			APIURL:    types.Ptr(d.Get("api_url").(string)),
			APIKey:    types.Ptr(d.Get("api_key").(string)),
			SecretKey: types.Ptr(d.Get("secret_key").(string)),
			Domain:    d.Get("domain").(string),
			Insecure:  d.Get("insecure").(bool),
		},
	}
	// for system pcg, set overlordType to "system" in annotation only for apache cloudstack account
	pcgID := d.Get("private_cloud_gateway_id").(string)
	pcg, _ := c.GetPCGByID(pcgID)
	if pcg.Metadata.Name == "System Private Gateway" {
		account.Metadata.Annotations["overlordType"] = "system"
	}

	return account
}

func resourceAccountApacheCloudStackImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	err := GetCommonAccount(d, c)
	if err != nil {
		return nil, err
	}

	diags := resourceCloudAccountApacheCloudStackRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read account for import: %v", diags)
	}

	return []*schema.ResourceData{d}, nil
}

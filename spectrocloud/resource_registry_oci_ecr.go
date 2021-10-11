package spectrocloud

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceRegistryOciEcr() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRegistryEcrCreate,
		ReadContext:   resourceRegistryEcrRead,
		UpdateContext: resourceRegistryEcrUpdate,
		DeleteContext: resourceRegistryEcrDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 2,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"is_private": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"endpoint": {
				Type:     schema.TypeString,
				Required: true,
			},
			"credentials": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"credential_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"secret", "sts"}, false),
						},
						"sts": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"arn": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"external_id": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceRegistryEcrCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	var diags diag.Diagnostics

	registry := toRegistryEcr(d)
	uid, err := c.CreateOciEcrRegistry(registry)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(uid)

	return diags
}

func resourceRegistryEcrRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	var diags diag.Diagnostics

	registry, err := c.GetRegistryOci(d.Id())
	if err != nil {
		return diag.FromErr(err)
	} else if registry == nil {
		// Deleted - Terraform will recreate it
		d.SetId("")
		return diags
	}

	if err := d.Set("name", registry.Metadata.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("is_private", registry.Spec.IsPrivate); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("endpoint", registry.Spec.Endpoint); err != nil {
		return diag.FromErr(err)
	}

	if registry.Spec.Credentials.CredentialType == models.V1AwsCloudAccountCredentialTypeSts {
		sts_list := make([]interface{}, 0, 1)
		sts := make(map[string]interface{})
		sts["arn"] = registry.Spec.Credentials.Sts.Arn
		sts["external_id"] = registry.Spec.Credentials.Sts.ExternalID
		sts_list = append(sts_list, sts)

		cred := make(map[string]interface{})
		cred["credential_type"]  = models.V1AwsCloudAccountCredentialTypeSts
		cred["sts"] = sts_list
		credentials := make([]interface{}, 0, 1)
		credentials = append(credentials, cred)
		if err := d.Set("credentials", credentials); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func resourceRegistryEcrUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	var diags diag.Diagnostics

	registry := toRegistryEcr(d)
	err := c.UpdateEcrRegistry(d.Id(), registry)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceRegistryEcrDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	var diags diag.Diagnostics
	err := c.DeleteRegistry(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func toRegistryEcr(d *schema.ResourceData) *models.V1EcrRegistry {
	endpoint := d.Get("endpoint").(string)
	isPrivate := d.Get("is_private").(bool)
	s3config := d.Get("credentials").([]interface{})[0].(map[string]interface{})
	return &models.V1EcrRegistry{
		Metadata: &models.V1ObjectMeta{
			Name: d.Get("name").(string),
		},
		Spec: &models.V1EcrRegistrySpec{
			Credentials: toRegistryAwsAccountCredential(s3config),
			Endpoint : &endpoint,
			IsPrivate: &isPrivate,
		},
	}
}

func toRegistryAwsAccountCredential(regCred map[string]interface{}) *models.V1AwsCloudAccount {
	account := &models.V1AwsCloudAccount{}
	if len(regCred["credential_type"].(string)) == 0 || regCred["credential_type"].(string) == "secret" {
		account.CredentialType = models.V1AwsCloudAccountCredentialTypeSecret
		account.AccessKey = regCred["access_key"].(string)
		account.SecretKey = regCred["secret_key"].(string)
	} else if regCred["credential_type"].(string) == "sts" {
		account.CredentialType = models.V1AwsCloudAccountCredentialTypeSts
		sts := regCred["sts"].([]interface{})[0].(map[string]interface{})
		account.Sts = &models.V1AwsStsCredentials{
			Arn:        sts["arn"].(string),
			ExternalID: sts["external_id"].(string),
		}
	}
	return account
}

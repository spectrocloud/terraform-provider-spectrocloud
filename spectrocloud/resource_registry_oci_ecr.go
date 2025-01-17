package spectrocloud

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
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
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the OCI registry.",
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"ecr", "basic"}, false),
				Description:  "The type of the registry. Possible values are 'ecr' (Amazon Elastic Container Registry) or 'basic' (for other types of OCI registries).",
			},
			"is_private": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Specifies whether the registry is private or public. Private registries require authentication to access.",
			},
			"is_synchronization": {
				Type:        schema.TypeBool,
				Default:     false,
				Optional:    true,
				Description: "Specifies whether the registry is synchronized.",
			},
			"endpoint": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The URL endpoint of the OCI registry. This is where the container images are hosted and accessed.",
			},
			"provider_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "helm",
				ValidateFunc: validation.StringInSlice([]string{"helm", "zarf", "pack"}, false),
				Description:  "The type of provider used for interacting with the registry. Supported value's are `helm`, `zarf` and `pack`, The default is 'helm'. `zarf` is allowed with `type=\"basic\"`  ",
			},
			"credentials": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Description: "Authentication credentials to access the private OCI registry. Required if `is_private` is set to `true`",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"credential_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"secret", "sts", "basic", "noAuth"}, false),
							Description:  "The type of authentication used for accessing the registry. Supported values are 'secret', 'sts', 'basic', and 'noAuth'.",
						},
						"access_key": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The access key for accessing the registry. Required if 'credential_type' is set to 'secret'.",
						},
						"secret_key": {
							Type:        schema.TypeString,
							Optional:    true,
							Sensitive:   true,
							Description: "The secret key for accessing the registry. Required if 'credential_type' is set to 'secret'.",
						},
						"arn": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The Amazon Resource Name (ARN) used for AWS-based authentication. Required if 'credential_type' is 'sts'.",
						},
						"external_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The external ID used for AWS STS (Security Token Service) authentication. Required if 'credential_type' is 'sts'.",
						},
						"username": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The username for basic authentication. Required if 'credential_type' is 'basic'.",
						},
						"password": {
							Type:        schema.TypeString,
							Optional:    true,
							Sensitive:   true,
							Description: "The password for basic authentication. Required if 'credential_type' is 'basic'.",
						},
					},
				},
			},
		},
		CustomizeDiff: func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
			providerType := d.Get("provider_type").(string)
			registryType := d.Get("type").(string)
			// Validate that `provider_type` is "zarf" only if `type` is "basic"
			if providerType == "zarf" && registryType != "basic" {
				return fmt.Errorf("`provider_type` set to `zarf` is only allowed when `type` is `basic`")
			}
			return nil
		},
	}
}

func resourceRegistryEcrCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "tenant")
	var diags diag.Diagnostics

	registryType := d.Get("type").(string)

	if registryType == "ecr" {

		registry := toRegistryEcr(d)

		uid, err := c.CreateOciEcrRegistry(registry)
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(uid)
	} else if registryType == "basic" {
		registry := toRegistryBasic(d)

		uid, err := c.CreateOciBasicRegistry(registry)
		if err != nil {
			return diag.FromErr(err)
		}

		d.SetId(uid)
	}

	return diags
}

func resourceRegistryEcrRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "tenant")
	var diags diag.Diagnostics

	registryType := d.Get("type").(string)

	if registryType == "ecr" {
		registry, err := c.GetOciEcrRegistry(d.Id())
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
		switch registry.Spec.Credentials.CredentialType {
		case models.V1AwsCloudAccountCredentialTypeSts:
			credentials := make([]interface{}, 0, 1)
			acc := make(map[string]interface{})
			acc["arn"] = registry.Spec.Credentials.Sts.Arn
			acc["external_id"] = registry.Spec.Credentials.Sts.ExternalID
			acc["credential_type"] = models.V1AwsCloudAccountCredentialTypeSts
			credentials = append(credentials, acc)
			if err := d.Set("credentials", credentials); err != nil {
				return diag.FromErr(err)
			}
		case models.V1AwsCloudAccountCredentialTypeSecret:
			credentials := make([]interface{}, 0, 1)
			acc := make(map[string]interface{})
			acc["access_key"] = registry.Spec.Credentials.AccessKey
			acc["credential_type"] = models.V1AwsCloudAccountCredentialTypeSecret
			credentials = append(credentials, acc)
			if err := d.Set("credentials", credentials); err != nil {
				return diag.FromErr(err)
			}
		default:
			errMsg := fmt.Sprintf("Registry type %s not implemented.", registry.Spec.Credentials.CredentialType)
			err = errors.New(errMsg)
			return diag.FromErr(err)
		}
		return diags

	} else if registryType == "basic" {
		registry, err := c.GetOciBasicRegistry(d.Id())
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
		if err := d.Set("endpoint", registry.Spec.Endpoint); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("provider_type", registry.Spec.ProviderType); err != nil {
			return diag.FromErr(err)
		}
		return diags
	}

	return diags
}

func resourceRegistryEcrUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "tenant")
	var diags diag.Diagnostics

	registryType := d.Get("type").(string)

	if registryType == "ecr" {
		registry := toRegistryEcr(d)
		err := c.UpdateOciEcrRegistry(d.Id(), registry)
		if err != nil {
			return diag.FromErr(err)
		}
	} else if registryType == "basic" {
		registry := toRegistryBasic(d)
		err := c.UpdateOciBasicRegistry(d.Id(), registry)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func resourceRegistryEcrDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "tenant")
	var diags diag.Diagnostics

	registryType := d.Get("type").(string)
	if registryType == "ecr" {
		err := c.DeleteOciEcrRegistry(d.Id())
		if err != nil {
			return diag.FromErr(err)
		}
	} else if registryType == "basic" {
		err := c.DeleteOciBasicRegistry(d.Id())
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func toRegistryEcr(d *schema.ResourceData) *models.V1EcrRegistry {
	endpoint := d.Get("endpoint").(string)
	isPrivate := d.Get("is_private").(bool)
	isSynchronization := d.Get("is_synchronization").(bool)
	providerType := d.Get("provider_type").(string)
	s3config := d.Get("credentials").([]interface{})[0].(map[string]interface{})
	return &models.V1EcrRegistry{
		Metadata: &models.V1ObjectMeta{
			Name: d.Get("name").(string),
		},
		Spec: &models.V1EcrRegistrySpec{
			Credentials:     toRegistryAwsAccountCredential(s3config),
			Endpoint:        &endpoint,
			IsPrivate:       &isPrivate,
			ProviderType:    &providerType,
			IsSyncSupported: isSynchronization,
		},
	}
}

func toRegistryBasic(d *schema.ResourceData) *models.V1BasicOciRegistry {
	endpoint := d.Get("endpoint").(string)
	provider := d.Get("provider_type").(string)
	isSynchronization := d.Get("is_synchronization").(bool)
	authConfig := d.Get("credentials").([]interface{})[0].(map[string]interface{})

	var username, password string

	username = authConfig["username"].(string)
	password = authConfig["password"].(string)

	return &models.V1BasicOciRegistry{
		Metadata: &models.V1ObjectMeta{
			Name: d.Get("name").(string),
		},
		Spec: &models.V1BasicOciRegistrySpec{
			Endpoint:        &endpoint,
			ProviderType:    &provider,
			BaseContentPath: "",
			Auth: &models.V1RegistryAuth{
				Username: username,
				Password: strfmt.Password(password),
				Type:     "basic",
				TLS: &models.V1TLSConfiguration{
					Enabled:            true,
					InsecureSkipVerify: false,
				},
			},
			IsSyncSupported: isSynchronization,
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
		account.Sts = &models.V1AwsStsCredentials{
			Arn:        regCred["arn"].(string),
			ExternalID: regCred["external_id"].(string),
		}
	}
	return account
}

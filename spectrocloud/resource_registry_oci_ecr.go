package spectrocloud

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spectrocloud/palette-sdk-go/client"

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
		Importer: &schema.ResourceImporter{
			StateContext: resourceRegistryOciImport,
		},

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
			"endpoint_suffix": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Specifies a suffix to append to the endpoint. This field is optional, but some registries (e.g., JFrog) may require it. The final registry URL is constructed by appending this suffix to the endpoint.",
			},
			"base_content_path": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The relative path to the endpoint specified.",
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
						"tls_config": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "TLS configuration for the registry.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"certificate": {
										Type:        schema.TypeString,
										Optional:    true,
										Default:     "",
										Description: "Specifies the TLS certificate used for secure communication. Required for enabling SSL/TLS encryption.",
									},
									"insecure_skip_verify": {
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     false,
										Description: "Disables TLS certificate verification when set to true. ⚠️ WARNING: Setting this to true disables SSL certificate verification and makes connections vulnerable to man-in-the-middle attacks. Only use this when connecting to registries with self-signed certificates in trusted networks.",
									},
								},
							},
						},
					},
				},
			},
		},
		CustomizeDiff: func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
			providerType := d.Get("provider_type").(string)
			registryType := d.Get("type").(string)
			isSync := d.Get("is_synchronization").(bool)
			// Validate that `provider_type` is "zarf" only if `type` is "basic"
			if providerType == "zarf" && registryType != "basic" {
				return fmt.Errorf("`provider_type` set to `zarf` is only allowed when `type` is `basic`")
			}
			// if providerType == "zarf" && isSync {
			// 	return fmt.Errorf("`provider_type` set to `zarf` is only allowed when `is_synchronization` is set to `false`")
			// }
			if providerType == "pack" && !isSync {
				return fmt.Errorf("`provider_type` set to `pack` is only allowed when `is_synchronization` is set to `true`")
			}
			return nil
		},
	}
}

func validateRegistryCred(c *client.V1Client, registryType string, providerType string, isSync bool, basicSpec *models.V1BasicOciRegistrySpec, ecrSpec *models.V1EcrRegistrySpec) error {
	if isSync && (providerType == "pack" || providerType == "helm" || providerType == "zarf") {
		switch registryType {
		case "basic":
			if basicSpec != nil {
				if err := c.ValidateOciBasicRegistry(basicSpec); err != nil {
					return err
				}
			}
		case "ecr":
			if ecrSpec != nil {
				if err := c.ValidateOciEcrRegistry(ecrSpec); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func resourceRegistryEcrCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "tenant")
	var diags diag.Diagnostics

	registryType := d.Get("type").(string)
	providerType := d.Get("provider_type").(string)
	isSync := d.Get("is_synchronization").(bool)

	switch registryType {
	case "ecr":
		registry := toRegistryEcr(d)
		if err := validateRegistryCred(c, registryType, providerType, isSync, nil, registry.Spec); err != nil {
			return diag.FromErr(err)
		}
		uid, err := c.CreateOciEcrRegistry(registry)
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(uid)
	case "basic":
		registry := toRegistryBasic(d)
		if err := validateRegistryCred(c, registryType, providerType, isSync, registry.Spec, nil); err != nil {
			return diag.FromErr(err)
		}
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
			return handleReadError(d, err, diags)
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
		if err := d.Set("base_content_path", registry.Spec.BaseContentPath); err != nil {
			return diag.FromErr(err)
		}
		credentials := make([]interface{}, 0, 1)
		acc := make(map[string]interface{})
		switch *registry.Spec.Credentials.CredentialType {
		case models.V1AwsCloudAccountCredentialTypeSts:
			acc["arn"] = registry.Spec.Credentials.Sts.Arn
			acc["external_id"] = registry.Spec.Credentials.Sts.ExternalID
			acc["credential_type"] = models.V1AwsCloudAccountCredentialTypeSts
		case models.V1AwsCloudAccountCredentialTypeSecret:
			acc["access_key"] = registry.Spec.Credentials.AccessKey
			acc["credential_type"] = models.V1AwsCloudAccountCredentialTypeSecret
		default:
			errMsg := fmt.Sprintf("Registry type %s not implemented.", *registry.Spec.Credentials.CredentialType)
			err = errors.New(errMsg)
			return diag.FromErr(err)
		}
		// tls configuration handling
		tlsConfig := make([]interface{}, 0, 1)
		tls := make(map[string]interface{})
		tls["certificate"] = registry.Spec.TLS.Certificate
		tls["insecure_skip_verify"] = registry.Spec.TLS.InsecureSkipVerify
		tlsConfig = append(tlsConfig, tls)
		acc["tls_config"] = tlsConfig
		credentials = append(credentials, acc)

		if err := d.Set("credentials", credentials); err != nil {
			return diag.FromErr(err)
		}
		return diags

	} else if registryType == "basic" {
		registry, err := c.GetOciBasicRegistry(d.Id())
		if err != nil {
			return handleReadError(d, err, diags)
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
		if err := d.Set("base_content_path", registry.Spec.BaseContentPath); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("endpoint_suffix", registry.Spec.BasePath); err != nil {
			return diag.FromErr(err)
		}
		credentials := make([]interface{}, 0, 1)
		acc := make(map[string]interface{})
		acc["username"] = registry.Spec.Auth.Username
		acc["password"] = registry.Spec.Auth.Password
		// tls configuration handling
		tlsConfig := make([]interface{}, 0, 1)
		tls := make(map[string]interface{})
		tls["certificate"] = registry.Spec.Auth.TLS.Certificate
		tls["insecure_skip_verify"] = registry.Spec.Auth.TLS.InsecureSkipVerify
		tlsConfig = append(tlsConfig, tls)
		acc["tls_config"] = tlsConfig
		credentials = append(credentials, acc)
		if err := d.Set("credentials", credentials); err != nil {
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
	providerType := d.Get("provider_type").(string)
	isSync := d.Get("is_synchronization").(bool)
	if registryType == "ecr" {
		registry := toRegistryEcr(d)
		if err := validateRegistryCred(c, registryType, providerType, isSync, nil, registry.Spec); err != nil {
			return diag.FromErr(err)
		}
		err := c.UpdateOciEcrRegistry(d.Id(), registry)
		if err != nil {
			return diag.FromErr(err)
		}
	} else if registryType == "basic" {
		registry := toRegistryBasic(d)
		if err := validateRegistryCred(c, registryType, providerType, isSync, registry.Spec, nil); err != nil {
			return diag.FromErr(err)
		}
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
	baseContentPath := d.Get("base_content_path").(string)
	s3config := d.Get("credentials").([]interface{})[0].(map[string]interface{})
	var tlsCertificate string
	var tlsSkipVerify bool
	if len(s3config["tls_config"].([]interface{})) > 0 {
		tlsCertificate = s3config["tls_config"].([]interface{})[0].(map[string]interface{})["certificate"].(string)
		tlsSkipVerify = s3config["tls_config"].([]interface{})[0].(map[string]interface{})["insecure_skip_verify"].(bool)
	}

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
			BaseContentPath: baseContentPath,
			TLS: &models.V1TLSConfiguration{
				Certificate:        tlsCertificate,
				Enabled:            true,
				InsecureSkipVerify: tlsSkipVerify,
			},
		},
	}
}

func toRegistryBasic(d *schema.ResourceData) *models.V1BasicOciRegistry {
	endpoint := d.Get("endpoint").(string)
	provider := d.Get("provider_type").(string)
	isSynchronization := d.Get("is_synchronization").(bool)
	endpointSuffix := d.Get("endpoint_suffix").(string)
	baseContentPath := d.Get("base_content_path").(string)
	authConfig := d.Get("credentials").([]interface{})[0].(map[string]interface{})
	var tlsCertificate string
	var tlsSkipVerify bool
	if len(authConfig["tls_config"].([]interface{})) > 0 {
		tlsCertificate = authConfig["tls_config"].([]interface{})[0].(map[string]interface{})["certificate"].(string)
		tlsSkipVerify = authConfig["tls_config"].([]interface{})[0].(map[string]interface{})["insecure_skip_verify"].(bool)
	}
	var username, password string

	username = authConfig["username"].(string)
	password = authConfig["password"].(string)

	return &models.V1BasicOciRegistry{
		Metadata: &models.V1ObjectMeta{
			Name: d.Get("name").(string),
		},
		Spec: &models.V1BasicOciRegistrySpec{
			Endpoint:        &endpoint,
			BasePath:        endpointSuffix,
			ProviderType:    &provider,
			BaseContentPath: baseContentPath,
			Auth: &models.V1RegistryAuth{
				Username: username,
				Password: strfmt.Password(password),
				Type:     "basic",
				TLS: &models.V1TLSConfiguration{
					Certificate:        tlsCertificate,
					Enabled:            true,
					InsecureSkipVerify: tlsSkipVerify,
				},
			},
			IsSyncSupported: isSynchronization,
		},
	}

}

func toRegistryAwsAccountCredential(regCred map[string]interface{}) *models.V1AwsCloudAccount {
	account := &models.V1AwsCloudAccount{}
	if len(regCred["credential_type"].(string)) == 0 || regCred["credential_type"].(string) == "secret" {
		account.CredentialType = models.V1AwsCloudAccountCredentialTypeSecret.Pointer()
		account.AccessKey = regCred["access_key"].(string)
		account.SecretKey = regCred["secret_key"].(string)
	} else if regCred["credential_type"].(string) == "sts" {
		account.CredentialType = models.V1AwsCloudAccountCredentialTypeSts.Pointer()
		account.Sts = &models.V1AwsStsCredentials{
			Arn:        regCred["arn"].(string),
			ExternalID: regCred["external_id"].(string),
		}
	}
	return account
}

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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
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
			"wait_for_sync": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "If `true`, Terraform will wait for the OCI registry to complete its initial synchronization before marking the resource as created or updated. This option is applicable when `provider_type` is set to `zarf` or `helm`. Default value is `false`.",
			},
			"wait_for_status_message": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The status message from the last sync operation. This is a computed field that is populated after sync completes.",
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

		// Wait for sync if requested and provider_type is zarf or helm
		if (providerType == "zarf" || providerType == "helm") && d.Get("wait_for_sync") != nil && d.Get("wait_for_sync").(bool) {
			diagnostics, isError := waitForOciRegistrySync(ctx, d, uid, diags, c, schema.TimeoutCreate)
			if len(diagnostics) > 0 {
				diags = append(diags, diagnostics...)
			}
			// Fetch final sync status and set wait_for_status_message
			syncStatus, statusErr := c.GetOciBasicRegistrySyncStatus(uid)
			if statusErr == nil && syncStatus != nil {
				statusMessage := ""
				if syncStatus.Message != "" {
					statusMessage = syncStatus.Message
				} else if syncStatus.Status != "" {
					statusMessage = fmt.Sprintf("Status: %s", syncStatus.Status)
				}
				if err := d.Set("wait_for_status_message", statusMessage); err != nil {
					diags = append(diags, diag.FromErr(err)...)
				}
			}
			if isError {
				return diagnostics
			}
		}
	}

	return diags
}

func resourceRegistryEcrRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "tenant")
	var diags diag.Diagnostics

	registryType := d.Get("type").(string)

	switch registryType {
	case "ecr":
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
		isSyncSupported := false
		if registry.Status != nil && registry.Status.SyncStatus != nil {
			isSyncSupported = registry.Status.SyncStatus.IsSyncSupported
		} else if registry.Spec.IsSyncSupported {
			// Fallback to Spec if Status is not available (for backward compatibility)
			isSyncSupported = registry.Spec.IsSyncSupported
		}
		if err := d.Set("is_synchronization", isSyncSupported); err != nil {
			return diag.FromErr(err)
		}
		providerType := "helm" // default per schema
		if registry.Spec.ProviderType != nil {
			providerType = *registry.Spec.ProviderType
		}
		if err := d.Set("provider_type", providerType); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("wait_for_sync", false); err != nil {
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

	case "basic":
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
		isPrivate := false
		if registry.Spec.Auth != nil && registry.Spec.Auth.Type == "basic" {
			isPrivate = true
		}
		if err := d.Set("is_private", isPrivate); err != nil {
			return diag.FromErr(err)
		}
		providerType := "helm" // default per schema
		if registry.Spec.ProviderType != nil {
			providerType = *registry.Spec.ProviderType
		}
		if err := d.Set("provider_type", providerType); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("base_content_path", registry.Spec.BaseContentPath); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("endpoint_suffix", registry.Spec.BasePath); err != nil {
			return diag.FromErr(err)
		}
		isSyncSupported := false
		if registry.Status != nil && registry.Status.SyncStatus != nil {
			isSyncSupported = registry.Status.SyncStatus.IsSyncSupported
		} else if registry.Spec.IsSyncSupported {
			// Fallback to Spec if Status is not available (for backward compatibility)
			isSyncSupported = registry.Spec.IsSyncSupported
		}
		if err := d.Set("is_synchronization", isSyncSupported); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("wait_for_sync", false); err != nil {
			return diag.FromErr(err)
		}
		credentials := make([]interface{}, 0, 1)
		acc := make(map[string]interface{})
		// Read the actual auth type from the API response
		switch registry.Spec.Auth.Type {
		case "noAuth":
			acc["credential_type"] = "noAuth"
			acc["username"] = ""
			acc["password"] = ""
		case "basic":
			acc["credential_type"] = "basic"
			acc["username"] = registry.Spec.Auth.Username
			// FIX: Preserve password from state to avoid drift detection when API returns masked/different format
			// This applies to ALL provider types: helm, zarf, and pack
			if currentCredsRaw := d.Get("credentials"); currentCredsRaw != nil {
				if currentCredsList, ok := currentCredsRaw.([]interface{}); ok && len(currentCredsList) > 0 {
					if currentCredMap, ok := currentCredsList[0].(map[string]interface{}); ok {
						if password, exists := currentCredMap["password"]; exists && password != nil {
							// Preserve password from state to avoid drift
							acc["password"] = password
						} else {
							acc["password"] = registry.Spec.Auth.Password.String()
						}
					}
				}
			} else {
				// No existing credentials in state, use API value
				acc["password"] = registry.Spec.Auth.Password.String()
			}
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
	}

	return diags
}

func resourceRegistryEcrUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "tenant")
	var diags diag.Diagnostics
	// VALIDATION: Prevent changing is_synchronization from true to false
	// Once synchronization is enabled, it cannot be disabled
	if d.HasChange("is_synchronization") {
		oldSync, newSync := d.GetChange("is_synchronization")
		oldSyncBool := oldSync.(bool)
		newSyncBool := newSync.(bool)

		// If old value was true and new value is false, reject the change
		if oldSyncBool && !newSyncBool {
			return diag.FromErr(fmt.Errorf(
				"cannot disable synchronization: `is_synchronization` cannot be modified during Day-2 Operations"))
		}
	}

	registryType := d.Get("type").(string)
	providerType := d.Get("provider_type").(string)
	isSync := d.Get("is_synchronization").(bool)
	switch registryType {
	case "ecr":
		registry := toRegistryEcr(d)
		if err := validateRegistryCred(c, registryType, providerType, isSync, nil, registry.Spec); err != nil {
			return diag.FromErr(err)
		}
		err := c.UpdateOciEcrRegistry(d.Id(), registry)
		if err != nil {
			return diag.FromErr(err)
		}
	case "basic":
		registry := toRegistryBasic(d)
		if err := validateRegistryCred(c, registryType, providerType, isSync, registry.Spec, nil); err != nil {
			return diag.FromErr(err)
		}
		err := c.UpdateOciBasicRegistry(d.Id(), registry)
		if err != nil {
			return diag.FromErr(err)
		}

		// Wait for sync if requested and provider_type is zarf or helm
		if (providerType == "zarf" || providerType == "helm") && d.Get("wait_for_sync") != nil && d.Get("wait_for_sync").(bool) {
			diagnostics, isError := waitForOciRegistrySync(ctx, d, d.Id(), diags, c, schema.TimeoutUpdate)
			if len(diagnostics) > 0 {
				diags = append(diags, diagnostics...)
			}
			// Fetch final sync status and set wait_for_status_message
			syncStatus, statusErr := c.GetOciBasicRegistrySyncStatus(d.Id())
			if statusErr == nil && syncStatus != nil {
				statusMessage := ""
				if syncStatus.Message != "" {
					statusMessage = syncStatus.Message
				} else if syncStatus.Status != "" {
					statusMessage = fmt.Sprintf("Status: %s", syncStatus.Status)
				}
				if err := d.Set("wait_for_status_message", statusMessage); err != nil {
					diags = append(diags, diag.FromErr(err)...)
				}
			}
			if isError {
				return diagnostics
			}
		}
	}

	return diags
}

func resourceRegistryEcrDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "tenant")
	var diags diag.Diagnostics

	registryType := d.Get("type").(string)
	switch registryType {
	case "ecr":
		err := c.DeleteOciEcrRegistry(d.Id())
		if err != nil {
			return diag.FromErr(err)
		}
	case "basic":
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
	// Initialize auth with noAuth as default
	credentialType := authConfig["credential_type"].(string)
	authType := "noAuth"
	var username, password string

	// Only set username/password if credential_type is "basic"
	if credentialType == "basic" {
		authType = "basic"
		if val, ok := authConfig["username"]; ok && val != nil {
			username = val.(string)
		}
		if val, ok := authConfig["password"]; ok && val != nil {
			password = val.(string)
		}
	}

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
				Type:     authType,
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

// waitForOciRegistrySync waits for an OCI registry to complete its synchronization
func waitForOciRegistrySync(ctx context.Context, d *schema.ResourceData, uid string, diags diag.Diagnostics, c *client.V1Client, timeoutType string) (diag.Diagnostics, bool) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			"InProgress",
			"Pending",
			"Unknown",
			"", // Handle empty status as pending
		},
		Target: []string{
			"Success",
			"Completed",
		},
		Refresh:    resourceOciRegistrySyncRefreshFunc(c, uid),
		Timeout:    d.Timeout(timeoutType) - 1*time.Minute,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	// Wait, catching any errors
	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		// Handle timeout errors gracefully
		var timeoutErr *retry.TimeoutError
		if errors.As(err, &timeoutErr) {
			// Get current sync status for warning message
			syncStatus, statusErr := c.GetOciBasicRegistrySyncStatus(uid)
			currentStatus := timeoutErr.LastState
			statusMessage := ""

			if statusErr == nil && syncStatus != nil {
				if syncStatus.Status != "" {
					currentStatus = syncStatus.Status
				}
				if syncStatus.Message != "" {
					statusMessage = fmt.Sprintf(" Message: %s", syncStatus.Message)
				}
			}

			if currentStatus == "" {
				currentStatus = "Unknown"
			}

			// Return warning instead of error for timeout
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "OCI registry sync timeout",
				Detail: fmt.Sprintf(
					"OCI registry synchronization timed out after waiting for %v. Current sync status is '%s'.%s "+
						"The registry sync may still be in progress and could eventually complete successfully. "+
						"You may need to increase the timeout or wait for the sync to complete manually.",
					d.Timeout(timeoutType)-1*time.Minute, currentStatus, statusMessage),
			})
			return diags, false
		}

		// Check if this is a sync failure (not a timeout or API error)
		// Get current sync status to provide detailed error information
		syncStatus, statusErr := c.GetOciBasicRegistrySyncStatus(uid)
		if statusErr == nil && syncStatus != nil {
			status := syncStatus.Status
			// Check if the sync explicitly failed
			if status == "Failed" || status == "Error" || status == "failed" || status == "error" {
				errorDetail := fmt.Sprintf("OCI registry synchronization failed with status '%s'.", status)
				if syncStatus.Message != "" {
					errorDetail += fmt.Sprintf("\n\nError details: %s", syncStatus.Message)
				}
				errorDetail += "\n\nPlease check the registry configuration (endpoint, credentials) and try again."

				diags = append(diags, diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  "OCI registry sync failed",
					Detail:   errorDetail,
				})
				return diags, false
			}
		}

		// For other non-timeout errors (API errors, network issues, etc.), return the original error
		return diag.FromErr(err), true
	}
	return nil, false
}

// resourceOciRegistrySyncRefreshFunc returns a retry.StateRefreshFunc that checks the sync status of an OCI registry
func resourceOciRegistrySyncRefreshFunc(c *client.V1Client, uid string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		syncStatus, err := c.GetOciBasicRegistrySyncStatus(uid)
		if err != nil {
			return nil, "", err
		}

		// If sync is not supported, consider it as successful
		if syncStatus != nil && !syncStatus.IsSyncSupported {
			return syncStatus, "Success", nil
		}

		if syncStatus == nil || syncStatus.Status == "" {
			return syncStatus, "", nil
		}

		status := syncStatus.Status

		// Map various status values to our state machine
		switch status {
		case "Success", "Completed", "success", "completed":
			return syncStatus, "Success", nil
		case "Failed", "Error", "failed", "error":
			if syncStatus.Message != "" {
				return syncStatus, status, fmt.Errorf("registry sync failed: %s", syncStatus.Message)
			}
			return syncStatus, status, fmt.Errorf("registry sync failed")
		case "InProgress", "Running", "Syncing", "inprogress", "running", "syncing":
			return syncStatus, "InProgress", nil
		default:
			// Unknown status, treat as pending
			return syncStatus, status, nil
		}
	}
}

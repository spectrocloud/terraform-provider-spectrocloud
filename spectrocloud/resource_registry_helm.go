package spectrocloud

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/go-openapi/strfmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
)

func resourceRegistryHelm() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRegistryHelmCreate,
		ReadContext:   resourceRegistryHelmRead,
		UpdateContext: resourceRegistryHelmUpdate,
		DeleteContext: resourceRegistryHelmDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRegistryHelmImport,
		},
		Description: "Resource for managing Helm registries in Spectro Cloud.",

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
				Description: "The name of the Helm registry. This must be unique.",
			},
			"is_private": {
				Type:         schema.TypeBool,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"is_private", "is_synchronization"},
				Deprecated:   "Use `is_synchronization` instead. This field is retained for backward compatibility and will be removed in a future release.",
				Description:  "Specifies whether the Helm registry is private or public. Private registries require authentication to access. **Deprecated:** use `is_synchronization` here instead.",
			},
			"is_synchronization": {
				Type:         schema.TypeBool,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"is_private", "is_synchronization"},
				Description:  "Specifies whether the Helm registry is private (requiring authentication) and, as a result, synchronized by Palette. Replaces `is_private` for naming parity with `spectrocloud_registry_oci`; the Helm registry API has no independent sync flag, so this maps onto the same underlying value as the deprecated `is_private`. Mutually exclusive with `is_private` — set only one.",
			},
			"endpoint": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The URL endpoint of the Helm registry where the charts are hosted.",
			},
			"credentials": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Description: "Authentication credentials for accessing the Helm registry.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"credential_type": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "The type of authentication used for the Helm registry. Supported values are 'noAuth' for no authentication, 'basic' for username/password, and 'token' for token-based authentication.",
							ValidateFunc: validation.StringInSlice([]string{"noAuth", "basic", "token"}, false),
						},
						"username": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The username for basic authentication. Required if 'credential_type' is set to 'basic'.",
						},
						"password": {
							Type:        schema.TypeString,
							Optional:    true,
							Sensitive:   true,
							Description: "Password for basic auth (credential). Required when credential_type is `basic`.",
						},
						"token": {
							Type:        schema.TypeString,
							Optional:    true,
							Sensitive:   true,
							Description: "Auth token (credential). Required when credential_type is `token`.",
						},
						"tls_config": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "TLS configuration for the registry. If omitted, no TLS configuration is sent.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enabled": {
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     true,
										Description: "Specifies whether TLS is enabled for the connection to the Helm registry. Default value is `true`.",
									},
									"ca": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "The certificate authority (CA) certificate, in PEM format, used to validate the Helm registry's TLS certificate.",
									},
									"certificate": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "The client certificate, in PEM format, used for mutual TLS (mTLS) authentication with the Helm registry.",
									},
									"key": {
										Type:        schema.TypeString,
										Optional:    true,
										Sensitive:   true,
										Description: "The private key, in PEM format, corresponding to the client certificate used for mutual TLS (mTLS) authentication with the Helm registry.",
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
			"wait_for_sync": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "If `true`, Terraform will wait for the Helm registry to complete its initial synchronization before marking the resource as created or updated. Default value is `false`.",
			},
		},
	}
}

func resourceRegistryHelmCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "tenant")
	var diags diag.Diagnostics

	registry := toRegistryEntityHelm(d)
	uid, err := c.CreateHelmRegistry(registry)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(uid)

	// Wait for sync if requested
	if d.Get("wait_for_sync") != nil && d.Get("wait_for_sync").(bool) {
		diagnostics, isError := waitForRegistrySync(ctx, d, uid, diags, c, schema.TimeoutCreate)
		if len(diagnostics) > 0 {
			diags = append(diags, diagnostics...)
		}
		if isError {
			return diagnostics
		}
	}

	return diags
}

// helmRegistryCredentialFieldForRead preserves a sensitive credential field
// (password, token) from state instead of the value the API just returned.
// Palette's GET does not echo back the plaintext secret (empty or masked),
// so writing it into state on every Read produces perpetual plan drift
// against the value configured in HCL. Mirrors the fix already applied to
// spectrocloud_registry_oci for the same root cause (see PLT-2400).
func helmRegistryCredentialFieldForRead(d *schema.ResourceData, field string, apiValue string) interface{} {
	if credsRaw, ok := d.Get("credentials").([]interface{}); ok && len(credsRaw) > 0 {
		if credMap, ok := credsRaw[0].(map[string]interface{}); ok {
			if val, exists := credMap[field]; exists && val != nil {
				return val
			}
		}
	}
	return apiValue
}

// helmRegistryTLSConfigForRead flattens the API TLS payload back into state.
// It mirrors ociBasicTLSConfigForRead: the block is only emitted when the
// practitioner configured one or the API returned something meaningful, so a
// registry stored without TLS does not produce a diff.
func helmRegistryTLSConfigForRead(d *schema.ResourceData, apiTLS *models.V1TLSConfiguration) []interface{} {
	tlsConfig := make([]interface{}, 0, 1)
	if apiTLS == nil {
		return tlsConfig
	}
	hasStateTLS := false
	if credsRaw, ok := d.Get("credentials").([]interface{}); ok && len(credsRaw) > 0 {
		if credMap, ok := credsRaw[0].(map[string]interface{}); ok {
			if tlsRaw, ok := credMap["tls_config"].([]interface{}); ok && len(tlsRaw) > 0 {
				hasStateTLS = true
			}
		}
	}
	hasMeaningfulTLS := apiTLS.Ca != "" || apiTLS.Certificate != "" || apiTLS.Key != "" || apiTLS.InsecureSkipVerify
	if hasStateTLS || hasMeaningfulTLS {
		tlsConfig = append(tlsConfig, map[string]interface{}{
			"enabled":              apiTLS.Enabled,
			"ca":                   apiTLS.Ca,
			"certificate":          apiTLS.Certificate,
			"key":                  apiTLS.Key,
			"insecure_skip_verify": apiTLS.InsecureSkipVerify,
		})
	}
	return tlsConfig
}

func resourceRegistryHelmRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "tenant")
	var diags diag.Diagnostics

	registry, err := c.GetHelmRegistry(d.Id())
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
	// is_synchronization mirrors is_private: the Helm API has no separate sync
	// flag, so both attributes always reflect the same underlying value. Both
	// are Optional+Computed with ExactlyOneOf, so writing the same value into
	// whichever one the practitioner did NOT configure produces no diff.
	if err := d.Set("is_synchronization", registry.Spec.IsPrivate); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("endpoint", registry.Spec.Endpoint); err != nil {
		return diag.FromErr(err)
	}
	waitForSync := false
	if v, ok := d.GetOk("wait_for_sync"); ok {
		waitForSync = v.(bool)
	}
	if err := d.Set("wait_for_sync", waitForSync); err != nil {
		return diag.FromErr(err)
	}

	tlsConfig := helmRegistryTLSConfigForRead(d, registry.Spec.Auth.TLS)

	switch registry.Spec.Auth.Type {
	case "noAuth":
		credentials := make([]interface{}, 0, 1)
		acc := make(map[string]interface{})
		acc["credential_type"] = "noAuth"
		acc["tls_config"] = tlsConfig
		credentials = append(credentials, acc)
		if err := d.Set("credentials", credentials); err != nil {
			return diag.FromErr(err)
		}
	case "basic":
		credentials := make([]interface{}, 0, 1)
		acc := make(map[string]interface{})
		acc["credential_type"] = "basic"
		acc["username"] = registry.Spec.Auth.Username
		acc["password"] = helmRegistryCredentialFieldForRead(d, "password", registry.Spec.Auth.Password.String())
		acc["tls_config"] = tlsConfig
		credentials = append(credentials, acc)
		if err := d.Set("credentials", credentials); err != nil {
			return diag.FromErr(err)
		}
	case "token":
		credentials := make([]interface{}, 0, 1)
		acc := make(map[string]interface{})
		acc["credential_type"] = "token"
		acc["username"] = registry.Spec.Auth.Username
		acc["token"] = helmRegistryCredentialFieldForRead(d, "token", registry.Spec.Auth.Token.String())
		acc["tls_config"] = tlsConfig
		credentials = append(credentials, acc)
		if err := d.Set("credentials", credentials); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func resourceRegistryHelmUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "tenant")
	var diags diag.Diagnostics

	registry := toRegistryHelm(d)
	err := c.UpdateHelmRegistry(d.Id(), registry)
	if err != nil {
		return diag.FromErr(err)
	}

	// Wait for sync if requested
	if d.Get("wait_for_sync") != nil && d.Get("wait_for_sync").(bool) {
		diagnostics, isError := waitForRegistrySync(ctx, d, d.Id(), diags, c, schema.TimeoutUpdate)
		if len(diagnostics) > 0 {
			diags = append(diags, diagnostics...)
		}
		if isError {
			return diagnostics
		}
	}

	return diags
}

func resourceRegistryHelmDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "tenant")
	var diags diag.Diagnostics
	err := c.DeleteHelmRegistry(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

// resolveHelmIsPrivate resolves the single `isPrivate` API value from
// whichever of `is_private` (deprecated) or `is_synchronization` the
// practitioner configured. `ExactlyOneOf` guarantees at most one is present
// in config; the other stays at its Computed zero/prior value, so an OR is
// sufficient to recover the configured value regardless of which attribute
// carries it (see PLT-2401).
func resolveHelmIsPrivate(d *schema.ResourceData) bool {
	return d.Get("is_private").(bool) || d.Get("is_synchronization").(bool)
}

func toRegistryEntityHelm(d *schema.ResourceData) *models.V1HelmRegistryEntity {
	endpoint := d.Get("endpoint").(string)
	isPrivate := resolveHelmIsPrivate(d)
	config := d.Get("credentials").([]interface{})[0].(map[string]interface{})
	return &models.V1HelmRegistryEntity{
		Metadata: &models.V1ObjectMeta{
			Name: d.Get("name").(string),
		},
		Spec: &models.V1HelmRegistrySpecEntity{
			Name:      d.Get("name").(string),
			Auth:      toRegistryHelmCredential(config),
			Endpoint:  &endpoint,
			IsPrivate: isPrivate,
		},
	}
}

func toRegistryHelm(d *schema.ResourceData) *models.V1HelmRegistry {
	endpoint := d.Get("endpoint").(string)
	isPrivate := resolveHelmIsPrivate(d)
	config := d.Get("credentials").([]interface{})[0].(map[string]interface{})
	return &models.V1HelmRegistry{
		Metadata: &models.V1ObjectMeta{
			Name: d.Get("name").(string),
		},
		Spec: &models.V1HelmRegistrySpec{
			Name:      d.Get("name").(string),
			Auth:      toRegistryHelmCredential(config),
			Endpoint:  &endpoint,
			IsPrivate: isPrivate,
		},
	}
}

func toRegistryHelmCredential(regCred map[string]interface{}) *models.V1RegistryAuth {
	auth := &models.V1RegistryAuth{
		Type: "noAuth",
	}

	switch regCred["credential_type"].(string) {
	case "basic":
		auth.Type = "basic"
		auth.Username = regCred["username"].(string)
		auth.Password = strfmt.Password(regCred["password"].(string))
	case "token":
		auth.Type = "token"
		auth.Username = regCred["username"].(string)
		auth.Token = strfmt.Password(regCred["token"].(string))
	}

	// TLS is only sent when configured, so existing configs are unaffected.
	if tlsCfg, ok := regCred["tls_config"].([]interface{}); ok && len(tlsCfg) > 0 && tlsCfg[0] != nil {
		tlsMap := tlsCfg[0].(map[string]interface{})
		auth.TLS = &models.V1TLSConfiguration{
			Enabled:            tlsMap["enabled"].(bool),
			Ca:                 tlsMap["ca"].(string),
			Certificate:        tlsMap["certificate"].(string),
			Key:                tlsMap["key"].(string),
			InsecureSkipVerify: tlsMap["insecure_skip_verify"].(bool),
		}
	}

	return auth
}

// waitForRegistrySync waits for a Helm registry to complete its synchronization
func waitForRegistrySync(ctx context.Context, d *schema.ResourceData, uid string, diags diag.Diagnostics, c *client.V1Client, timeoutType string) (diag.Diagnostics, bool) {
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
		Refresh:    resourceRegistrySyncRefreshFunc(c, uid),
		Timeout:    d.Timeout(timeoutType) - 1*time.Minute,
		MinTimeout: 10 * time.Second,
		Delay:      resolveWaitDelay(30 * time.Second),
	}

	// Wait, catching any errors
	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		// Handle timeout errors gracefully
		var timeoutErr *retry.TimeoutError
		if errors.As(err, &timeoutErr) {
			log.Printf("waitForRegistrySync: timeout occurred, returning warning instead of error")

			// Get current sync status for warning message
			syncStatus, statusErr := c.GetHelmRegistrySyncStatus(uid)
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
				Summary:  "Helm registry sync timeout",
				Detail: fmt.Sprintf(
					"Helm registry synchronization timed out after waiting for %v. Current sync status is '%s'.%s "+
						"The registry sync may still be in progress and could eventually complete successfully. "+
						"You may need to increase the timeout or wait for the sync to complete manually.",
					d.Timeout(timeoutType)-1*time.Minute, currentStatus, statusMessage),
			})
			return diags, false
		}

		// Check if this is a sync failure (not a timeout or API error)
		// Get current sync status to provide detailed error information
		syncStatus, statusErr := c.GetHelmRegistrySyncStatus(uid)
		if statusErr == nil && syncStatus != nil {
			status := syncStatus.Status
			// Check if the sync explicitly failed
			if status == "Failed" || status == "Error" || status == "failed" || status == "error" {
				log.Printf("waitForRegistrySync: registry sync failed with status: %s", status)
				errorDetail := fmt.Sprintf("Helm registry synchronization failed with status '%s'.", status)
				if syncStatus.Message != "" {
					errorDetail += fmt.Sprintf("\n\nError details: %s", syncStatus.Message)
				}
				errorDetail += "\n\nPlease check the registry configuration (endpoint, credentials) and try again."

				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Helm registry sync failed",
					Detail:   errorDetail,
				})
				return diags, true
			}
		}

		// For other non-timeout errors (API errors, network issues, etc.), return the original error
		log.Printf("waitForRegistrySync: unexpected error: %v", err)
		return diag.FromErr(err), true
	}
	return nil, false
}

// resourceRegistrySyncRefreshFunc returns a retry.StateRefreshFunc that checks the sync status of a Helm registry
func resourceRegistrySyncRefreshFunc(c *client.V1Client, uid string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		syncStatus, err := c.GetHelmRegistrySyncStatus(uid)
		if err != nil {
			return nil, "", err
		}

		// If sync is not supported, consider it as successful
		if syncStatus != nil && !syncStatus.IsSyncSupported {
			log.Printf("[DEBUG] Registry sync is not supported, considering as completed")
			return syncStatus, "Success", nil
		}

		if syncStatus == nil || syncStatus.Status == "" {
			log.Printf("[DEBUG] Registry sync status is empty, treating as pending")
			return syncStatus, "", nil
		}

		status := syncStatus.Status
		log.Printf("[DEBUG] Registry sync status: %s", status)

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
			log.Printf("[DEBUG] Unknown registry sync status '%s', treating as pending", status)
			return syncStatus, status, nil
		}
	}
}

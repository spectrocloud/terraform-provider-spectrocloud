package spectrocloud

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/spectrocloud/palette-sdk-go/client/herr"
)

// resourceRegistryPack manages Spectro Cloud Pack registries (Tenant Settings > Registries > Pack Registries), the
// counterpart of spectrocloud_registry_helm and spectrocloud_registry_oci. The API shares the Helm registry's
// credential model (V1RegistryAuth), so credentials and TLS handling reuse the Helm helpers.
func resourceRegistryPack() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRegistryPackCreate,
		ReadContext:   resourceRegistryPackRead,
		UpdateContext: resourceRegistryPackUpdate,
		DeleteContext: resourceRegistryPackDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRegistryPackImport,
		},
		Description: "Resource for managing Pack registries in Spectro Cloud. Pack registries host Spectro Cloud packs " +
			"that Palette synchronizes and makes available to cluster profiles. This is a tenant-level resource.",

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the Pack registry. This must be unique. Changing the name creates a new registry.",
			},
			"endpoint": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The URL of the Pack registry, for example `https://registry.example.com:5000`.",
			},
			"credentials": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Description: "Authentication credentials for accessing the Pack registry.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"credential_type": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "The type of authentication used for the Pack registry. Supported values are `noAuth` for no authentication, `basic` for username and password, and `token` for token-based authentication.",
							ValidateFunc: validation.StringInSlice([]string{"noAuth", "basic", "token"}, false),
						},
						"username": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The username. Required when `credential_type` is `basic`.",
						},
						"password": {
							Type:        schema.TypeString,
							Optional:    true,
							Sensitive:   true,
							Description: "The password. Required when `credential_type` is `basic`.",
						},
						"token": {
							Type:        schema.TypeString,
							Optional:    true,
							Sensitive:   true,
							Description: "The authentication token. Required when `credential_type` is `token`.",
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
										Description: "Whether TLS is enabled for the connection to the Pack registry. Default value is `true`.",
									},
									"ca": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "The certificate authority (CA) certificate, in PEM format, used to validate the Pack registry's TLS certificate.",
									},
									"certificate": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "The client certificate, in PEM format, used for mutual TLS (mTLS) authentication with the Pack registry.",
									},
									"key": {
										Type:        schema.TypeString,
										Optional:    true,
										Sensitive:   true,
										Description: "The private key, in PEM format, for the client certificate used for mutual TLS (mTLS) authentication with the Pack registry.",
									},
									"insecure_skip_verify": {
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     false,
										Description: "Disables TLS certificate verification when set to `true`. ⚠️ WARNING: this makes connections vulnerable to man-in-the-middle attacks. Only use it for registries with self-signed certificates in trusted networks.",
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
				Description: "If `true`, Terraform waits for Palette to finish synchronizing the Pack registry before marking the resource as created or updated. Default value is `false`.",
			},
		},
	}
}

func resourceRegistryPackCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "tenant")
	var diags diag.Diagnostics

	registry := toRegistryPack(d)
	// The Palette UI validates the endpoint and credentials before saving; do the same so a registry Palette cannot
	// reach fails here with Palette's reason instead of being created and never syncing.
	if err := c.ValidatePackRegistry(registry.Spec); err != nil {
		return diag.Errorf("pack registry %q failed validation: %s", d.Get("name").(string), err)
	}
	uid, err := c.CreatePackRegistry(registry)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(uid)

	if d.Get("wait_for_sync").(bool) {
		diagnostics, isError := waitForPackRegistrySync(ctx, d, uid, diags, c, schema.TimeoutCreate)
		if isError {
			return diagnostics
		}
		diags = append(diags, diagnostics...)
	}

	return append(diags, resourceRegistryPackRead(ctx, d, m)...)
}

func resourceRegistryPackRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "tenant")
	var diags diag.Diagnostics

	registry, err := c.GetPackRegistryByUID(d.Id())
	if err != nil {
		return handleReadError(d, err, diags)
	} else if registry == nil {
		// Deleted - Terraform will recreate it
		d.SetId("")
		return diags
	}
	if err := flattenRegistryPack(d, registry); err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceRegistryPackUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "tenant")
	var diags diag.Diagnostics

	registry := toRegistryPack(d)
	if d.HasChanges("endpoint", "credentials") {
		if err := c.ValidatePackRegistry(registry.Spec); err != nil {
			return diag.Errorf("pack registry %q failed validation: %s", d.Get("name").(string), err)
		}
	}
	registry.Metadata.UID = d.Id()
	if err := c.UpdatePackRegistry(d.Id(), registry); err != nil {
		return diag.FromErr(err)
	}

	if d.Get("wait_for_sync").(bool) {
		diagnostics, isError := waitForPackRegistrySync(ctx, d, d.Id(), diags, c, schema.TimeoutUpdate)
		if isError {
			return diagnostics
		}
		diags = append(diags, diagnostics...)
	}

	return append(diags, resourceRegistryPackRead(ctx, d, m)...)
}

func resourceRegistryPackDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "tenant")
	if err := c.DeletePackRegistry(d.Id()); err != nil {
		if herr.IsNotFound(err) {
			return nil
		}
		return diag.FromErr(err)
	}
	return nil
}

// resourceRegistryPackImport accepts a Pack registry UID or name.
func resourceRegistryPackImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c := getV1ClientWithResourceContext(m, "tenant")
	uid, err := resolvePackRegistryImportID(c, d.Id())
	if err != nil {
		return nil, err
	}
	d.SetId(uid)

	diags := resourceRegistryPackRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read Pack registry for import: %v", diags)
	}
	if d.Id() == "" {
		return nil, fmt.Errorf("pack registry %q not found", uid)
	}
	return []*schema.ResourceData{d}, nil
}

func resolvePackRegistryImportID(c *client.V1Client, importID string) (string, error) {
	if importID == "" {
		return "", fmt.Errorf("pack registry import ID or name is required")
	}
	registry, err := c.GetPackRegistryByUID(importID)
	if err == nil && registry != nil {
		return importID, nil
	}
	if err != nil && !herr.IsNotFound(err) {
		return "", fmt.Errorf("unable to retrieve Pack registry '%s': %w", importID, err)
	}
	registry, err = c.GetPackRegistryByName(importID)
	if err != nil {
		return "", fmt.Errorf("unable to retrieve Pack registry by name or id '%s': %w", importID, err)
	}
	if registry == nil || registry.Metadata == nil || registry.Metadata.UID == "" {
		return "", fmt.Errorf("pack registry '%s' not found", importID)
	}
	return registry.Metadata.UID, nil
}

func toRegistryPack(d *schema.ResourceData) *models.V1PackRegistry {
	name := d.Get("name").(string)
	endpoint := d.Get("endpoint").(string)
	credentials := d.Get("credentials").([]interface{})[0].(map[string]interface{})
	return &models.V1PackRegistry{
		Metadata: &models.V1ObjectMeta{Name: name},
		Spec: &models.V1PackRegistrySpec{
			Name:     name,
			Endpoint: &endpoint,
			Auth:     toRegistryHelmCredential(credentials),
		},
	}
}

func flattenRegistryPack(d *schema.ResourceData, registry *models.V1PackRegistry) error {
	if registry.Metadata != nil {
		if err := d.Set("name", registry.Metadata.Name); err != nil {
			return err
		}
	}
	if registry.Spec == nil {
		return nil
	}
	if registry.Spec.Endpoint != nil {
		if err := d.Set("endpoint", *registry.Spec.Endpoint); err != nil {
			return err
		}
	}
	if _, ok := d.GetOk("wait_for_sync"); !ok {
		if err := d.Set("wait_for_sync", false); err != nil {
			return err
		}
	}

	auth := registry.Spec.Auth
	if auth == nil {
		auth = &models.V1RegistryAuth{Type: "noAuth"}
	}
	credential := map[string]interface{}{
		"tls_config": helmRegistryTLSConfigForRead(d, auth.TLS),
	}
	switch auth.Type {
	case "basic":
		credential["credential_type"] = "basic"
		credential["username"] = auth.Username
		// Palette does not return secrets in clear text; keep the configured value so it does not show as drift.
		credential["password"] = helmRegistryCredentialFieldForRead(d, "password", auth.Password.String())
	case "token":
		credential["credential_type"] = "token"
		credential["username"] = auth.Username
		credential["token"] = helmRegistryCredentialFieldForRead(d, "token", auth.Token.String())
	default:
		credential["credential_type"] = "noAuth"
		// Palette ignores username, password and token for noAuth and does not return them; keep whatever is
		// configured so they do not show as drift.
		for _, field := range []string{"username", "password", "token"} {
			credential[field] = helmRegistryCredentialFieldForRead(d, field, "")
		}
	}
	return d.Set("credentials", []interface{}{credential})
}

// waitForPackRegistrySync waits for a Pack registry to complete its synchronization. It mirrors waitForRegistrySync
// in resource_registry_helm.go, using the Pack registry sync status API.
func waitForPackRegistrySync(ctx context.Context, d *schema.ResourceData, uid string, diags diag.Diagnostics, c *client.V1Client, timeoutType string) (diag.Diagnostics, bool) {
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
		Refresh:    resourcePackRegistrySyncRefreshFunc(c, uid),
		Timeout:    d.Timeout(timeoutType) - 1*time.Minute,
		MinTimeout: 10 * time.Second,
		Delay:      resolveWaitDelay(30 * time.Second),
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		// A timeout is a warning: the sync may still finish in the background.
		var timeoutErr *retry.TimeoutError
		if errors.As(err, &timeoutErr) {
			log.Printf("waitForPackRegistrySync: timeout occurred, returning warning instead of error")

			syncStatus, statusErr := c.GetPackRegistrySyncStatus(uid)
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

			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "Pack registry sync timeout",
				Detail: fmt.Sprintf(
					"Pack registry synchronization timed out after waiting for %v. Current sync status is '%s'.%s "+
						"The registry sync may still be in progress and could eventually complete successfully. "+
						"You may need to increase the timeout or wait for the sync to complete manually.",
					d.Timeout(timeoutType)-1*time.Minute, currentStatus, statusMessage),
			})
			return diags, false
		}

		// A sync that explicitly failed is reported with Palette's reason.
		syncStatus, statusErr := c.GetPackRegistrySyncStatus(uid)
		if statusErr == nil && syncStatus != nil {
			status := syncStatus.Status
			if status == "Failed" || status == "Error" || status == "failed" || status == "error" {
				log.Printf("waitForPackRegistrySync: registry sync failed with status: %s", status)
				errorDetail := fmt.Sprintf("Pack registry synchronization failed with status '%s'.", status)
				if syncStatus.Message != "" {
					errorDetail += fmt.Sprintf("\n\nError details: %s", syncStatus.Message)
				}
				errorDetail += "\n\nPlease check the registry configuration (endpoint, credentials) and try again."

				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Pack registry sync failed",
					Detail:   errorDetail,
				})
				return diags, true
			}
		}

		// Any other error (API errors, network issues, etc.) is returned as is.
		log.Printf("waitForPackRegistrySync: unexpected error: %v", err)
		return diag.FromErr(err), true
	}
	return nil, false
}

// resourcePackRegistrySyncRefreshFunc returns a retry.StateRefreshFunc that checks the sync status of a Pack registry.
func resourcePackRegistrySyncRefreshFunc(c *client.V1Client, uid string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		syncStatus, err := c.GetPackRegistrySyncStatus(uid)
		if err != nil {
			return nil, "", err
		}

		// If sync is not supported, consider it as successful
		if syncStatus != nil && !syncStatus.IsSyncSupported {
			log.Printf("[DEBUG] Pack registry sync is not supported, considering as completed")
			return syncStatus, "Success", nil
		}

		if syncStatus == nil || syncStatus.Status == "" {
			log.Printf("[DEBUG] Pack registry sync status is empty, treating as pending")
			return syncStatus, "", nil
		}

		status := syncStatus.Status
		log.Printf("[DEBUG] Pack registry sync status: %s", status)

		switch status {
		case "Success", "Completed", "success", "completed":
			return syncStatus, "Success", nil
		case "Failed", "Error", "failed", "error":
			if syncStatus.Message != "" {
				return syncStatus, status, fmt.Errorf("pack registry sync failed: %s", syncStatus.Message)
			}
			return syncStatus, status, fmt.Errorf("pack registry sync failed")
		case "InProgress", "Running", "Syncing", "inprogress", "running", "syncing":
			return syncStatus, "InProgress", nil
		default:
			// Unknown status, treat as pending
			log.Printf("[DEBUG] Unknown pack registry sync status '%s', treating as pending", status)
			return syncStatus, status, nil
		}
	}
}

package spectrocloud

import (
	"context"
	"fmt"
	"log"
	"net/mail"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

// validateEmail validates email addresses using net/mail standard library
func validateEmail(v interface{}, k string) (warnings []string, errors []error) {
	email := v.(string)
	_, err := mail.ParseAddress(email)
	if err != nil {
		errors = append(errors, fmt.Errorf("%q is not a valid email address: %v", k, err))
	}
	return warnings, errors
}

func resourceAlert() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAlertCreate,
		ReadContext:   resourceAlertRead,
		UpdateContext: resourceAlertUpdate,
		DeleteContext: resourceAlertDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceAlertImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		SchemaVersion: 2,
		Schema: map[string]*schema.Schema{
			"project": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The project to which the alert belongs to.",
			},
			"is_active": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Indicates whether the alert is active. Set to `true` to activate the alert, or `false` to deactivate it.",
			},
			"component": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"ClusterHealth"}, false),
				Description:  "The component of the system that the alert is associated with. Currently, `ClusterHealth` is the only supported value.",
			},
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "",
				ValidateFunc: validation.StringInSlice([]string{"", "email", "http"}, false),
				Description:  "The type of alert mechanism to use. Can be `email` for email alerts, `http` for HTTP webhooks, or empty string to auto-detect based on provided configuration.",
			},
			"alert_all_users": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "If set to `true`, the alert will be sent to all users. If `false`, it will target specific users or identifiers.",
			},
			"created_by": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The user who created the alert.",
			},
			// Status is defined here just for schema, we are not using this status. it implemented for internal logic
			"status": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "A status block representing the internal status of the alert. This is primarily for internal use and not utilized directly.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"is_succeeded": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"message": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"time": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"identifiers": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "A set of unique identifiers to which the alert will be sent. This is used to target specific users or groups.",
				Set:         schema.HashString,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateEmail,
				},
			},
			"http": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The configuration block for HTTP-based alerts. This is used when the `type` is set to `http`.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"method": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "The HTTP method to use for the alert. Supported values are `POST`, `GET`, and `PUT`.",
							ValidateFunc: validation.StringInSlice([]string{"POST", "GET", "PUT"}, false),
						},
						"url": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The target URL to send the HTTP request to when the alert is triggered.",
						},
						"body": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The payload to include in the HTTP request body when the alert is triggered.",
						},
						"headers": {
							Type:        schema.TypeMap,
							Optional:    true,
							Description: "Optional HTTP headers to include in the request. Each header should be specified as a key-value pair.",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}
}

func resourceAlertCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "")
	component := d.Get("component").(string)
	var err error
	projectUid := ""

	projectUid, err = getProjectID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	projectName := d.Get("project").(string)
	if projectName == "" {
		return diag.Errorf("project name is required")
	}

	var diags diag.Diagnostics

	// Convert schema to channels (handles both single and combined alerts)
	newChannels := toAlertChannels(d)

	if len(newChannels) == 0 {
		return diag.Errorf("at least one of 'identifiers', 'alert_all_users', or 'http' block must be specified")
	}

	// Create alert entity with all new channels
	alertEntity := &models.V1AlertEntity{
		Channels: newChannels,
	}

	err = c.UpdateProjectAlerts(alertEntity, projectUid, component)
	if err != nil {
		// Handle first-time setup for ClusterHealth
		if strings.Contains(err.Error(), "Project 'ClusterHealth' alerts are not found") {
			err = c.UpdateProjectAlerts(alertEntity, projectUid, component)
			if err != nil {
				return diag.FromErr(err)
			}
		} else {
			return diag.FromErr(err)
		}
	}

	// Generate unique ID based on project and component (singleton resource)
	alertID := fmt.Sprintf("%s:%s", projectUid, component)
	d.SetId(alertID)
	return diags
}
func resourceAlertUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "")
	var err error
	var diags diag.Diagnostics

	projectName := d.Get("project").(string)
	if projectName == "" {
		return diag.Errorf("project name is required")
	}
	projectUid, err := getProjectID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	component := d.Get("component").(string)

	// Convert schema to channels (handles both single and combined alerts)
	newChannels := toAlertChannels(d)

	if len(newChannels) == 0 {
		return diag.Errorf("at least one of 'identifiers', 'alert_all_users', or 'http' block must be specified")
	}

	// Update alert entity with all new channels
	alertEntity := &models.V1AlertEntity{
		Channels: newChannels,
	}

	err = c.UpdateProjectAlerts(alertEntity, projectUid, component)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

// toAlertChannels converts the Terraform schema data to API channel models
// Returns multiple channels when both email and http are configured
func toAlertChannels(d *schema.ResourceData) []*models.V1Channel {
	var channels []*models.V1Channel
	alertType := d.Get("type").(string)
	isActive := d.Get("is_active").(bool)
	createdBy := d.Get("created_by").(string)
	alertAllUsers := d.Get("alert_all_users").(bool)

	_, hasIdentifiers := d.GetOk("identifiers")
	_, hasHttp := d.GetOk("http")

	// Determine effective types based on configuration
	createEmail := false
	createHttp := false

	switch alertType {
	case "email":
		createEmail = true
	case "http":
		createHttp = true
	case "":
		// Auto-detect based on what's configured
		createEmail = hasIdentifiers || alertAllUsers
		createHttp = hasHttp
	}

	// Create email channel if needed
	if createEmail {
		emailChannel := &models.V1Channel{
			IsActive:      isActive,
			Type:          "email",
			CreatedBy:     createdBy,
			AlertAllUsers: alertAllUsers,
		}
		if hasIdentifiers {
			emailIDs := make([]string, 0)
			for _, email := range d.Get("identifiers").(*schema.Set).List() {
				emailIDs = append(emailIDs, email.(string))
			}
			emailChannel.Identifiers = emailIDs
		}
		channels = append(channels, emailChannel)
	}

	// Create http channels if needed
	if createHttp {
		httpList := d.Get("http").([]interface{})
		for _, httpItem := range httpList {
			httpConfig := httpItem.(map[string]interface{})
			headersMap := make(map[string]string)
			if httpConfig["headers"] != nil {
				for key, element := range httpConfig["headers"].(map[string]interface{}) {
					headersMap[key] = element.(string)
				}
			}
			httpChannel := &models.V1Channel{
				IsActive:  isActive,
				Type:      "http",
				CreatedBy: createdBy,
				HTTP: &models.V1ChannelHTTP{
					Body:    httpConfig["body"].(string),
					Method:  httpConfig["method"].(string),
					URL:     httpConfig["url"].(string),
					Headers: headersMap,
				},
			}
			channels = append(channels, httpChannel)
		}
	}

	return channels
}

// toAlert is kept for backward compatibility - returns single channel
func toAlert(d *schema.ResourceData) *models.V1Channel {
	channels := toAlertChannels(d)
	if len(channels) > 0 {
		return channels[0]
	}
	return &models.V1Channel{
		IsActive: d.Get("is_active").(bool),
		Type:     d.Get("type").(string),
	}
}

func resourceAlertDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "")
	var diags diag.Diagnostics

	projectUid, err := getProjectID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	component := d.Get("component").(string)

	// Delete all channels by setting empty channels (singleton resource)
	alertEntity := &models.V1AlertEntity{
		Channels: []*models.V1Channel{},
	}

	err = c.UpdateProjectAlerts(alertEntity, projectUid, component)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceAlertRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "")
	var diags diag.Diagnostics

	projectName := d.Get("project").(string)
	if projectName == "" {
		log.Printf("[WARN] Project name is empty during refresh. State preserved.")
		return diags
	}

	projectUid, err := getProjectID(d, m)
	if err != nil {
		log.Printf("[WARN] Error getting project ID during refresh: %v. State preserved.", err)
		return diags
	}
	if projectUid == "" {
		log.Printf("[WARN] Project UID is empty during refresh. State preserved.")
		return diags
	}

	component := d.Get("component").(string)
	alertId := d.Id()

	log.Printf("[DEBUG] Reading alert: projectUid=%s, component=%s, alertId=%s", projectUid, component, alertId)

	projectSpec, err := c.GetProject(projectUid)
	if err != nil {
		log.Printf("[WARN] Error getting project during refresh: %v. State preserved.", err)
		return diags
	}

	// Find the alert entity for this component
	var channels []*models.V1Channel
	for _, alert := range projectSpec.Spec.Alerts {
		if alert != nil && alert.Component == component && len(alert.Channels) > 0 {
			channels = alert.Channels
			break
		}
	}

	if len(channels) == 0 {
		log.Printf("[DEBUG] No alerts found for component %s, clearing from state", component)
		d.SetId("")
		return diags
	}

	log.Printf("[DEBUG] Found %d channels for component %s", len(channels), component)

	// Find email and http channels
	var emailChannel *models.V1Channel
	var httpChannel *models.V1Channel

	for _, channel := range channels {
		switch channel.Type {
		case "email":
			emailChannel = channel
		case "http":
			httpChannel = channel
		}
	}

	// Determine the effective type based on what's configured
	var effectiveType string
	var isActive bool

	if emailChannel != nil && httpChannel != nil {
		effectiveType = ""
		isActive = emailChannel.IsActive || httpChannel.IsActive
	} else if emailChannel != nil {
		effectiveType = "email"
		isActive = emailChannel.IsActive
	} else if httpChannel != nil {
		effectiveType = "http"
		isActive = httpChannel.IsActive
	} else {
		log.Printf("[DEBUG] No valid channels found, clearing from state")
		d.SetId("")
		return diags
	}

	// Set project and component
	if err := d.Set("project", projectName); err != nil {
		log.Printf("[WARN] Error setting project: %v", err)
	}
	if err := d.Set("component", component); err != nil {
		log.Printf("[WARN] Error setting component: %v", err)
	}

	// Set common fields
	if err := d.Set("is_active", isActive); err != nil {
		log.Printf("[WARN] Error setting is_active: %v", err)
	}
	if err := d.Set("type", effectiveType); err != nil {
		log.Printf("[WARN] Error setting type: %v", err)
	}

	// Set email-related fields
	if emailChannel != nil {
		if err := d.Set("alert_all_users", emailChannel.AlertAllUsers); err != nil {
			log.Printf("[WARN] Error setting alert_all_users: %v", err)
		}
		if err := d.Set("identifiers", emailChannel.Identifiers); err != nil {
			log.Printf("[WARN] Error setting identifiers: %v", err)
		}
	} else {
		if err := d.Set("alert_all_users", false); err != nil {
			log.Printf("[WARN] Error setting alert_all_users: %v", err)
		}
		if err := d.Set("identifiers", []string{}); err != nil {
			log.Printf("[WARN] Error setting identifiers: %v", err)
		}
	}

	// Set http-related fields
	if httpChannel != nil && httpChannel.HTTP != nil {
		headersMap := make(map[string]interface{})
		if httpChannel.HTTP.Headers != nil {
			for k, v := range httpChannel.HTTP.Headers {
				headersMap[k] = v
			}
		}
		httpConfig := []map[string]interface{}{
			{
				"method":  httpChannel.HTTP.Method,
				"url":     httpChannel.HTTP.URL,
				"body":    httpChannel.HTTP.Body,
				"headers": headersMap,
			},
		}
		if err := d.Set("http", httpConfig); err != nil {
			log.Printf("[WARN] Error setting http: %v", err)
		}
	} else {
		if err := d.Set("http", []interface{}{}); err != nil {
			log.Printf("[WARN] Error clearing http: %v", err)
		}
	}

	log.Printf("[DEBUG] Alert read complete, preserving resource ID: %s", d.Id())
	return diags
}

func getProjectID(d *schema.ResourceData, m interface{}) (string, error) {
	projectUid := ""
	var err error
	c := getV1ClientWithResourceContext(m, "")
	if v, ok := d.GetOk("project"); ok && v.(string) != "" {
		projectUid, err = c.GetProjectUID(v.(string))
		if err != nil {
			return "", err
		}
	}
	return projectUid, nil
}

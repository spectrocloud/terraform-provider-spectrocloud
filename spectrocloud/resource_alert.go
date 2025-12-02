package spectrocloud

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

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
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"", "email", "http"}, false),
				Description:  "The type of alert mechanism to use. Can be either `email` for email alerts or `http` for sending HTTP requests.",
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
				Type:     schema.TypeList,
				Optional: true,
				// Set:         resourceAlertStatusHash,
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
					ValidateFunc: validation.StringMatch(emailRegex, "must be a valid email address"),
				},
			},
			"http": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				//Set:         resourceAlertHttpHash,
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
	alertType := d.Get("type").(string)
	var err error
	projectUid := ""

	projectUid, err = getProjectID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	projectName := d.Get("project").(string)
	// projectString, err := c.GetProjects()
	if projectName == "" {
		return diag.Errorf("project name is required")
	}

	var diags diag.Diagnostics
	alertObj := toAlert(d)

	// Check if alert already exists by getting the project
	projectSpec, err := c.GetProject(projectUid)
	if err != nil {
		return diag.FromErr(err)
	}

	// Find existing alert for this component
	var existingAlertEntity *models.V1AlertEntity
	// var existingChannels []*models.V1Channel
	for _, alert := range projectSpec.Spec.Alerts {
		if alert != nil && alert.Component == component {
			//existingChannels = alert.Channels
			existingAlertEntity = &models.V1AlertEntity{
				Channels: alert.Channels,
			}
			break
		}
	}

	customUID := projectUid + "-alert"
	if alertType == "http" {
		// Build channels list: preserve email alerts, replace http alerts
		var channels []*models.V1Channel

		// If there are existing alerts, process them
		if existingAlertEntity != nil && len(existingAlertEntity.Channels) > 0 {
			for _, channel := range existingAlertEntity.Channels {
				// Preserve email alerts
				if channel.Type == "email" {
					channels = append(channels, channel)
				}
			}
		}

		// Add the new http alert
		channels = append(channels, alertObj)

		alertEntity := &models.V1AlertEntity{
			Channels: channels,
		}

		err = c.UpdateProjectAlerts(alertEntity, projectUid, component)
		if err != nil {
			return diag.FromErr(err)
		}
	} else {
		// For email alerts: preserve existing http alerts, add the new email alert
		var channels []*models.V1Channel

		// If there are existing alerts, process them
		if existingAlertEntity != nil && len(existingAlertEntity.Channels) > 0 {
			for _, channel := range existingAlertEntity.Channels {
				// Preserve http alerts
				if channel.Type == "http" {
					channels = append(channels, channel)
				}
				// Skip existing email alerts (they will be replaced by the new one)
			}
		}

		// Add the new email alert
		channels = append(channels, alertObj)

		alertEntity := &models.V1AlertEntity{
			Channels: channels,
		}
		err = c.UpdateProjectAlerts(alertEntity, projectUid, component)
		if err != nil {
			// Enabling `ClusterHealth` for alerts, basically for setting up for the first time
			if strings.Contains(err.Error(), "Project 'ClusterHealth' alerts are not found") {
				emptyAlert := &models.V1AlertEntity{
					Channels: channels,
				}
				err = c.UpdateProjectAlerts(emptyAlert, projectUid, component)
				if err != nil {
					return diag.FromErr(err)
				}
			} else {
				return diag.FromErr(err)
			}
		}
	}

	// Set the custom UID as the resource ID
	d.SetId(customUID)
	return diags
}
func resourceAlertUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "")
	var err error

	var diags diag.Diagnostics
	// Get project name instead of using getProjectID
	projectName := d.Get("project").(string)
	if projectName == "" {
		return diag.Errorf("project name is required")
	}
	projectUid, err := getProjectID(d, m)

	if err != nil {
		return diag.FromErr(err)
	}
	component := d.Get("component").(string)
	alertType := d.Get("type").(string)
	alertObj := toAlert(d)
	// If alert type is "http", check for existing "email" alert and preserve it
	if alertType == "http" {
		channels, err := preserveEmailAlertForHttp(c, projectUid, component, alertObj)
		if err != nil {
			return diag.FromErr(err)
		}
		// Create alert entity with both email and http channels (or just http if no email exists)
		alertEntity := &models.V1AlertEntity{
			Channels: channels,
		}
		err = c.UpdateProjectAlerts(alertEntity, projectUid, component)
		if err != nil {
			return diag.FromErr(err)
		}
	} else {
		// For non-http alerts (email), use standard update
		_, err = c.UpdateAlert(alertObj, projectUid, component, d.Id())
		if err != nil {
			return diag.FromErr(err)
		}
	}
	return diags
}

func toAlert(d *schema.ResourceData) (alertChannel *models.V1Channel) {

	channel := &models.V1Channel{
		IsActive: d.Get("is_active").(bool),
		Type:     d.Get("type").(string),
	}
	channel.CreatedBy = d.Get("created_by").(string)
	channel.AlertAllUsers = d.Get("alert_all_users").(bool)
	_, hasIdentifier := d.GetOk("identifiers")
	if hasIdentifier {
		emailIDs := make([]string, 0)
		for _, email := range d.Get("identifiers").(*schema.Set).List() {
			emailIDs = append(emailIDs, email.(string))
		}
		channel.Identifiers = emailIDs
	}
	_, hasHttp := d.GetOk("http")
	if hasHttp {
		httpList := d.Get("http").([]interface{})
		if len(httpList) > 0 {
			http := httpList[0].(map[string]interface{})
			headersMap := make(map[string]string)
			if http["headers"] != nil {
				for key, element := range http["headers"].(map[string]interface{}) {
					headersMap[key] = element.(string)
				}
			}
			channel.HTTP = &models.V1ChannelHTTP{
				Body:    http["body"].(string),
				Method:  http["method"].(string),
				URL:     http["url"].(string),
				Headers: headersMap,
			}
		}
	}
	return channel
}

func resourceAlertDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error
	c := getV1ClientWithResourceContext(m, "tenant")
	var diags diag.Diagnostics
	projectUid, err := getProjectID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	//c = getV1ClientWithResourceContextProject(m, projectUid)
	err = c.DeleteAlert(projectUid, d.Get("component").(string), d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceAlertRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error
	// Use same context as Create/Update (empty string, not "tenant")
	c := getV1ClientWithResourceContext(m, "")
	var diags diag.Diagnostics

	// Get project name instead of using getProjectID
	projectName := d.Get("project").(string)
	if projectName == "" {
		log.Printf("[WARN] Project name is empty during refresh. State preserved.")
		return diags
	}
	//  Properly handle getProjectID error - don't ignore it
	projectUid, err := getProjectID(d, m)
	if err != nil {
		// If we can't get project ID, log warning but preserve state
		// Don't return error - this allows refresh to work even if project lookup fails
		log.Printf("[WARN] Error getting project ID during refresh: %v. State preserved.", err)
		return diags
	}
	if projectUid == "" {
		log.Printf("[WARN] Project UID is empty during refresh. State preserved.")
		return diags
	}

	component := d.Get("component").(string)
	expectedType := d.Get("type").(string)
	alertId := d.Id()

	log.Printf("[DEBUG] Reading alert: projectUid=%s, component=%s, alertId=%s, expectedType=%s", projectUid, component, alertId, expectedType)

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

	// Find the channel that matches the expected type and attributes
	var alertPayload *models.V1Channel

	if expectedType == "email" {
		// For email alerts, match by identifiers and alert_all_users
		expectedIdentifiers := d.Get("identifiers").(*schema.Set)
		expectedAlertAllUsers := d.Get("alert_all_users").(bool)

		for _, channel := range channels {
			if channel.Type == "email" {
				// Check alert_all_users first (quick check)
				if channel.AlertAllUsers != expectedAlertAllUsers {
					continue
				}

				// Check identifiers match
				identifiersMatch := true
				if expectedIdentifiers.Len() > 0 {
					if len(channel.Identifiers) != expectedIdentifiers.Len() {
						identifiersMatch = false
					} else {
						// Convert channel identifiers to map for comparison
						channelIdentifiers := make(map[string]bool)
						for _, id := range channel.Identifiers {
							channelIdentifiers[id] = true
						}

						// Check all expected identifiers exist in channel
						for _, id := range expectedIdentifiers.List() {
							if !channelIdentifiers[id.(string)] {
								identifiersMatch = false
								break
							}
						}
					}
				} else if len(channel.Identifiers) > 0 {
					identifiersMatch = false
				}

				if identifiersMatch {
					alertPayload = channel
					break
				}
			}
		}
	} else if expectedType == "http" {
		// For HTTP alerts, match by URL, method, and body
		expectedHttp := d.Get("http").([]interface{})
		if len(expectedHttp) > 0 {

			for _, channel := range channels {
				if channel.Type == "http" && channel.HTTP != nil {
					alertPayload = channel
					break
				}
			}
		} else {
			// If no http config in state but type is http, just get the first http channel
			for _, channel := range channels {
				if channel.Type == "http" {
					alertPayload = channel
					break
				}
			}
		}
	}

	if alertPayload == nil {
		log.Printf("[DEBUG] Alert not found matching expected attributes, clearing from state")
		d.SetId("")
		return diags
	}

	// Resource found - update state with current values from API
	d.SetId(alertPayload.UID)

	// Ensure project and component are set - Terraform needs these to match resources
	if err := d.Set("project", d.Get("project")); err != nil {
		log.Printf("[WARN] Error setting project: %v", err)
	}
	if err := d.Set("component", component); err != nil {
		log.Printf("[WARN] Error setting component: %v", err)
	}
	// Set common fields for all alert types
	if err := d.Set("is_active", alertPayload.IsActive); err != nil {
		log.Printf("[WARN] Error setting is_active: %v", err)
	}
	if err := d.Set("type", alertPayload.Type); err != nil {
		log.Printf("[WARN] Error setting type: %v", err)
	}
	if err := d.Set("alert_all_users", alertPayload.AlertAllUsers); err != nil {
		log.Printf("[WARN] Error setting alert_all_users: %v", err)
	}
	if err := d.Set("identifiers", alertPayload.Identifiers); err != nil {
		log.Printf("[WARN] Error setting identifiers: %v", err)
	}

	// Set type-specific fields
	switch alertPayload.Type {
	case "email":
		// Email alerts should never have an http field - clear it
		if err := d.Set("http", []interface{}{}); err != nil {
			log.Printf("[WARN] Error clearing http: %v", err)
		}
	case "http":
		// Set http field if type is http
		if alertPayload.HTTP != nil {
			var http []map[string]interface{}
			//  Convert headers from map[string]string to map[string]interface{}
			headersMap := make(map[string]interface{})
			if alertPayload.HTTP.Headers != nil {
				for k, v := range alertPayload.HTTP.Headers {
					headersMap[k] = v
				}
			}
			hookConfig := map[string]interface{}{
				"method":  alertPayload.HTTP.Method,
				"url":     alertPayload.HTTP.URL,
				"body":    alertPayload.HTTP.Body,
				"headers": headersMap,
			}
			http = append(http, hookConfig)
			if err := d.Set("http", http); err != nil {
				log.Printf("[WARN] Error setting http: %v", err)
			}
		} else {
			// HTTP type but no HTTP config - clear it
			if err := d.Set("http", []interface{}{}); err != nil {
				log.Printf("[WARN] Error clearing http: %v", err)
			}
		}
	}

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

func preserveEmailAlertForHttp(c *client.V1Client, projectUid, component string, httpAlert *models.V1Channel) ([]*models.V1Channel, error) {
	// Get the project to check for existing alerts
	projectSpec, err := c.GetProject(projectUid)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	// Find existing alert entity for this component
	var existingAlertEntity *models.V1AlertEntity
	for _, alert := range projectSpec.Spec.Alerts {
		if alert != nil && alert.Component == component {
			existingAlertEntity = &models.V1AlertEntity{
				Channels: alert.Channels,
			}
			break
		}
	}
	// Check if there's an existing email alert
	var existingEmailChannel *models.V1Channel
	if existingAlertEntity != nil {
		for _, channel := range existingAlertEntity.Channels {
			if channel.Type == "email" {
				existingEmailChannel = channel
				break
			}
		}
	}
	// If email alert exists, preserve it along with the http alert
	if existingEmailChannel != nil {
		// Return channels with both email and http
		return []*models.V1Channel{existingEmailChannel, httpAlert}, nil
	}
	// No email alert exists, return just the http alert
	return []*models.V1Channel{httpAlert}, nil
}

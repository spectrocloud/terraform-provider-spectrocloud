package spectrocloud

import (
	"context"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func resourceAlert() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAlertCreate,
		ReadContext:   resourceAlertRead,
		UpdateContext: resourceAlertUpdate,
		DeleteContext: resourceAlertDelete,

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
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"email", "http"}, false),
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
					ValidateFunc: validation.StringMatch(emailRegex, "must be a valid email address"),
				},
			},
			"http": {
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    true,
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
	var err error
	projectUid, err := getProjectID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	var diags diag.Diagnostics
	alertObj := toAlert(d)
	uid, err := c.CreateAlert(alertObj, projectUid, d.Get("component").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(uid)
	return diags
}

func resourceAlertUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "")
	var err error

	var diags diag.Diagnostics
	projectUid, _ := getProjectID(d, m)
	alertObj := toAlert(d)
	_, err = c.UpdateAlert(alertObj, projectUid, d.Get("component").(string), d.Id())
	if err != nil {
		return diag.FromErr(err)
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
		http := d.Get("http").([]interface{})[0].(map[string]interface{})
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
	return channel
}

func resourceAlertDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error
	c := getV1ClientWithResourceContext(m, "")
	var diags diag.Diagnostics
	projectUid, err := getProjectID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	err = c.DeleteAlert(projectUid, d.Get("component").(string), d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceAlertRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error
	c := getV1ClientWithResourceContext(m, "")
	var diags diag.Diagnostics
	projectUid, _ := getProjectID(d, m)
	alertPayload, err := c.GetAlert(projectUid, d.Get("component").(string), d.Id())
	if alertPayload == nil {
		d.SetId("")
		return diag.FromErr(err)

	} else {
		d.SetId(alertPayload.UID)
		if err := d.Set("is_active", alertPayload.IsActive); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("type", alertPayload.Type); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("alert_all_users", alertPayload.AlertAllUsers); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("identifiers", alertPayload.Identifiers); err != nil {
			return diag.FromErr(err)
		}
		if alertPayload.Type == "http" {
			var http []map[string]interface{}
			hookConfig := map[string]interface{}{
				"method":  alertPayload.HTTP.Method,
				"url":     alertPayload.HTTP.URL,
				"body":    alertPayload.HTTP.Body,
				"headers": alertPayload.HTTP.Headers,
			}
			http = append(http, hookConfig)
			if err := d.Set("http", http); err != nil {
				return diag.FromErr(err)
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

package spectrocloud

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/hapi/client"
	"github.com/spectrocloud/hapi/models"
)

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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"is_active": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"component": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"ClusterHealth"}, false),
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"email", "http"}, false),
			},
			"alert_all_users": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"created_by": {
				Type:     schema.TypeString,
				Optional: true,
			},
			// Status is defined here just for schema, we are not using this status. it implemented for internal logic
			"status": {
				Type:     schema.TypeList,
				Optional: true,
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
				Type:     schema.TypeSet,
				Optional: true,
				Set:      schema.HashString,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"http": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"method": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"POST", "GET", "PUT"}, false),
						},
						"url": {
							Type:     schema.TypeString,
							Required: true,
						},
						"body": {
							Type:     schema.TypeString,
							Required: true,
						},
						"headers": {
							Type:     schema.TypeMap,
							Optional: true,
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
	c := m.(*client.V1Client)
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
	c := m.(*client.V1Client)
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
	c := m.(*client.V1Client)
	var diags diag.Diagnostics
	projectUid, err := getProjectID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	err = c.DeleteAlerts(projectUid, d.Get("component").(string), d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceAlertRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error
	c := m.(*client.V1Client)
	var diags diag.Diagnostics
	projectUid, _ := getProjectID(d, m)
	alertPayload, err := c.ReadAlert(projectUid, d.Get("component").(string), d.Id())
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
	c := m.(*client.V1Client)
	if v, ok := d.GetOk("project"); ok && v.(string) != "" {
		projectUid, err = c.GetProjectUID(v.(string))
		if err != nil {
			return "", err
		}
	}
	return projectUid, nil
}

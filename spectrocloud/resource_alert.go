package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"
	"time"
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
			"email": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"alert_all_users": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"identifiers": {
							Type:     schema.TypeSet,
							Optional: true,
							Set:      schema.HashString,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
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
							Optional: true,
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
	projectUid := ""
	var diags diag.Diagnostics
	alertObj := toAlert(d)
	if v, ok := d.GetOk("project"); ok && v.(string) != "" { //if project name is set it's a project scope
		projectUid, err = c.GetProjectUID(v.(string))
		if err != nil {
			return diag.FromErr(err)
		}
	}
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
	projectUid := ""
	var diags diag.Diagnostics
	alertObj := toAlert(d)
	if v, ok := d.GetOk("project"); ok && v.(string) != "" { //if project name is set it's a project scope
		projectUid, err = c.GetProjectUID(v.(string))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	uid, err := c.UpdateAlert(alertObj, projectUid, d.Get("component").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(uid)

	return diags
}

func toAlert(d *schema.ResourceData) (alertEntity *models.V1AlertEntity) {
	channels := make([]*models.V1Channel, 0)
	isEmail, isHttp := getAlertTypes(d)
	if isEmail == true {
		emailInfo := d.Get("email").([]interface{})[0].(map[string]interface{})
		emailIDs := make([]string, 0)
		for _, email := range emailInfo["identifiers"].(*schema.Set).List() {
			emailIDs = append(emailIDs, email.(string))
		}
		emailAlert := &models.V1Channel{
			IsActive:      d.Get("is_active").(bool),
			AlertAllUsers: emailInfo["alert_all_users"].(bool),
			Identifiers:   emailIDs,
			Type:          "email",
		}
		channels = append(channels, emailAlert)
	}
	if isHttp == true {
		for _, val := range d.Get("http").([]interface{}) {
			http := val.(map[string]interface{})
			headersMap := make(map[string]string)
			for key, element := range http["headers"].(map[string]interface{}) {
				headersMap[key] = element.(string)
			}
			channelHttp := &models.V1ChannelHTTP{
				Body:    http["body"].(string),
				Headers: headersMap,
				Method:  http["method"].(string),
				URL:     http["url"].(string),
			}
			httpAlert := &models.V1Channel{
				IsActive: d.Get("is_active").(bool),
				Type:     "http",
				HTTP:     channelHttp,
			}
			channels = append(channels, httpAlert)
		}
	}
	alertEntity = &models.V1AlertEntity{
		Channels: channels,
	}
	return alertEntity
}

func resourceAlertDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	projectUid := ""
	var err error
	c := m.(*client.V1Client)
	var diags diag.Diagnostics
	if v, ok := d.GetOk("project"); ok && v.(string) != "" { //if project name is set it's a project scope
		projectUid, err = c.GetProjectUID(v.(string))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	err = c.DeleteAlerts(projectUid, d.Get("component").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceAlertRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	projectUid := ""
	var err error
	c := m.(*client.V1Client)
	var diags diag.Diagnostics
	if v, ok := d.GetOk("project"); ok && v.(string) != "" { //if project name is set it's a project scope
		projectUid, err = c.GetProjectUID(v.(string))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	projectDetails, _ := c.GetProjectByUID(projectUid)
	channels := projectDetails.Spec.Alerts[0].Channels
	_, err = c.ReadAlert(channels, projectUid, d.Get("component").(string))
	return diags
}

func getAlertTypes(d *schema.ResourceData) (hasEmail bool, hasHttp bool) {
	email := false
	http := false
	for range d.Get("email").([]interface{}) {
		email = true
		break
	}
	for range d.Get("http").([]interface{}) {
		http = true
		break
	}
	return email, http
}

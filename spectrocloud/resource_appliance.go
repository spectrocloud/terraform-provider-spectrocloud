package spectrocloud

import (
	"context"
	"log"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/spectrocloud/hapi/client"
	"github.com/spectrocloud/hapi/models"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAppliance() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceApplianceCreate,
		ReadContext:   resourceApplianceRead,
		UpdateContext: resourceApplianceUpdate,
		DeleteContext: resourceApplianceDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 2,
		Schema: map[string]*schema.Schema{
			"uid": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tags": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"pairing_key": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"wait": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

/*
{"metadata":{"name":"test_id","uid":"test_id","tags":{"name":"test_tag"}}}
*/
var resourceApplianceCreatePendingStates = []string{
	"unpaired_", "ready_",
}

func resourceApplianceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	var diags diag.Diagnostics

	appliance := toApplianceEntity(d)
	uid, err := c.CreateAppliance(appliance)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(uid)

	// Wait, catching any errors
	if d.Get("wait") != nil && d.Get("wait").(bool) {
		stateConf := &resource.StateChangeConf{
			Pending:    resourceApplianceCreatePendingStates,
			Target:     []string{"ready_healthy"},
			Refresh:    resourceApplianceStateRefreshFunc(c, d.Id()),
			Timeout:    d.Timeout(schema.TimeoutCreate) - 1*time.Minute,
			MinTimeout: 10 * time.Second,
			Delay:      30 * time.Second,
		}

		_, err = stateConf.WaitForStateContext(ctx)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func resourceApplianceStateRefreshFunc(c *client.V1Client, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		appliance, err := c.GetAppliance(id)
		if err != nil {
			return nil, "", err
		} else if appliance == nil {
			return nil, "Deleted", nil
		}

		state := appliance.Status.State + "_" + appliance.Status.Health.State
		log.Printf("Appliance state (%s): %s", id, state)

		return appliance, state, nil
	}
}

func resourceApplianceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	var diags diag.Diagnostics
	if id, okId := d.GetOk("uid"); okId {
		appliance, err := c.GetAppliance(id.(string))
		if err != nil {
			return diag.FromErr(err)
		} else if appliance == nil {
			d.SetId("")
			return diags
		}
		d.SetId(appliance.Metadata.UID)
		/*err = d.Set("name", appliance.Metadata.Name)
		if err != nil {
			return diag.FromErr(err)
		}*/
	}
	return diags
}

func resourceApplianceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	var diags diag.Diagnostics

	appliance := toAppliance(d)
	err := c.UpdateAppliance(d.Id(), appliance)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceApplianceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	var diags diag.Diagnostics
	err := c.DeleteAppliance(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func toApplianceEntity(d *schema.ResourceData) *models.V1EdgeHostDeviceEntity {
	id := d.Get("uid").(string)
	tags := map[string]string{}
	if d.Get("tags") != nil {
		tags = expandStringMap(d.Get("tags").(map[string]interface{}))
	}

	metadata := &models.V1ObjectTagsEntity{
		UID:    id,
		Name:   id,
		Labels: tags,
	}

	key := ""
	if d.Get("pairing_key") != nil {
		key = d.Get("pairing_key").(string)
	}
	return &models.V1EdgeHostDeviceEntity{
		Metadata: metadata,
		Spec: &models.V1EdgeHostDeviceSpecEntity{
			HostPairingKey: strfmt.Password(key),
		},
	}
}

func toAppliance(d *schema.ResourceData) *models.V1EdgeHostDevice {

	if d.Get("tags") != nil {
		tags := d.Get("tags").(map[string]interface{})

		appliance := SetFields(d, tags)

		return &appliance
	}

	return &models.V1EdgeHostDevice{}

}

func SetFields(d *schema.ResourceData, tags map[string]interface{}) models.V1EdgeHostDevice {
	appliance := models.V1EdgeHostDevice{}
	appliance.Metadata = &models.V1ObjectMeta{}
	appliance.Metadata.UID = d.Id()
	if tags["name"] != nil {
		appliance.Metadata.Name = tags["name"].(string)
	}
	appliance.Metadata.Labels = expandStringMap(tags)
	return appliance
}

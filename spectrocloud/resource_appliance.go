package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAppliance() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceApplianceCreate,
		ReadContext:   resourceApplianceRead,
		UpdateContext: resourceApplianceUpdate,
		DeleteContext: resourceApplianceDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
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
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

/*
{"metadata":{"name":"test_id","uid":"test_id","labels":{"name":"test_tag"}}}
*/
var resourceApplianceCreatePendingStates = []string{
	"unpaired",
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

	stateConf := &resource.StateChangeConf{
		Pending:    resourceApplianceCreatePendingStates,
		Target:     []string{"ready"},
		Refresh:    resourceApplianceStateRefreshFunc(c, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutCreate) - 1*time.Minute,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	// Wait, catching any errors
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(err)
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

		state := appliance.Status.State
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
		}
		d.SetId(appliance.Metadata.UID)
		d.Set("name", appliance.Metadata.Name)
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
	err := c.DeleteRegistry(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func toApplianceEntity(d *schema.ResourceData) *models.V1EdgeHostDeviceEntity {
	id := d.Get("uid").(string)
	labels := map[string]string{}
	if d.Get("labels") != nil {
		for k, val := range d.Get("labels").(map[string]interface{}) {
			labels[k] = val.(string)
		}
	}
	return &models.V1EdgeHostDeviceEntity{
		Metadata: &models.V1ObjectMeta{
			UID:    id,
			Name:   id,
			Labels: labels,
		},
	}
}

func toAppliance(d *schema.ResourceData) *models.V1EdgeHostDevice {
	//description := d.Get("description").(string)
	return &models.V1EdgeHostDevice{
		Metadata: &models.V1ObjectMeta{
			Name: d.Get("name").(string),
		},
	}
}

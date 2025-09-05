package spectrocloud

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/api/apiutil/transport"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAppliance() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceApplianceCreate,
		ReadContext:   resourceApplianceRead,
		UpdateContext: resourceApplianceUpdate,
		DeleteContext: resourceApplianceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceApplianceImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Description: "A resource for creating and managing appliances for Edge Native cluster provisioning.",

		SchemaVersion: 2,
		Schema: map[string]*schema.Schema{
			"uid": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The unique identifier (UID) for the appliance.",
			},
			"tags": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "A set of key-value pairs that can be used to organize and categorize the appliance.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"pairing_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The pairing key used for appliance pairing.",
			},
			"remote_shell": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "disabled",
				ValidateFunc: validation.StringInSlice([]string{"enabled", "disabled"}, false),
				Description:  "Activate remote shell access to troubleshoot edge hosts by initiating an SSH connection from Palette using the configured username and password credentials. https://docs.spectrocloud.com/clusters/edge/cluster-management/remote-shell/",
			},
			"temporary_shell_credentials": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "disabled",
				ValidateFunc: validation.StringInSlice([]string{"enabled", "disabled"}, false),
				Description:  "Enable the creation of a temporary user on the edge host with sudo privileges for SSH access from Palette. These credentials will be embedded in the SSH connection string for auto login, and the temporary user is deleted upon deactivation.",
			},
			"wait": {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
				Description: "If set to `true`, the resource creation will wait for the appliance provisioning process to complete before returning. Defaults to `false`.",
			},
		},
		CustomizeDiff: func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
			remoteShell := d.Get("remote_shell").(string)
			temporaryCreds := d.Get("temporary_shell_credentials").(string)

			if temporaryCreds == "enabled" && remoteShell == "disabled" {
				return fmt.Errorf("temporary_shell_credentials can only be set to 'enabled' when remote_shell is also enabled")
			}

			return nil
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
	c := getV1ClientWithResourceContext(m, "")
	var diags diag.Diagnostics

	appliance := toApplianceEntity(d)
	uid, err := c.CreateAppliance(appliance)

	if err != nil {
		var e *transport.TransportError
		if errors.As(err, &e) && e.Payload.Code == "AlreadyRegisteredEdgeHostDevice" {
			uid = d.Get("uid").(string)
			d.SetId(uid)
		} else {
			return diag.FromErr(err)
		}
	}

	d.SetId(uid)

	// Wait, catching any errors
	if d.Get("wait") != nil && d.Get("wait").(bool) {
		stateConf := &retry.StateChangeConf{
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
	diags = commonApplianceUpdate(ctx, d, c)

	return diags
}

func resourceApplianceStateRefreshFunc(c *client.V1Client, id string) retry.StateRefreshFunc {
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
	c := getV1ClientWithResourceContext(m, "")
	var diags diag.Diagnostics
	if id, okId := d.GetOk("uid"); okId {
		appliance, err := c.GetAppliance(id.(string))
		if err != nil {
			return handleReadError(d, err, diags)
		} else if appliance == nil {
			d.SetId("")
			return diags
		}
		d.SetId(appliance.Metadata.UID)
		if appliance.Spec.TunnelConfig != nil {
			err = d.Set("remote_shell", appliance.Spec.TunnelConfig.RemoteSSH)
			if err != nil {
				return diag.FromErr(err)
			}
			err = d.Set("temporary_shell_credentials", appliance.Spec.TunnelConfig.RemoteSSHTempUser)
			if err != nil {
				return diag.FromErr(err)
			}
		}

		/*err = d.Set("name", appliance.Metadata.Name)
		if err != nil {
			return diag.FromErr(err)
		}*/
	}
	return diags
}

func commonApplianceUpdate(ctx context.Context, d *schema.ResourceData, c *client.V1Client) diag.Diagnostics {

	var diags diag.Diagnostics

	if d.HasChange("tags") {
		applianceMeta := toApplianceMeta(d)
		err := c.UpdateApplianceMeta(d.Id(), applianceMeta)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("remote_shell") || d.HasChange("temporary_shell_credentials") {
		err := c.UpdateEdgeHostTunnelConfig(d.Id(), d.Get("remote_shell").(string), d.Get("temporary_shell_credentials").(string))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func resourceApplianceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "")
	var diags diag.Diagnostics
	commonApplianceUpdate(ctx, d, c)
	return diags
}

func resourceApplianceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "")
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

func toApplianceMeta(d *schema.ResourceData) *models.V1EdgeHostDeviceMetaUpdateEntity {
	if d.Get("tags") != nil {
		return &models.V1EdgeHostDeviceMetaUpdateEntity{
			Metadata: &models.V1ObjectTagsEntity{
				Labels: expandStringMap(d.Get("tags").(map[string]interface{})),
				Name:   d.Id(),
				UID:    d.Id(),
			},
		}
	}
	return &models.V1EdgeHostDeviceMetaUpdateEntity{}
}

func toAppliance(d *schema.ResourceData) *models.V1EdgeHostDevice {

	if d.Get("tags") != nil {
		tags := d.Get("tags").(map[string]interface{})

		appliance := setFields(d, tags)

		return &appliance
	}

	return &models.V1EdgeHostDevice{}

}

func setFields(d *schema.ResourceData, tags map[string]interface{}) models.V1EdgeHostDevice {
	appliance := models.V1EdgeHostDevice{}
	appliance.Metadata = &models.V1ObjectMeta{}
	appliance.Metadata.UID = d.Id()
	if tags["name"] != nil {
		appliance.Metadata.Name = tags["name"].(string)
	}
	appliance.Metadata.Labels = expandStringMap(tags)
	return appliance
}

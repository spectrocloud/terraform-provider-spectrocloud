package spectrocloud

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	schemas "github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/schemas"
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
				Description: "The unique identifier (UID) for the appliance. Note: This field is required and must be unique across all appliances in the tenant.",
			},
			"arch_type": func() *schema.Schema {
				s := schemas.MachinePoolArchTypeSchema()
				s.ForceNew = true
				return s
			}(),
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
				Description:  "Activate remote shell access to troubleshoot edge hosts by initiating an SSH connection from Palette using the configured credentials. See https://docs.spectrocloud.com/clusters/edge/cluster-management/remote-shell/.",
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
		return diag.FromErr(err)
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
	if diags.HasError() {
		return diags
	}
	return resourceApplianceRead(ctx, d, m)
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

	applianceID := d.Id()
	if uid, ok := d.GetOk("uid"); ok && uid.(string) != "" {
		applianceID = uid.(string)
	}
	if applianceID == "" {
		return diags
	}

	appliance, err := c.GetAppliance(applianceID)
	if err != nil {
		return handleReadError(d, err, diags)
	} else if appliance == nil {
		d.SetId("")
		return diags
	}

	d.SetId(appliance.Metadata.UID)
	if err := d.Set("uid", appliance.Metadata.UID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("tags", flattenTagsMap(appliance.Metadata.Labels)); err != nil {
		return diag.FromErr(err)
	}
	remoteShell, temporaryShellCredentials := flattenApplianceTunnelConfig(appliance.Spec)
	if err := d.Set("remote_shell", remoteShell); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("temporary_shell_credentials", temporaryShellCredentials); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("arch_type", flattenApplianceArchType(appliance)); err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func flattenApplianceArchType(appliance *models.V1EdgeHostDevice) string {
	if appliance != nil && appliance.Spec != nil && appliance.Spec.Device != nil &&
		appliance.Spec.Device.ArchType != nil && *appliance.Spec.Device.ArchType != "" {
		return *appliance.Spec.Device.ArchType
	}
	return "amd64"
}

func toApplianceArchType(d *schema.ResourceData) *models.V1ArchType {
	archType := "amd64"
	if v, ok := d.GetOk("arch_type"); ok && v.(string) != "" {
		archType = v.(string)
	}
	arch := models.V1ArchType(archType)
	return &arch
}

func flattenApplianceTunnelConfig(spec *models.V1EdgeHostDeviceSpec) (remoteShell, temporaryShellCredentials string) {
	remoteShell = models.V1SpectroTunnelConfigRemoteSSHDisabled
	temporaryShellCredentials = models.V1SpectroTunnelConfigRemoteSSHTempUserDisabled
	if spec == nil || spec.TunnelConfig == nil {
		return remoteShell, temporaryShellCredentials
	}
	if spec.TunnelConfig.RemoteSSH != nil && *spec.TunnelConfig.RemoteSSH != "" {
		remoteShell = *spec.TunnelConfig.RemoteSSH
	}
	if spec.TunnelConfig.RemoteSSHTempUser != nil && *spec.TunnelConfig.RemoteSSHTempUser != "" {
		temporaryShellCredentials = *spec.TunnelConfig.RemoteSSHTempUser
	}
	return remoteShell, temporaryShellCredentials
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
	diags := commonApplianceUpdate(ctx, d, c)
	if diags.HasError() {
		return diags
	}
	return resourceApplianceRead(ctx, d, m)
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
	tags := expandApplianceTags(d)

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
			ArchType:       toApplianceArchType(d),
		},
	}
}

func toApplianceMeta(d *schema.ResourceData) *models.V1EdgeHostDeviceMetaUpdateEntity {
	if tags, ok := d.GetOk("tags"); ok {
		return &models.V1EdgeHostDeviceMetaUpdateEntity{
			Metadata: &models.V1ObjectTagsEntity{
				Labels: expandApplianceTagsMap(tags.(map[string]interface{})),
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
	appliance.Metadata.Labels = expandApplianceTagsMap(tags)
	return appliance
}

func expandApplianceTags(d *schema.ResourceData) map[string]string {
	if tags, ok := d.GetOk("tags"); ok {
		return expandApplianceTagsMap(tags.(map[string]interface{}))
	}
	return map[string]string{}
}

func expandApplianceTagsMap(configured map[string]interface{}) map[string]string {
	if len(configured) == 0 {
		return map[string]string{}
	}
	tags := make(map[string]string, len(configured))
	for k, v := range configured {
		vStr := v.(string)
		if vStr != "" && vStr != "spectro__tag" {
			tags[k] = vStr
		} else {
			tags[k] = "spectro__tag"
		}
	}
	return tags
}

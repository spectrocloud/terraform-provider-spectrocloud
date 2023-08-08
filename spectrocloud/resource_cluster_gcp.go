package spectrocloud

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/schemas"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/palette-sdk-go/client"
)

func resourceClusterGcp() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterGcpCreate,
		ReadContext:   resourceClusterGcpRead,
		UpdateContext: resourceClusterGcpUpdate,
		DeleteContext: resourceClusterDelete,
		Description:   "Resource for managing GCP clusters in Spectro Cloud through Palette.",

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

		SchemaVersion: 2,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"", "project", "tenant"}, false),
			},
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      schema.HashString,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "A list of tags to be applied to the cluster. Tags must be in the form of `key:value`.",
			},
			"cluster_profile": schemas.ClusterProfileSchema(),
			"apply_setting": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"cloud_account_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cloud_config_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the cloud config used for the cluster. This cloud config must be of type `azure`.",
				Deprecated:  "This field is deprecated and will be removed in the future. Use `cloud_config` instead.",
			},
			"os_patch_on_boot": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether to apply OS patch on boot. Default is `false`.",
			},
			"os_patch_schedule": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateOsPatchSchedule,
				Description:      "Cron schedule for OS patching. This must be in the form of `0 0 * * *`.",
			},
			"os_patch_after": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateOsPatchOnDemandAfter,
				Description:      "Date and time after which to patch cluster `RFC3339: 2006-01-02T15:04:05Z07:00`",
			},
			"kubeconfig": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Kubeconfig for the cluster. This can be used to connect to the cluster using `kubectl`.",
			},
			"cloud_config": {
				Type:     schema.TypeList,
				ForceNew: true,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"network": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"project": {
							Type:     schema.TypeString,
							Required: true,
						},
						"region": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"machine_pool": {
				Type:     schema.TypeSet,
				Required: true,
				Set:      resourceMachinePoolGcpHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"additional_labels": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"taints": schemas.ClusterTaintsSchema(),
						"control_plane": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
							//ForceNew: true,
							Description: "Whether this machine pool is a control plane. Defaults to `false`.",
						},
						"control_plane_as_worker": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
							//ForceNew: true,
							Description: "Whether this machine pool is a control plane and a worker. Defaults to `false`.",
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
							//ForceNew: true,
						},
						"count": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Number of nodes in the machine pool.",
						},
						"node_repave_interval": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     0,
							Description: "Minimum number of seconds node should be Ready, before the next node is selected for repave. Default value is `0`, Applicable only for worker pools.",
						},
						"instance_type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"update_strategy": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "RollingUpdateScaleOut",
							Description:  "Update strategy for the machine pool. Valid values are `RollingUpdateScaleOut` and `RollingUpdateScaleIn`.",
							ValidateFunc: validation.StringInSlice([]string{"RollingUpdateScaleOut", "RollingUpdateScaleIn"}, false),
						},
						"disk_size_gb": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  65,
						},
						"azs": {
							Type:     schema.TypeSet,
							Required: true,
							MinItems: 1,
							Set:      schema.HashString,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"backup_policy":        schemas.BackupPolicySchema(),
			"scan_policy":          schemas.ScanPolicySchema(),
			"cluster_rbac_binding": schemas.ClusterRbacBindingSchema(),
			"namespaces":           schemas.ClusterNamespacesSchema(),
			"host_config":          schemas.ClusterHostConfigSchema(),
			"location_config":      schemas.ClusterLocationSchemaComputed(),
			"skip_completion": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "If `true`, the cluster will be created asynchronously. Default value is `false`.",
			},
		},
	}
}

func resourceClusterGcpCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	cluster, err := toGcpCluster(c, d)
	if err != nil {
		return diag.FromErr(err)
	}

	ClusterContext := d.Get("context").(string)
	uid, err := c.CreateClusterGcp(cluster, ClusterContext)
	if err != nil {
		return diag.FromErr(err)
	}

	diagnostics, isError := waitForClusterCreation(ctx, d, ClusterContext, uid, diags, c, true)
	if isError {
		return diagnostics
	}

	resourceClusterGcpRead(ctx, d, m)

	return diags
}

//goland:noinspection GoUnhandledErrorResult
func resourceClusterGcpRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	var diags diag.Diagnostics

	cluster, err := resourceClusterRead(d, c, diags)
	if err != nil {
		return diag.FromErr(err)
	} else if cluster == nil {
		// Deleted - Terraform will recreate it
		d.SetId("")
		return diags
	}

	diagnostics, done := readCommonFields(c, d, cluster)
	if done {
		return diagnostics
	}

	return flattenCloudConfigGcp(cluster.Spec.CloudConfigRef.UID, d, c)
}

func flattenCloudConfigGcp(configUID string, d *schema.ResourceData, c *client.V1Client) diag.Diagnostics {
	ClusterContext := d.Get("context").(string)
	if err := d.Set("cloud_config_id", configUID); err != nil {
		return diag.FromErr(err)
	}
	if config, err := c.GetCloudConfigGcp(configUID, ClusterContext); err != nil {
		return diag.FromErr(err)
	} else {
		mp := flattenMachinePoolConfigsGcp(config.Spec.MachinePoolConfig)
		if err := d.Set("machine_pool", mp); err != nil {
			return diag.FromErr(err)
		}
	}

	return diag.Diagnostics{}
}

func flattenMachinePoolConfigsGcp(machinePools []*models.V1GcpMachinePoolConfig) []interface{} {

	if machinePools == nil {
		return make([]interface{}, 0)
	}

	ois := make([]interface{}, len(machinePools))

	for i, machinePool := range machinePools {
		oi := make(map[string]interface{})

		FlattenAdditionalLabelsAndTaints(machinePool.AdditionalLabels, machinePool.Taints, oi)
		FlattenControlPlaneAndRepaveInterval(machinePool.IsControlPlane, oi, machinePool.NodeRepaveInterval)

		oi["control_plane_as_worker"] = machinePool.UseControlPlaneAsWorker
		oi["name"] = machinePool.Name
		oi["count"] = int(machinePool.Size)
		flattenUpdateStrategy(machinePool.UpdateStrategy, oi)

		oi["instance_type"] = *machinePool.InstanceType

		oi["disk_size_gb"] = int(machinePool.RootDeviceSize)

		oi["azs"] = machinePool.Azs
		ois[i] = oi
	}

	return ois
}

func resourceClusterGcpUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	cloudConfigId := d.Get("cloud_config_id").(string)
	ClusterContext := d.Get("context").(string)
	if d.HasChange("machine_pool") {
		oraw, nraw := d.GetChange("machine_pool")
		if oraw == nil {
			oraw = new(schema.Set)
		}
		if nraw == nil {
			nraw = new(schema.Set)
		}

		os := oraw.(*schema.Set)
		ns := nraw.(*schema.Set)

		osMap := make(map[string]interface{})
		for _, mp := range os.List() {
			machinePool := mp.(map[string]interface{})
			osMap[machinePool["name"].(string)] = machinePool
		}

		for _, mp := range ns.List() {
			machinePoolResource := mp.(map[string]interface{})
			// since known issue in TF SDK: https://github.com/hashicorp/terraform-plugin-sdk/issues/588
			if machinePoolResource["name"].(string) != "" {
				name := machinePoolResource["name"].(string)
				hash := resourceMachinePoolGcpHash(machinePoolResource)
				var err error
				machinePool, err := toMachinePoolGcp(machinePoolResource)
				if err != nil {
					return diag.FromErr(err)
				}

				if oldMachinePool, ok := osMap[name]; !ok {
					log.Printf("Create machine pool %s", name)
					err = c.CreateMachinePoolGcp(cloudConfigId, ClusterContext, machinePool)
				} else if hash != resourceMachinePoolGcpHash(oldMachinePool) {
					log.Printf("Change in machine pool %s", name)
					err = c.UpdateMachinePoolGcp(cloudConfigId, ClusterContext, machinePool)
				}

				if err != nil {
					return diag.FromErr(err)
				}

				// Processed (if exists)
				delete(osMap, name)
			}
		}

		// Deleted old machine pools
		for _, mp := range osMap {
			machinePool := mp.(map[string]interface{})
			name := machinePool["name"].(string)
			log.Printf("Deleted machine pool %s", name)
			if err := c.DeleteMachinePoolGcp(cloudConfigId, name, ClusterContext); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	diagnostics, done := updateCommonFields(d, c)
	if done {
		return diagnostics
	}

	resourceClusterGcpRead(ctx, d, m)

	return diags
}

func toGcpCluster(c *client.V1Client, d *schema.ResourceData) (*models.V1SpectroGcpClusterEntity, error) {
	// gnarly, I know! =/
	cloudConfig := d.Get("cloud_config").([]interface{})[0].(map[string]interface{})
	//clientSecret := strfmt.Password(d.Get("gcp_client_secret").(string))

	profiles, err := toProfiles(c, d)
	if err != nil {
		return nil, err
	}
	cluster := &models.V1SpectroGcpClusterEntity{
		Metadata: &models.V1ObjectMeta{
			Name:   d.Get("name").(string),
			UID:    d.Id(),
			Labels: toTags(d),
		},
		Spec: &models.V1SpectroGcpClusterEntitySpec{
			CloudAccountUID: types.Ptr(d.Get("cloud_account_id").(string)),
			Profiles:        profiles,
			Policies:        toPolicies(d),
			CloudConfig: &models.V1GcpClusterConfig{
				Network: cloudConfig["network"].(string),
				Project: types.Ptr(cloudConfig["project"].(string)),
				Region:  types.Ptr(cloudConfig["region"].(string)),
			},
		},
	}

	machinePoolConfigs := make([]*models.V1GcpMachinePoolConfigEntity, 0)
	for _, machinePool := range d.Get("machine_pool").(*schema.Set).List() {
		mp, err := toMachinePoolGcp(machinePool)
		if err != nil {
			return nil, err
		}
		machinePoolConfigs = append(machinePoolConfigs, mp)
	}

	cluster.Spec.Machinepoolconfig = machinePoolConfigs
	cluster.Spec.ClusterConfig = toClusterConfig(d)

	return cluster, nil
}

func toMachinePoolGcp(machinePool interface{}) (*models.V1GcpMachinePoolConfigEntity, error) {
	m := machinePool.(map[string]interface{})

	labels := make([]string, 0)
	controlPlane := m["control_plane"].(bool)
	controlPlaneAsWorker := m["control_plane_as_worker"].(bool)
	if controlPlane {
		labels = append(labels, "master")
	}

	azs := make([]string, 0)
	for _, az := range m["azs"].(*schema.Set).List() {
		azs = append(azs, az.(string))
	}

	mp := &models.V1GcpMachinePoolConfigEntity{
		CloudConfig: &models.V1GcpMachinePoolCloudConfigEntity{
			Azs:            azs,
			InstanceType:   types.Ptr(m["instance_type"].(string)),
			RootDeviceSize: int64(m["disk_size_gb"].(int)),
		},
		PoolConfig: &models.V1MachinePoolConfigEntity{
			AdditionalLabels: toAdditionalNodePoolLabels(m),
			Taints:           toClusterTaints(m),
			IsControlPlane:   controlPlane,
			Labels:           labels,
			Name:             types.Ptr(m["name"].(string)),
			Size:             types.Ptr(int32(m["count"].(int))),
			UpdateStrategy: &models.V1UpdateStrategy{
				Type: getUpdateStrategy(m),
			},
			UseControlPlaneAsWorker: controlPlaneAsWorker,
		},
	}

	if !controlPlane {
		nodeRepaveInterval := 0
		if m["node_repave_interval"] != nil {
			nodeRepaveInterval = m["node_repave_interval"].(int)
		}
		mp.PoolConfig.NodeRepaveInterval = int32(nodeRepaveInterval)
	} else {
		err := ValidationNodeRepaveIntervalForControlPlane(m["node_repave_interval"].(int))
		if err != nil {
			return mp, err
		}
	}

	return mp, nil
}

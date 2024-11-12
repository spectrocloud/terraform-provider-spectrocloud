package spectrocloud

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/schemas"
	"github.com/spectrocloud/terraform-provider-spectrocloud/util/ptr"
)

func resourceClusterCustomCloud() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterCustomCloudCreate,
		ReadContext:   resourceClusterCustomCloudRead,
		UpdateContext: resourceClusterCustomCloudUpdate,
		DeleteContext: resourceClusterDelete,
		Importer: &schema.ResourceImporter{

			StateContext: resourceClusterCustomImport,
		},
		Description: "Resource for managing custom cloud clusters in Spectro Cloud through Palette.",

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

		SchemaVersion: 2,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the cluster.",
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"", "project", "tenant"}, false),
				Description: "The context of the EKS cluster. Allowed values are `project` or `tenant`. " +
					"Default is `project`. " + PROJECT_NAME_NUANCE,
			},
			"cloud": {
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
				Description: "The cloud provider name.",
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
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The description of the cluster. Default value is empty string.",
			},
			"cluster_profile": schemas.ClusterProfileSchema(),
			"apply_setting": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "DownloadAndInstall",
				ValidateFunc: validation.StringInSlice([]string{"DownloadAndInstall", "DownloadAndInstallLater"}, false),
				Description: "The setting to apply the cluster profile. `DownloadAndInstall` will download and install packs in one action. " +
					"`DownloadAndInstallLater` will only download artifact and postpone install for later. " +
					"Default value is `DownloadAndInstall`.",
			},
			"cloud_account_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The cloud account id to use for this cluster.",
			},
			"cloud_config_id": {
				Type:       schema.TypeString,
				Computed:   true,
				Deprecated: "This field is deprecated and will be removed in the future. Use `cloud_config` instead.",
			},
			"cloud_config": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Description: "The Cloud environment configuration settings such as network parameters and encryption parameters that apply to this cluster.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"values": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The values of the cloud config. The values are specified in YAML format. ",
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								// UI strips the trailing newline on save
								return strings.TrimSpace(old) == strings.TrimSpace(new)
							},
						},
					},
				},
			},

			"machine_pool": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "The machine pool configuration for the cluster.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the machine pool. This will be derived from the name value in the `node_pool_config`.",
						},
						"count": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of nodes in the machine pool. This will be derived from the replica value in the 'node_pool_config'.",
						},
						"control_plane": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Whether this machine pool is a control plane. Defaults to `false`.",
						},
						"control_plane_as_worker": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Whether this machine pool is a control plane and a worker. Defaults to `false`.",
						},
						"node_pool_config": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The values of the node pool config. The values are specified in YAML format. ",
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								// UI strips the trailing newline on save
								return strings.TrimSpace(old) == strings.TrimSpace(new)
							},
						},
						// Planned for support on future release's - "update_strategy", "node_repave_interval"
					},
				},
			},
			"pause_agent_upgrades": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "unlock",
				ValidateFunc: validation.StringInSlice([]string{"lock", "unlock"}, false),
				Description:  "The pause agent upgrades setting allows to control the automatic upgrade of the Palette component and agent for an individual cluster. The default value is `unlock`, meaning upgrades occur automatically. Setting it to `lock` pauses automatic agent upgrades for the cluster.",
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
				Description:      "The cron schedule for OS patching. This must be in the form of cron syntax. Ex: `0 0 * * *`.",
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
			"admin_kube_config": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Admin Kube-config for the cluster. This can be used to connect to the cluster using `kubectl`, With admin privilege.",
			},
			"backup_policy":        schemas.BackupPolicySchema(),
			"scan_policy":          schemas.ScanPolicySchema(),
			"cluster_rbac_binding": schemas.ClusterRbacBindingSchema(),
			"namespaces":           schemas.ClusterNamespacesSchema(),
			"location_config":      schemas.ClusterLocationSchema(),

			"skip_completion": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "If `true`, the cluster will be created asynchronously. Default value is `false`.",
			},
			"force_delete": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "If set to `true`, the cluster will be force deleted and user has to manually clean up the provisioned cloud resources.",
			},
			"force_delete_delay": {
				Type:             schema.TypeInt,
				Optional:         true,
				Default:          20,
				Description:      "Delay duration in minutes to before invoking cluster force delete. Default and minimum is 20.",
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(20)),
			},
			// Planned for support on future release's - "review_repave_state",
		},
	}
}

func resourceClusterCustomCloudCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	var diags diag.Diagnostics

	cluster, err := toCustomCloudCluster(c, d)
	if err != nil {
		return diag.FromErr(err)
	}
	cloudType := d.Get("cloud").(string)

	err = c.ValidateCustomCloudType(cloudType)
	if err != nil {
		return diag.FromErr(err)
	}

	uid, err := c.CreateClusterCustomCloud(cluster, cloudType)
	if err != nil {
		return diag.FromErr(err)
	}

	diagnostics, isError := waitForClusterCreation(ctx, d, uid, diags, c, true)
	if isError && diagnostics != nil {
		return diagnostics
	}

	resourceClusterCustomCloudRead(ctx, d, m)

	return diags
}

func resourceClusterCustomCloudRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	var diags diag.Diagnostics

	cluster, err := resourceClusterRead(d, c, diags)
	if err != nil {
		return diag.FromErr(err)
	} else if cluster == nil {
		d.SetId("")
		return diags
	}
	diagnostics, hasError := readCommonFields(c, d, cluster)
	if hasError {
		return diagnostics
	}
	diagnostics, hasError = flattenCloudConfigCustom(cluster.Spec.CloudConfigRef.UID, d, c)
	if hasError {
		return diagnostics
	}

	return diags
}

func resourceClusterCustomCloudUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	cloudConfigId := d.Get("cloud_config_id").(string)
	//clusterContext := d.Get("context").(string)
	cloudType := d.Get("cloud").(string)

	_, err := c.GetCloudConfigCustomCloud(cloudConfigId, cloudType)
	if err != nil {
		return diag.FromErr(err)
	}
	if d.HasChange("cloud_config") {
		config := toCustomCloudConfig(d)
		configEntity := &models.V1CustomCloudClusterConfigEntity{
			ClusterConfig: config,
		}
		err = c.UpdateCloudConfigCustomCloud(configEntity, cloudConfigId, cloudType)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("machine_pool") {
		oraw, nraw := d.GetChange("machine_pool")
		if oraw == nil {
			oraw = new(schema.Set)
		}
		if nraw == nil {
			nraw = new(schema.Set)
		}

		os := oraw.([]interface{})
		ns := nraw.([]interface{})

		osMap := make(map[string]interface{})
		for _, mp := range os {
			machinePool := mp.(map[string]interface{})
			osMap[machinePool["name"].(string)] = machinePool
		}

		nsMap := make(map[string]interface{})
		for _, mp := range ns {
			machinePoolResource := mp.(map[string]interface{})
			nsMap[machinePoolResource["name"].(string)] = machinePoolResource
			if machinePoolResource["name"].(string) != "" {
				name := machinePoolResource["name"].(string)
				hash := resourceMachinePoolCustomCloudHash(machinePoolResource)
				var err error
				machinePool := toMachinePoolCustomCloud(mp)
				if oldMachinePool, ok := osMap[name]; !ok {
					log.Printf("Create machine pool %s", name)
					if err = c.CreateMachinePoolCustomCloud(machinePool, cloudConfigId, cloudType); err != nil {
						return diag.FromErr(err)
					}
				} else if hash != resourceMachinePoolCustomCloudHash(oldMachinePool) {
					log.Printf("Change in machine pool %s", name)
					if err = c.UpdateMachinePoolCustomCloud(machinePool, name, cloudConfigId, cloudType); err != nil {
						return diag.FromErr(err)
					}
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
			if err = c.DeleteMachinePoolCustomCloud(name, cloudConfigId, cloudType); err != nil {
				return diag.FromErr(err)
			}
		}

	}

	diagnostics, done := updateCommonFields(d, c)
	if done {
		return diagnostics
	}

	resourceClusterCustomCloudRead(ctx, d, m)

	return diags
}

func toCustomCloudCluster(c *client.V1Client, d *schema.ResourceData) (*models.V1SpectroCustomClusterEntity, error) {

	clusterContext := d.Get("context").(string)
	profiles, err := toProfiles(c, d, clusterContext)
	if err != nil {
		return nil, err
	}

	// policies in not supported for custom cluster during cluster creation UI also its same.
	// policies := toPolicies(d)

	customCloudConfig := toCustomCloudConfig(d)

	customClusterConfig := toCustomClusterConfig(d)

	machinePoolConfigs := make([]*models.V1CustomMachinePoolConfigEntity, 0)
	for _, machinePool := range d.Get("machine_pool").([]interface{}) {
		mp := toMachinePoolCustomCloud(machinePool)
		machinePoolConfigs = append(machinePoolConfigs, mp)
	}

	cluster := &models.V1SpectroCustomClusterEntity{
		Metadata: toClusterMetadataUpdate(d),
		Spec: &models.V1SpectroCustomClusterEntitySpec{
			CloudAccountUID:   ptr.To(d.Get("cloud_account_id").(string)),
			CloudConfig:       customCloudConfig,
			ClusterConfig:     customClusterConfig,
			Machinepoolconfig: machinePoolConfigs,
			Profiles:          profiles,
		},
	}

	return cluster, nil
}

func toCustomCloudConfig(d *schema.ResourceData) *models.V1CustomClusterConfig {
	cloudConfig := d.Get("cloud_config").([]interface{})[0].(map[string]interface{})
	valuesYamlStr := strings.TrimSpace(cloudConfig["values"].(string))
	customCloudConfig := &models.V1CustomClusterConfig{
		Values: ptr.To(valuesYamlStr),
	}

	return customCloudConfig
}

func toCustomClusterConfig(d *schema.ResourceData) *models.V1CustomClusterConfigEntity {
	customClusterConfig := &models.V1CustomClusterConfigEntity{
		Location:                toClusterLocationConfigs(d),
		MachineManagementConfig: toMachineManagementConfig(d),
		Resources:               toClusterResourceConfig(d),
	}

	return customClusterConfig
}

func toMachinePoolCustomCloud(machinePool interface{}) *models.V1CustomMachinePoolConfigEntity {
	mp := &models.V1CustomMachinePoolConfigEntity{}
	node := machinePool.(map[string]interface{})
	controlPlane, _ := node["control_plane"].(bool)
	controlPlaneAsWorker, _ := node["control_plane_as_worker"].(bool)
	mp.CloudConfig = &models.V1CustomMachinePoolCloudConfigEntity{
		Values: node["node_pool_config"].(string),
	}
	mp.PoolConfig = &models.V1CustomMachinePoolBaseConfigEntity{
		IsControlPlane:          controlPlane,
		UseControlPlaneAsWorker: controlPlaneAsWorker,
	}
	return mp
}

func flattenMachinePoolConfigsCustomCloud(machinePools []*models.V1CustomMachinePoolConfig) []interface{} {
	if len(machinePools) == 0 {
		return make([]interface{}, 0)
	}
	mps := make([]interface{}, len(machinePools))

	for i, machinePool := range machinePools {
		mp := make(map[string]interface{})
		mp["control_plane_as_worker"] = machinePool.UseControlPlaneAsWorker
		mp["control_plane"] = machinePool.IsControlPlane
		mp["node_pool_config"] = machinePool.Values
		mp["name"] = machinePool.Name
		mp["count"] = machinePool.Size
		mps[i] = mp
	}

	return mps
}

func flattenCloudConfigCustom(configUID string, d *schema.ResourceData, c *client.V1Client) (diag.Diagnostics, bool) {
	cloudType := d.Get("cloud").(string)
	if err := d.Set("cloud_config_id", configUID); err != nil {
		return diag.FromErr(err), true
	}

	if err := ReadCommonAttributes(d); err != nil {
		return diag.FromErr(err), true
	}
	if config, err := c.GetCloudConfigCustomCloud(configUID, cloudType); err != nil {
		return diag.FromErr(err), true
	} else {
		if config.Spec != nil && config.Spec.CloudAccountRef != nil {
			if err := d.Set("cloud_account_id", config.Spec.CloudAccountRef.UID); err != nil {
				return diag.FromErr(err), true
			}
		}
		if err := d.Set("cloud_config", flattenCloudConfigsValuesCustomCloud(config)); err != nil {
			return diag.FromErr(err), true
		}
		if err := d.Set("machine_pool", flattenMachinePoolConfigsCustomCloud(config.Spec.MachinePoolConfig)); err != nil {
			return diag.FromErr(err), true
		}
	}

	return nil, false
}

func flattenCloudConfigsValuesCustomCloud(config *models.V1CustomCloudConfig) []interface{} {
	if config == nil || config.Spec == nil || config.Spec.ClusterConfig == nil {
		return make([]interface{}, 0)
	}

	m := make(map[string]interface{})

	if ptr.DeRef(config.Spec.ClusterConfig.Values) != "" {
		m["values"] = ptr.DeRef(config.Spec.ClusterConfig.Values)
	}
	return []interface{}{m}
}

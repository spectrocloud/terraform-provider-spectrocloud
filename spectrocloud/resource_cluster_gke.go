package spectrocloud

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/schemas"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

func resourceClusterGke() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterGkeCreate,
		ReadContext:   resourceClusterGkeRead,
		UpdateContext: resourceClusterGkeUpdate,
		DeleteContext: resourceClusterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceClusterGkeImport,
		},
		Description: "Resource for managing GKE clusters through Palette.",

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

		SchemaVersion: 2,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceClusterGkeResourceV1().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceClusterGkeStateUpgradeV1,
				Version: 1,
			},
		},
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
				Description: "The context of the GKE cluster. Allowed values are `project` or `tenant`. " +
					"Default is `project`. " + PROJECT_NAME_NUANCE,
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The description of the cluster. Default value is empty string.",
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
			"cluster_meta_attribute": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "`cluster_meta_attribute` can be used to set additional cluster metadata information, eg `{'nic_name': 'test', 'env': 'stage'}`",
			},
			"cluster_profile":  schemas.ClusterProfileSchema(),
			"cluster_template": schemas.ClusterTemplateSchema(),
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

			"cloud_config": {
				Type:        schema.TypeList,
				ForceNew:    true,
				Required:    true,
				MaxItems:    1,
				Description: "The GKE environment configuration settings such as project parameters and region parameters that apply to this cluster.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"project": {
							Type:        schema.TypeString,
							ForceNew:    true,
							Required:    true,
							Description: "GCP project name.",
						},
						"region": {
							Type:     schema.TypeString,
							ForceNew: true,
							Required: true,
						},
					},
				},
			},
			"update_worker_pool_in_parallel": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"cluster_timezone": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "",
				ValidateFunc: validateTimezone,
				Description:  "Defines the time zone used by this cluster to interpret scheduled operations. Maintenance tasks like upgrades will follow this time zone to ensure they run at the appropriate local time for the cluster. Must be in IANA timezone format (e.g., 'America/New_York', 'Asia/Kolkata', 'Europe/London').",
			},
			"machine_pool": {
				Type:        schema.TypeSet,
				Required:    true,
				Set:         resourceMachinePoolGkeHash,
				Description: "The machine pool configuration for the cluster.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"count": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Number of nodes in the machine pool.",
						},
						"disk_size_gb": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  60,
						},
						"additional_labels": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "Additional labels to be applied to the machine pool. Labels must be in the form of `key:value`.",
						},
						"additional_annotations": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "Additional annotations to be applied to the machine pool. Annotations must be in the form of `key:value`.",
						},
						"instance_type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"update_strategy": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "RollingUpdateScaleOut",
							Description: "Update strategy for the machine pool. Valid values are `RollingUpdateScaleOut` and `RollingUpdateScaleIn`.",
						},
						"override_kubeadm_configuration": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "YAML config for kubeletExtraArgs, preKubeadmCommands, postKubeadmCommands. Overrides pack-level settings. Worker pools only.",
						},
						"node":   schemas.NodeSchema(),
						"taints": schemas.ClusterTaintsSchema(),
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
			"review_repave_state": {
				Type:         schema.TypeString,
				Default:      "",
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"", "Approved", "Pending"}, false),
				Description:  "To authorize the cluster repave, set the value to `Approved` for approval and `\"\"` to decline. Default value is `\"\"`.",
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
			"admin_kube_config": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Admin Kube-config for the cluster. This can be used to connect to the cluster using `kubectl`, With admin privilege.",
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
		},
	}
}

func resourceClusterGkeCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	cluster, err := toGkeCluster(c, d)
	if err != nil {
		return diag.FromErr(err)
	}

	uid, err := c.CreateClusterGke(cluster)
	if err != nil {
		return diag.FromErr(err)
	}

	diagnostics, isError := waitForClusterCreation(ctx, d, uid, diags, c, true)
	if len(diagnostics) > 0 {
		diags = append(diags, diagnostics...)
	}
	if isError {
		return diagnostics
	}
	resourceClusterGkeRead(ctx, d, m)
	return diags
}

func resourceClusterGkeRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	var diags diag.Diagnostics
	cluster, err := resourceClusterRead(d, c, diags)
	if err != nil {
		return handleReadError(d, err, diags)
	} else if cluster == nil {
		// Deleted - Terraform will recreate it
		d.SetId("")
		return diags
	}

	configUID := cluster.Spec.CloudConfigRef.UID
	if err := d.Set("cloud_config_id", configUID); err != nil {
		return diag.FromErr(err)
	}

	// verify cluster type
	err = ValidateCloudType("spectrocloud_cluster_gke", cluster)
	if err != nil {
		return diag.FromErr(err)
	}

	diagnostics, done := readCommonFields(c, d, cluster)
	if done {
		return diagnostics
	}

	// Flatten cluster_template variables using variables API
	if err := flattenClusterTemplateVariables(c, d, d.Id()); err != nil {
		return diag.FromErr(err)
	}

	return flattenCloudConfigGke(cluster.Spec.CloudConfigRef.UID, d, c)
}

func resourceClusterGkeUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	var diags diag.Diagnostics
	err := validateSystemRepaveApproval(d, c)
	if err != nil {
		return diag.FromErr(err)
	}
	cloudConfigId := d.Get("cloud_config_id").(string)

	CloudConfig, err := c.GetCloudConfigGke(cloudConfigId)
	if err != nil {
		return diag.FromErr(err)
	}
	if d.HasChange("machine_pool") {
		log.Printf("[DEBUG] === MACHINE POOL CHANGE DETECTED ===")
		oraw, nraw := d.GetChange("machine_pool")
		if oraw == nil {
			oraw = new(schema.Set)
		}
		if nraw == nil {
			nraw = new(schema.Set)
		}

		os := oraw.(*schema.Set)
		ns := nraw.(*schema.Set)

		log.Printf("[DEBUG] Old machine pools count: %d, New machine pools count: %d", os.Len(), ns.Len())

		// Create maps by machine pool name for proper comparison
		osMap := make(map[string]interface{})
		for _, mp := range os.List() {
			machinePoolResource := mp.(map[string]interface{})
			name := machinePoolResource["name"].(string)
			if name != "" {
				osMap[name] = machinePoolResource
			}
		}

		nsMap := make(map[string]interface{})
		for _, mp := range ns.List() {
			machinePoolResource := mp.(map[string]interface{})
			name := machinePoolResource["name"].(string)
			if name != "" {
				nsMap[name] = machinePoolResource

				// Check if this is a new, updated, or unchanged machine pool
				if oldMachinePool, exists := osMap[name]; !exists {
					// NEW machine pool - CREATE
					log.Printf("[DEBUG] Creating new machine pool %s", name)
					machinePool, err := toMachinePoolGke(machinePoolResource)
					if err != nil {
						return diag.FromErr(err)
					}
					if err := c.CreateMachinePoolGke(cloudConfigId, machinePool); err != nil {
						return diag.FromErr(err)
					}
				} else {
					// EXISTING machine pool - check if hash changed
					oldHash := resourceMachinePoolGkeHash(oldMachinePool)
					newHash := resourceMachinePoolGkeHash(machinePoolResource)

					if oldHash != newHash {
						// MODIFIED machine pool - UPDATE
						log.Printf("[DEBUG] Updating machine pool %s (hash changed: %d -> %d)", name, oldHash, newHash)
						machinePool, err := toMachinePoolGke(machinePoolResource)
						if err != nil {
							return diag.FromErr(err)
						}
						if err := c.UpdateMachinePoolGke(cloudConfigId, machinePool); err != nil {
							return diag.FromErr(err)
						}
						// Node Maintenance Actions
						err = resourceNodeAction(c, ctx, machinePoolResource, c.GetNodeMaintenanceStatusGke, CloudConfig.Kind, cloudConfigId, name)
						if err != nil {
							return diag.FromErr(err)
						}
					} else {
						// UNCHANGED machine pool - no action needed
						log.Printf("[DEBUG] Machine pool %s unchanged (hash: %d)", name, oldHash)
					}
				}

				// Mark as processed
				delete(osMap, name)
			} else {
				log.Printf("[DEBUG] WARNING: Machine pool has empty name!")
			}
		}

		// REMOVED machine pools - DELETE
		for name := range osMap {
			log.Printf("[DEBUG] Deleting removed machine pool %s", name)
			if err := c.DeleteMachinePoolGke(cloudConfigId, name); err != nil {
				return diag.FromErr(err)
			}
		}
	}
	diagnostics, done := updateCommonFields(d, c)
	if done {
		return diagnostics
	}

	resourceClusterGkeRead(ctx, d, m)

	return diags
}

func flattenCloudConfigGke(configUID string, d *schema.ResourceData, c *client.V1Client) diag.Diagnostics {
	var diags diag.Diagnostics
	if err := d.Set("cloud_config_id", configUID); err != nil {
		return diag.FromErr(err)
	}
	if config, err := c.GetCloudConfigGke(configUID); err != nil {
		return diag.FromErr(err)
	} else {
		if err := d.Set("cloud_account_id", config.Spec.CloudAccountRef.UID); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("cloud_config", flattenClusterConfigsGke(config)); err != nil {
			return diag.FromErr(err)
		}
		mp := flattenMachinePoolConfigsGke(config.Spec.MachinePoolConfig)
		mp, err := flattenNodeMaintenanceStatus(c, d, c.GetNodeStatusMapGke, mp, configUID)
		if err != nil {
			return diag.FromErr(err)
		}

		if err := d.Set("machine_pool", mp); err != nil {
			return diag.FromErr(err)
		}
	}

	generalWarningForRepave(&diags)
	return diags
}

func flattenClusterConfigsGke(config *models.V1GcpCloudConfig) []interface{} {
	if config == nil || config.Spec == nil || config.Spec.ClusterConfig == nil {
		return make([]interface{}, 0)
	}
	m := make(map[string]interface{})

	if config.Spec.ClusterConfig.Project != nil {
		m["project"] = config.Spec.ClusterConfig.Project
	}
	if String(config.Spec.ClusterConfig.Region) != "" {
		m["region"] = String(config.Spec.ClusterConfig.Region)
	}
	return []interface{}{m}
}

func flattenMachinePoolConfigsGke(machinePools []*models.V1GcpMachinePoolConfig) []interface{} {
	if machinePools == nil {
		return make([]interface{}, 0)
	}

	ois := make([]interface{}, len(machinePools))

	for i, machinePool := range machinePools {
		oi := make(map[string]interface{})

		FlattenAdditionalLabelsAnnotationsAndTaints(machinePool.AdditionalLabels, machinePool.AdditionalAnnotations, machinePool.Taints, oi)
		oi["name"] = machinePool.Name
		oi["count"] = int(machinePool.Size)
		flattenUpdateStrategy(machinePool.UpdateStrategy, oi)

		// Flatten override_kubeadm_configuration (worker pools only)
		if machinePool.IsControlPlane != nil && !*machinePool.IsControlPlane && machinePool.OverrideKubeadmConfiguration != "" {
			oi["override_kubeadm_configuration"] = machinePool.OverrideKubeadmConfiguration
		}

		oi["instance_type"] = *machinePool.InstanceType

		oi["disk_size_gb"] = int(machinePool.RootDeviceSize)
		ois[i] = oi
	}

	return ois
}

func toGkeCluster(c *client.V1Client, d *schema.ResourceData) (*models.V1SpectroGcpClusterEntity, error) {
	cloudConfig := d.Get("cloud_config").([]interface{})[0].(map[string]interface{})

	clusterContext := d.Get("context").(string)
	profiles, err := toProfiles(c, d, clusterContext)
	if err != nil {
		return nil, err
	}
	cluster := &models.V1SpectroGcpClusterEntity{
		Metadata: getClusterMetadata(d),
		Spec: &models.V1SpectroGcpClusterEntitySpec{
			CloudAccountUID: types.Ptr(d.Get("cloud_account_id").(string)),
			Profiles:        profiles,
			ClusterTemplate: toClusterTemplateReference(d),
			Policies:        toPolicies(d),
			CloudConfig: &models.V1GcpClusterConfig{
				Project: types.Ptr(cloudConfig["project"].(string)),
				Region:  types.Ptr(cloudConfig["region"].(string)),
				ManagedClusterConfig: &models.V1GcpManagedClusterConfig{
					Location: cloudConfig["region"].(string),
				},
			},
		},
	}

	machinePoolConfigs := make([]*models.V1GcpMachinePoolConfigEntity, 0)
	for _, machinePool := range d.Get("machine_pool").(*schema.Set).List() {
		mp, err := toMachinePoolGke(machinePool)
		if err != nil {
			return nil, err
		}
		machinePoolConfigs = append(machinePoolConfigs, mp)
	}
	cluster.Spec.Machinepoolconfig = machinePoolConfigs
	cluster.Spec.ClusterConfig = toClusterConfig(d)
	return cluster, err
}

func toMachinePoolGke(machinePool interface{}) (*models.V1GcpMachinePoolConfigEntity, error) {
	m := machinePool.(map[string]interface{})

	mp := &models.V1GcpMachinePoolConfigEntity{
		CloudConfig: &models.V1GcpMachinePoolCloudConfigEntity{
			InstanceType:   types.Ptr(m["instance_type"].(string)),
			RootDeviceSize: SafeInt64(m["disk_size_gb"].(int)),
		},
		PoolConfig: &models.V1MachinePoolConfigEntity{
			AdditionalLabels:      toAdditionalNodePoolLabels(m),
			AdditionalAnnotations: toAdditionalNodePoolAnnotations(m),
			Taints:                toClusterTaints(m),
			Name:                  types.Ptr(m["name"].(string)),
			Size:                  types.Ptr(SafeInt32(m["count"].(int))),
			UpdateStrategy: &models.V1UpdateStrategy{
				Type: getUpdateStrategy(m),
			},
		},
	}
	if !mp.PoolConfig.IsControlPlane {
		mp.PoolConfig.Labels = []string{"worker"}
		// Handle override_kubeadm_configuration (worker pools only)
		if overrideKubeadm, ok := m["override_kubeadm_configuration"].(string); ok && overrideKubeadm != "" {
			mp.PoolConfig.OverrideKubeadmConfiguration = overrideKubeadm
		}
	}
	return mp, nil
}

func resourceClusterGkeResourceV1() *schema.Resource {
	return &schema.Resource{
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
				Description: "The context of the GKE cluster. Allowed values are `project` or `tenant`. " +
					"Default is `project`. " + PROJECT_NAME_NUANCE,
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The description of the cluster. Default value is empty string.",
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
			"cluster_meta_attribute": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "`cluster_meta_attribute` can be used to set additional cluster metadata information, eg `{'nic_name': 'test', 'env': 'stage'}`",
			},
			"cluster_profile":  schemas.ClusterProfileSchema(),
			"cluster_template": schemas.ClusterTemplateSchema(),
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

			"cloud_config": {
				Type:        schema.TypeList,
				ForceNew:    true,
				Required:    true,
				MaxItems:    1,
				Description: "The GKE environment configuration settings such as project parameters and region parameters that apply to this cluster.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"project": {
							Type:        schema.TypeString,
							ForceNew:    true,
							Required:    true,
							Description: "GCP project name.",
						},
						"region": {
							Type:     schema.TypeString,
							ForceNew: true,
							Required: true,
						},
					},
				},
			},
			"update_worker_pool_in_parallel": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"machine_pool": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "The machine pool configuration for the cluster.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"count": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Number of nodes in the machine pool.",
						},
						"disk_size_gb": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  60,
						},
						"additional_labels": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "Additional labels to be applied to the machine pool. Labels must be in the form of `key:value`.",
						},
						"additional_annotations": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "Additional annotations to be applied to the machine pool. Annotations must be in the form of `key:value`.",
						},
						"instance_type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"update_strategy": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "RollingUpdateScaleOut",
							Description: "Update strategy for the machine pool. Valid values are `RollingUpdateScaleOut` and `RollingUpdateScaleIn`.",
						},
						"override_kubeadm_configuration": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "YAML config for kubeletExtraArgs, preKubeadmCommands, postKubeadmCommands. Overrides pack-level settings. Worker pools only.",
						},
						"node":   schemas.NodeSchema(),
						"taints": schemas.ClusterTaintsSchema(),
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
			"review_repave_state": {
				Type:         schema.TypeString,
				Default:      "",
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"", "Approved", "Pending"}, false),
				Description:  "To authorize the cluster repave, set the value to `Approved` for approval and `\"\"` to decline. Default value is `\"\"`.",
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
			"admin_kube_config": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Admin Kube-config for the cluster. This can be used to connect to the cluster using `kubectl`, With admin privilege.",
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
		},
	}
}

func resourceClusterGkeStateUpgradeV1(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	log.Printf("[DEBUG] Upgrading cluster GKE state from version 1 to 2")

	// Convert machine_pool from TypeList to TypeSet
	// Note: We keep the data as a list in rawState and let Terraform's schema processing
	// convert it to TypeSet during normal resource loading. This avoids JSON serialization
	// issues with schema.Set objects that contain hash functions.
	if machinePoolRaw, exists := rawState["machine_pool"]; exists {
		if machinePoolList, ok := machinePoolRaw.([]interface{}); ok {
			log.Printf("[DEBUG] Keeping machine_pool as list during state upgrade with %d items", len(machinePoolList))

			// Keep the machine pool data as-is (as a list)
			// Terraform will convert it to TypeSet when loading the resource using the schema
			rawState["machine_pool"] = machinePoolList

			log.Printf("[DEBUG] Successfully prepared machine_pool for TypeSet conversion")
		} else {
			log.Printf("[DEBUG] machine_pool is not a list, skipping conversion")
		}
	} else {
		log.Printf("[DEBUG] No machine_pool found in state, skipping conversion")
	}

	return rawState, nil
}

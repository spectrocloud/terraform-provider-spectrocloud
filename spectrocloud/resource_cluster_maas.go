package spectrocloud

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/schemas"
	"github.com/spectrocloud/terraform-provider-spectrocloud/util/ptr"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
)

func resourceClusterMaas() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterMaasCreate,
		ReadContext:   resourceClusterMaasRead,
		UpdateContext: resourceClusterMaasUpdate,
		DeleteContext: resourceClusterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceClusterMaasImport,
		},
		Description: "Resource for managing MAAS clusters in Spectro Cloud through Palette.",

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
				Description: "The context of the MAAS configuration. Allowed values are `project` or `tenant`. " +
					"Default is `project`. " + PROJECT_NAME_NUANCE,
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
			"cluster_meta_attribute": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "`cluster_meta_attribute` can be used to set additional cluster metadata information, eg `{'nic_name': 'test', 'env': 'stage'}`",
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
				Optional:    true,
				ForceNew:    true,
				Description: "ID of the Maas cloud account used for the cluster. This cloud account must be of type `maas`.",
			},
			"cloud_config_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the cloud config used for the cluster. This cloud config must be of type `maas`.",
				Deprecated:  "This field is deprecated and will be removed in the future. Use `cloud_config` instead.",
			},
			"review_repave_state": {
				Type:         schema.TypeString,
				Default:      "",
				Optional:     true,
				ValidateFunc: validateReviewRepaveValue,
				Description:  "To authorize the cluster repave, set the value to `Approved` for approval and `\"\"` to decline. Default value is `\"\"`.",
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
				Description:      "Cron schedule for OS patching. This must be in the form of `0 0 * * *`.",
			},
			"os_patch_after": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateOsPatchOnDemandAfter,
				Description:      "The date and time after which to patch the cluster. Prefix the time value with the respective RFC. Ex: `RFC3339: 2006-01-02T15:04:05Z07:00`",
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
			"cloud_config": {
				Type:     schema.TypeList,
				ForceNew: true,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"domain": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Domain name in which the cluster to be provisioned.",
						},
					},
				},
			},
			"machine_pool": {
				Type:     schema.TypeSet,
				Required: true,
				Set:      resourceMachinePoolMaasHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"additional_labels": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "Additional labels to be applied to the machine pool. Labels must be in the form of `key:value`.",
						},
						"node":   schemas.NodeSchema(),
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
							Description: "Name of the machine pool.",
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
						"min": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Minimum number of nodes in the machine pool. This is used for autoscaling the machine pool.",
						},
						"max": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Maximum number of nodes in the machine pool. This is used for autoscaling the machine pool.",
						},
						"instance_type": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"min_memory_mb": {
										Type:        schema.TypeInt,
										Required:    true,
										Description: "Minimum memory in MB required for the machine pool node.",
									},
									"min_cpu": {
										Type:        schema.TypeInt,
										Required:    true,
										Description: "Minimum number of CPU required for the machine pool node.",
									},
								},
							},
						},
						"update_strategy": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "RollingUpdateScaleOut",
							Description:  "Update strategy for the machine pool. Valid values are `RollingUpdateScaleOut` and `RollingUpdateScaleIn`.",
							ValidateFunc: validation.StringInSlice([]string{"RollingUpdateScaleOut", "RollingUpdateScaleIn"}, false),
						},
						"azs": {
							Type:     schema.TypeSet,
							Optional: true,
							Set:      schema.HashString,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "Availability zones in which the machine pool nodes to be provisioned.",
						},
						"node_tags": {
							Type:     schema.TypeSet,
							Optional: true,
							Set:      schema.HashString,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "Node tags to dynamically place nodes in a pool by using MAAS automatic tags. Specify the tag values that you want to apply to all nodes in the node pool.",
						},
						"placement": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "This is a computed(read-only) ID of the placement that is used to connect to the Maas cloud.",
									},
									"resource_pool": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The name of the resource pool in the Maas cloud.",
									},
								},
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
		},
	}
}

func resourceClusterMaasCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	cluster, err := toMaasCluster(c, d)
	if err != nil {
		return diag.FromErr(err)
	}

	uid, err := c.CreateClusterMaas(cluster)
	if err != nil {
		return diag.FromErr(err)
	}

	diagnostics, isError := waitForClusterCreation(ctx, d, uid, diags, c, true)
	if isError {
		return diagnostics
	}

	resourceClusterMaasRead(ctx, d, m)

	return diags
}

func resourceClusterMaasRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	var diags diag.Diagnostics

	cluster, err := resourceClusterRead(d, c, diags)
	if err != nil {
		return diag.FromErr(err)
	} else if cluster == nil {
		// Deleted - Terraform will recreate it
		d.SetId("")
		return diags
	}

	// verify cluster type
	err = ValidateCloudType("spectrocloud_cluster_maas", cluster)
	if err != nil {
		return diag.FromErr(err)
	}

	diagnostics, done := readCommonFields(c, d, cluster)
	if done {
		return diagnostics
	}

	return flattenCloudConfigMaas(cluster.Spec.CloudConfigRef.UID, d, c)
}

func flattenCloudConfigMaas(configUID string, d *schema.ResourceData, c *client.V1Client) diag.Diagnostics {
	var diags diag.Diagnostics
	err := d.Set("cloud_config_id", configUID)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := ReadCommonAttributes(d); err != nil {
		return diag.FromErr(err)
	}

	if config, err := c.GetCloudConfigMaas(configUID); err != nil {
		return diag.FromErr(err)
	} else {
		if config.Spec != nil && config.Spec.CloudAccountRef != nil {
			if err := d.Set("cloud_account_id", config.Spec.CloudAccountRef.UID); err != nil {
				return diag.FromErr(err)
			}
		}
		if err := d.Set("cloud_config", flattenClusterConfigsMaas(config)); err != nil {
			return diag.FromErr(err)
		}
		mp := flattenMachinePoolConfigsMaas(config.Spec.MachinePoolConfig, config.Spec.ClusterConfig)
		mp, err := flattenNodeMaintenanceStatus(c, d, c.GetNodeStatusMapMaas, mp, configUID)
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

func flattenClusterConfigsMaas(config *models.V1MaasCloudConfig) []interface{} {
	if config == nil || config.Spec == nil || config.Spec.ClusterConfig == nil {
		return make([]interface{}, 0)
	}

	m := make(map[string]interface{})

	if config.Spec.ClusterConfig.Domain != nil {
		m["domain"] = *config.Spec.ClusterConfig.Domain
	}

	return []interface{}{m}
}

func flattenMachinePoolConfigsMaas(machinePools []*models.V1MaasMachinePoolConfig, config *models.V1MaasClusterConfig) []interface{} {

	if machinePools == nil {
		return make([]interface{}, 0)
	}

	ois := make([]interface{}, len(machinePools))

	for i, machinePool := range machinePools {
		oi := make(map[string]interface{})

		FlattenAdditionalLabelsAndTaints(machinePool.AdditionalLabels, machinePool.Taints, oi)
		FlattenControlPlaneAndRepaveInterval(&machinePool.IsControlPlane, oi, machinePool.NodeRepaveInterval)

		oi["control_plane_as_worker"] = machinePool.UseControlPlaneAsWorker
		oi["name"] = machinePool.Name
		oi["count"] = int(machinePool.Size)
		flattenUpdateStrategy(machinePool.UpdateStrategy, oi)

		oi["min"] = int(machinePool.MinSize)
		oi["max"] = int(machinePool.MaxSize)
		oi["instance_type"] = machinePool.InstanceType

		if machinePool.InstanceType != nil {
			s := make(map[string]interface{})
			s["min_memory_mb"] = int(machinePool.InstanceType.MinMemInMB)
			s["min_cpu"] = int(machinePool.InstanceType.MinCPU)
			oi["instance_type"] = []interface{}{s}
		}

		oi["azs"] = machinePool.Azs
		if config != nil {
			placement := make(map[string]interface{})
			if config.Domain != nil {
				placement["resource_pool"] = *config.Domain
				oi["placement"] = []interface{}{placement}
			}
		}
		oi["node_tags"] = machinePool.Tags
		ois[i] = oi
	}

	return ois
}

func resourceClusterMaasUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	err := validateSystemRepaveApproval(d, c)
	if err != nil {
		return diag.FromErr(err)
	}

	cloudConfigId := d.Get("cloud_config_id").(string)

	CloudConfig, err := c.GetCloudConfigMaas(cloudConfigId)
	if err != nil {
		return diag.FromErr(err)
	}
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

		nsMap := make(map[string]interface{})

		for _, mp := range ns.List() {
			machinePoolResource := mp.(map[string]interface{})
			nsMap[machinePoolResource["name"].(string)] = machinePoolResource
			// since known issue in TF SDK: https://github.com/hashicorp/terraform-plugin-sdk/issues/588
			if machinePoolResource["name"].(string) != "" {
				name := machinePoolResource["name"].(string)
				hash := resourceMachinePoolMaasHash(machinePoolResource)

				var err error
				machinePool, err := toMachinePoolMaas(machinePoolResource)
				if err != nil {
					return diag.FromErr(err)
				}

				if oldMachinePool, ok := osMap[name]; !ok {
					log.Printf("Create machine pool %s", name)
					err = c.CreateMachinePoolMaas(cloudConfigId, machinePool)
				} else if hash != resourceMachinePoolMaasHash(oldMachinePool) {
					log.Printf("Change in machine pool %s", name)
					err = c.UpdateMachinePoolMaas(cloudConfigId, machinePool)
					// Node Maintenance Actions
					err := resourceNodeAction(c, ctx, nsMap[name], c.GetNodeMaintenanceStatusMaas, CloudConfig.Kind, cloudConfigId, name)
					if err != nil {
						return diag.FromErr(err)
					}
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
			if err := c.DeleteMachinePoolMaas(cloudConfigId, name); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	diagnostics, done := updateCommonFields(d, c)
	if done {
		return diagnostics
	}

	resourceClusterMaasRead(ctx, d, m)

	return diags
}

func toMaasCluster(c *client.V1Client, d *schema.ResourceData) (*models.V1SpectroMaasClusterEntity, error) {
	// gnarly, I know! =/
	cloudConfig := d.Get("cloud_config").([]interface{})[0].(map[string]interface{})
	DomainVal := cloudConfig["domain"].(string)

	clusterContext := d.Get("context").(string)
	profiles, err := toProfiles(c, d, clusterContext)
	if err != nil {
		return nil, err
	}
	cluster := &models.V1SpectroMaasClusterEntity{
		Metadata: getClusterMetadata(d),
		Spec: &models.V1SpectroMaasClusterEntitySpec{
			CloudAccountUID: ptr.To(d.Get("cloud_account_id").(string)),
			Profiles:        profiles,
			Policies:        toPolicies(d),
			CloudConfig: &models.V1MaasClusterConfig{
				Domain: &DomainVal,
			},
		},
	}

	//for _, machinePool := range d.Get("machine_pool").([]interface{}) {
	machinePoolConfigs := make([]*models.V1MaasMachinePoolConfigEntity, 0)
	for _, machinePool := range d.Get("machine_pool").(*schema.Set).List() {
		mp, err := toMachinePoolMaas(machinePool)
		if err != nil {
			return nil, err
		}
		machinePoolConfigs = append(machinePoolConfigs, mp)
	}

	cluster.Spec.Machinepoolconfig = machinePoolConfigs
	cluster.Spec.ClusterConfig = toClusterConfig(d)

	return cluster, nil
}

func toMachinePoolMaas(machinePool interface{}) (*models.V1MaasMachinePoolConfigEntity, error) {
	m := machinePool.(map[string]interface{})

	labels := make([]string, 0)
	controlPlane := m["control_plane"].(bool)
	controlPlaneAsWorker := m["control_plane_as_worker"].(bool)
	if controlPlane {
		labels = append(labels, "control-plane")
	} else {
		labels = append(labels, "worker")
	}

	azs := make([]string, 0)
	for _, az := range m["azs"].(*schema.Set).List() {
		azs = append(azs, az.(string))
	}

	InstanceType := m["instance_type"].([]interface{})[0].(map[string]interface{})
	log.Printf("Create machine pool %s", InstanceType)

	min := int32(m["count"].(int))
	max := int32(m["count"].(int))

	if m["min"] != nil {
		min = int32(m["min"].(int))
	}

	if m["max"] != nil {
		max = int32(m["max"].(int))
	}
	var nodePoolTags []string
	for _, nt := range m["node_tags"].(*schema.Set).List() {
		nodePoolTags = append(nodePoolTags, nt.(string))
	}

	mp := &models.V1MaasMachinePoolConfigEntity{
		CloudConfig: &models.V1MaasMachinePoolCloudConfigEntity{
			Azs: azs,
			InstanceType: &models.V1MaasInstanceType{
				MinCPU:     int32(InstanceType["min_cpu"].(int)),
				MinMemInMB: int32(InstanceType["min_memory_mb"].(int)),
			},
			Tags: nodePoolTags,
		},
		PoolConfig: &models.V1MachinePoolConfigEntity{
			AdditionalLabels: toAdditionalNodePoolLabels(m),
			Taints:           toClusterTaints(m),
			IsControlPlane:   controlPlane,
			Labels:           labels,
			Name:             ptr.To(m["name"].(string)),
			Size:             ptr.To(int32(m["count"].(int))),
			UpdateStrategy: &models.V1UpdateStrategy{
				Type: getUpdateStrategy(m),
			},
			UseControlPlaneAsWorker: controlPlaneAsWorker,
			MinSize:                 min,
			MaxSize:                 max,
		},
	}
	if len(m["placement"].([]interface{})) > 0 {
		Placement := m["placement"].([]interface{})[0].(map[string]interface{})
		mp.CloudConfig.ResourcePool = ptr.To(Placement["resource_pool"].(string))
	} else {
		rp := ""
		mp.CloudConfig.ResourcePool = &rp // backend is not accepting nil, rather pointer to empty string.
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

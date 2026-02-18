package spectrocloud

import (
	"context"
	"log"
	"sort"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/schemas"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
)

func resourceClusterOpenStack() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterOpenStackCreate,
		ReadContext:   resourceClusterOpenStackRead,
		UpdateContext: resourceClusterOpenStackUpdate,
		DeleteContext: resourceClusterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceClusterOpenstackImport,
		},
		Description: "Resource for managing Openstack clusters in Spectro Cloud through Palette.",

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(180 * time.Minute),
			Update: schema.DefaultTimeout(180 * time.Minute),
			Delete: schema.DefaultTimeout(180 * time.Minute),
		},
		SchemaVersion: 2,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceClusterOpenStackResourceV1().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceClusterOpenStackStateUpgradeV1,
				Version: 0,
			},
			{
				Type:    resourceClusterOpenStackResourceV1().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceClusterOpenStackStateUpgradeV2,
				Version: 1,
			},
		},

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
				Description: "The context of the OpenStack cluster. Allowed values are `project` or `tenant`. " +
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
			"cluster_profile":  schemas.ClusterProfileSchemaV2(),
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
			"cluster_timezone": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "",
				ValidateFunc: validateTimezone,
				Description:  "Defines the time zone used by this cluster to interpret scheduled operations. Maintenance tasks like upgrades will follow this time zone to ensure they run at the appropriate local time for the cluster. Must be in IANA timezone format (e.g., 'America/New_York', 'Asia/Kolkata', 'Europe/London').",
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
							Type:     schema.TypeString,
							Required: true,
						},
						"region": {
							Type:     schema.TypeString,
							Required: true,
						},
						"project": {
							Type:     schema.TypeString,
							Required: true,
						},
						"ssh_key": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Public SSH key to be used for the cluster nodes.",
						},
						"network_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"subnet_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"dns_servers": {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							Set:      schema.HashString,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"subnet_cidr": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"machine_pool": {
				Type:     schema.TypeSet,
				Required: true,
				Set:      resourceMachinePoolOpenStackHash,
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
						"additional_annotations": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "Additional annotations to be applied to the machine pool. Annotations must be in the form of `key:value`.",
						},
						"taints": schemas.ClusterTaintsSchema(),
						"node":   schemas.NodeSchema(),
						"control_plane": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
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
						"update_strategy": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "RollingUpdateScaleOut",
							Description:  "Update strategy for the machine pool. Valid values are `RollingUpdateScaleOut`, `RollingUpdateScaleIn` and `OverrideScaling`. If `OverrideScaling` is used, `override_scaling` must be specified with both `max_surge` and `max_unavailable`.",
							ValidateFunc: validation.StringInSlice([]string{"RollingUpdateScaleOut", "RollingUpdateScaleIn", "OverrideScaling"}, false),
						},
						"override_scaling": schemas.OverrideScalingSchema(),
						"override_kubeadm_configuration": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "YAML config for kubeletExtraArgs, preKubeadmCommands, postKubeadmCommands. Overrides pack-level settings. Worker pools only.",
						},
						"instance_type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"azs": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"subnet_id": {
							Type:     schema.TypeString,
							Optional: true,
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

func resourceClusterOpenStackStateUpgradeV2(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	log.Printf("[DEBUG] Upgrading OpenStack cluster state from version 1 to 2")

	// Convert cluster_profile from TypeList (v1) to TypeSet (v2).
	//
	// Note: We keep the data as a list in rawState and let Terraform's schema processing
	// convert it to TypeSet during normal resource loading. This avoids JSON serialization
	// issues with schema.Set objects that contain hash functions.
	if clusterProfileRaw, exists := rawState["cluster_profile"]; exists {
		if clusterProfileList, ok := clusterProfileRaw.([]interface{}); ok {
			log.Printf("[DEBUG] Keeping cluster_profile as list during state upgrade with %d items", len(clusterProfileList))
			rawState["cluster_profile"] = clusterProfileList
			log.Printf("[DEBUG] Successfully prepared cluster_profile for TypeSet conversion")
		} else {
			log.Printf("[DEBUG] cluster_profile is not a list, skipping conversion")
		}
	} else {
		log.Printf("[DEBUG] No cluster_profile found in state, skipping conversion")
	}

	return rawState, nil
}

func resourceClusterOpenStackCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// Validate override_Scaling configuration
	if err := validateOverrideScaling(d, "machine_pool"); err != nil {
		return diag.FromErr(err)
	}

	cluster, err := toOpenStackCluster(c, d)
	if err != nil {
		return diag.FromErr(err)
	}

	uid, err := c.CreateClusterOpenStack(cluster)
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
	resourceClusterOpenStackRead(ctx, d, m)

	return diags
}

func toOpenStackCluster(c *client.V1Client, d *schema.ResourceData) (*models.V1SpectroOpenStackClusterEntity, error) {
	cloudConfig := d.Get("cloud_config").([]interface{})[0].(map[string]interface{})

	clusterContext := d.Get("context").(string)
	profiles, err := toProfiles(c, d, clusterContext)
	if err != nil {
		return nil, err
	}
	cluster := &models.V1SpectroOpenStackClusterEntity{
		Metadata: getClusterMetadata(d),
		Spec: &models.V1SpectroOpenStackClusterEntitySpec{
			CloudAccountUID: types.Ptr(d.Get("cloud_account_id").(string)),
			Profiles:        profiles,
			ClusterTemplate: toClusterTemplateReference(d),
			Policies:        toPolicies(d),
			CloudConfig: &models.V1OpenStackClusterConfig{
				Region:     cloudConfig["region"].(string),
				SSHKeyName: cloudConfig["ssh_key"].(string),
				Domain: &models.V1OpenStackResource{
					ID:   cloudConfig["domain"].(string),
					Name: cloudConfig["domain"].(string),
				},
				Network: &models.V1OpenStackResource{
					ID: cloudConfig["network_id"].(string),
				},
				Project: &models.V1OpenStackResource{
					Name: cloudConfig["project"].(string),
				},
				Subnet: &models.V1OpenStackResource{
					ID: cloudConfig["subnet_id"].(string),
				},
				NodeCidr: cloudConfig["subnet_cidr"].(string),
			},
		},
	}

	if cloudConfig["dns_servers"] != nil {
		dnsServers := make([]string, 0)
		for _, dns := range cloudConfig["dns_servers"].(*schema.Set).List() {
			dnsServers = append(dnsServers, dns.(string))
		}

		cluster.Spec.CloudConfig.DNSNameservers = dnsServers
	}

	machinePoolConfigs := make([]*models.V1OpenStackMachinePoolConfigEntity, 0)

	for _, machinePool := range d.Get("machine_pool").(*schema.Set).List() {
		mp, err := toMachinePoolOpenStack(machinePool)
		if err != nil {
			return nil, err
		}
		machinePoolConfigs = append(machinePoolConfigs, mp)
	}

	// sort
	sort.SliceStable(machinePoolConfigs, func(i, j int) bool {
		return machinePoolConfigs[i].PoolConfig.IsControlPlane
	})

	cluster.Spec.Machinepoolconfig = machinePoolConfigs
	cluster.Spec.ClusterConfig = toClusterConfig(d)

	return cluster, nil
}

//goland:noinspection GoUnhandledErrorResult
func resourceClusterOpenStackRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	err = ValidateCloudType("spectrocloud_cluster_openstack", cluster)
	if err != nil {
		return diag.FromErr(err)
	}

	configUID := cluster.Spec.CloudConfigRef.UID
	if err := d.Set("cloud_config_id", configUID); err != nil {
		return diag.FromErr(err)
	}

	if config, err := c.GetCloudConfigOpenStack(configUID); err != nil {
		return diag.FromErr(err)
	} else {
		if config.Spec != nil && config.Spec.CloudAccountRef != nil {
			if err := d.Set("cloud_account_id", config.Spec.CloudAccountRef.UID); err != nil {
				return diag.FromErr(err)
			}
		}
		if err := d.Set("cloud_config", flattenClusterConfigsOpenstack(config)); err != nil {
			return diag.FromErr(err)
		}

		mp := flattenMachinePoolConfigsOpenStack(config.Spec.MachinePoolConfig)
		mp, err := flattenNodeMaintenanceStatus(c, d, c.GetNodeStatusMapOpenStack, mp, configUID)
		if err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("machine_pool", mp); err != nil {
			return diag.FromErr(err)
		}
	}

	diagnostics, done := readCommonFields(c, d, cluster)
	if done {
		return diagnostics
	}

	// Flatten cluster_template variables using variables API
	if err := flattenClusterTemplateVariables(c, d, d.Id()); err != nil {
		return diag.FromErr(err)
	}

	generalWarningForRepave(&diags)
	return diags
}

func flattenClusterConfigsOpenstack(config *models.V1OpenStackCloudConfig) []interface{} {
	if config == nil || config.Spec == nil || config.Spec.ClusterConfig == nil {
		return make([]interface{}, 0)
	}

	cloudConfig := make(map[string]interface{})

	if config.Spec.ClusterConfig.Domain != nil {
		cloudConfig["domain"] = config.Spec.ClusterConfig.Domain.Name
	}
	if config.Spec.ClusterConfig.Region != "" {
		cloudConfig["region"] = config.Spec.ClusterConfig.Region
	}
	if config.Spec.ClusterConfig.Project != nil {
		cloudConfig["project"] = config.Spec.ClusterConfig.Project.Name
	}
	if config.Spec.ClusterConfig.SSHKeyName != "" {
		cloudConfig["ssh_key"] = config.Spec.ClusterConfig.SSHKeyName
	}
	if config.Spec.ClusterConfig.Network != nil {
		cloudConfig["network_id"] = config.Spec.ClusterConfig.Network.ID
	}
	if config.Spec.ClusterConfig.Subnet != nil {
		cloudConfig["subnet_id"] = config.Spec.ClusterConfig.Subnet.ID
	}
	if config.Spec.ClusterConfig.DNSNameservers != nil {
		cloudConfig["dns_servers"] = config.Spec.ClusterConfig.DNSNameservers
	}
	if config.Spec.ClusterConfig.NodeCidr != "" {
		cloudConfig["subnet_cidr"] = config.Spec.ClusterConfig.NodeCidr
	}

	return []interface{}{cloudConfig}
}

func flattenMachinePoolConfigsOpenStack(machinePools []*models.V1OpenStackMachinePoolConfig) []interface{} {
	if machinePools == nil {
		return make([]interface{}, 0)
	}

	ois := make([]interface{}, 0)

	for _, machinePool := range machinePools {
		oi := make(map[string]interface{})

		FlattenAdditionalLabelsAnnotationsAndTaints(machinePool.AdditionalLabels, machinePool.AdditionalAnnotations, machinePool.Taints, oi)
		FlattenControlPlaneAndRepaveInterval(&machinePool.IsControlPlane, oi, machinePool.NodeRepaveInterval)

		oi["control_plane_as_worker"] = machinePool.UseControlPlaneAsWorker
		oi["name"] = machinePool.Name
		oi["count"] = int(machinePool.Size)
		if machinePool.UpdateStrategy != nil && machinePool.UpdateStrategy.Type != "" {
			oi["update_strategy"] = machinePool.UpdateStrategy.Type
			// Flatten override_Scaling if using OverrideScaling strategy
			flattenOverrideScaling(machinePool.UpdateStrategy, oi)
		} else {
			oi["update_strategy"] = "RollingUpdateScaleOut"
		}

		// Flatten override_kubeadm_configuration (worker pools only)
		if !machinePool.IsControlPlane && machinePool.OverrideKubeadmConfiguration != "" {
			oi["override_kubeadm_configuration"] = machinePool.OverrideKubeadmConfiguration
		}

		oi["subnet_id"] = machinePool.Subnet.ID
		oi["azs"] = machinePool.Azs
		oi["instance_type"] = machinePool.FlavorConfig.Name

		ois = append(ois, oi)
	}

	return ois
}

func resourceClusterOpenStackUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	err := validateSystemRepaveApproval(d, c)
	if err != nil {
		return diag.FromErr(err)
	}
	cloudConfigId := d.Get("cloud_config_id").(string)
	CloudConfig, err := c.GetCloudConfigOpenStack(cloudConfigId)
	if err != nil {
		return diag.FromErr(err)
	}
	if d.HasChange("machine_pool") {
		// Validate override_Scaling configuration
		if err := validateOverrideScaling(d, "machine_pool"); err != nil {
			return diag.FromErr(err)
		}

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
				if oldMachinePool, exists := osMap[name]; !exists {
					log.Printf("[DEBUG] Create machine pool %s", name)
					machinePool, err := toMachinePoolOpenStack(machinePoolResource)
					if err != nil {
						return diag.FromErr(err)
					}
					if err := c.CreateMachinePoolOpenStack(cloudConfigId, machinePool); err != nil {
						return diag.FromErr(err)
					}
				} else {
					oldHash := resourceMachinePoolOpenStackHash(oldMachinePool)
					newHash := resourceMachinePoolOpenStackHash(machinePoolResource)

					if oldHash != newHash {
						log.Printf("[DEBUG] Updating machine pool %s (hash changed: %d -> %d)", name, oldHash, newHash)
						machinePool, err := toMachinePoolOpenStack(machinePoolResource)
						if err != nil {
							return diag.FromErr(err)
						}
						if err := c.UpdateMachinePoolOpenStack(cloudConfigId, machinePool); err != nil {
							return diag.FromErr(err)
						}
						err = resourceNodeAction(c, ctx, machinePoolResource, c.GetNodeMaintenanceStatusOpenStack, CloudConfig.Kind, cloudConfigId, name)
						if err != nil {
							return diag.FromErr(err)
						}
					} else {
						log.Printf("[DEBUG] Machine pool %s unchanged (hash: %d)", name, oldHash)
					}
				}
				delete(osMap, name)
			}
		}

		// REMOVED machine pools - DELETE
		for name := range osMap {
			log.Printf("[DEBUG] Deleting removed machine pool %s", name)
			if err := c.DeleteMachinePoolOpenStack(cloudConfigId, name); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	diagnostics, done := updateCommonFields(d, c)
	if done {
		return diagnostics
	}

	resourceClusterOpenStackRead(ctx, d, m)

	return diags
}

func toMachinePoolOpenStack(machinePool interface{}) (*models.V1OpenStackMachinePoolConfigEntity, error) {
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
	if _, ok := m["azs"]; ok && m["azs"] != nil {
		azsSet := m["azs"].(*schema.Set)
		for _, val := range azsSet.List() {
			azs = append(azs, val.(string))
		}
	}

	mp := &models.V1OpenStackMachinePoolConfigEntity{
		CloudConfig: &models.V1OpenStackMachinePoolCloudConfigEntity{
			Azs: azs,
			Subnet: &models.V1OpenStackResource{
				ID: m["subnet_id"].(string),
			},
			FlavorConfig: &models.V1OpenstackFlavorConfig{
				Name: types.Ptr(m["instance_type"].(string)),
			},
		},
		PoolConfig: &models.V1MachinePoolConfigEntity{
			AdditionalLabels:        toAdditionalNodePoolLabels(m),
			AdditionalAnnotations:   toAdditionalNodePoolAnnotations(m),
			Taints:                  toClusterTaints(m),
			IsControlPlane:          controlPlane,
			Labels:                  labels,
			Name:                    types.Ptr(m["name"].(string)),
			Size:                    types.Ptr(SafeInt32(m["count"].(int))),
			UpdateStrategy:          toUpdateStrategy(m),
			UseControlPlaneAsWorker: controlPlaneAsWorker,
		},
	}

	// Handle override_kubeadm_configuration (worker pools only)
	if !controlPlane {
		if overrideKubeadm, ok := m["override_kubeadm_configuration"].(string); ok && overrideKubeadm != "" {
			mp.PoolConfig.OverrideKubeadmConfiguration = overrideKubeadm
		}
	}

	if !controlPlane {
		nodeRepaveInterval := 0
		if m["node_repave_interval"] != nil {
			nodeRepaveInterval = m["node_repave_interval"].(int)
		}
		mp.PoolConfig.NodeRepaveInterval = SafeInt32(nodeRepaveInterval)
	} else {
		err := ValidationNodeRepaveIntervalForControlPlane(m["node_repave_interval"].(int))
		if err != nil {
			return mp, err
		}
	}

	return mp, nil
}

func resourceClusterOpenStackStateUpgradeV1(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	log.Printf("[DEBUG] Upgrading cluster OpenStack state from version 2 to 3")

	// Convert machine_pool from TypeList to TypeSet
	// Note: We keep the data as a list in rawState and let Terraform's schema processing
	// convert it to TypeSet when loading the resource using the schema. This avoids JSON serialization
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

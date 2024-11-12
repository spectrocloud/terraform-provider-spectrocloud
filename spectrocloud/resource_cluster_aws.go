package spectrocloud

import (
	"context"
	"log"
	"sort"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/schemas"
	"github.com/spectrocloud/terraform-provider-spectrocloud/util/ptr"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
)

func resourceClusterAws() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterAwsCreate,
		ReadContext:   resourceClusterAwsRead,
		UpdateContext: resourceClusterAwsUpdate,
		DeleteContext: resourceClusterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceClusterAwsImport,
		},
		Description: "Resource for managing AWS clusters in Spectro Cloud through Palette.",

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
				Description: "The context of the AWS cluster. Allowed values are `project` or `tenant`. " +
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
			"cloud_config": {
				Type:     schema.TypeList,
				ForceNew: true,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ssh_key_name": {
							Type:        schema.TypeString,
							ForceNew:    true,
							Required:    true,
							Description: "Public SSH key to be used for the cluster nodes.",
						},
						"region": {
							Type:        schema.TypeString,
							ForceNew:    true,
							Required:    true,
							Description: "The AWS region to deploy the cluster in.",
						},
						"vpc_id": {
							Type:        schema.TypeString,
							ForceNew:    true,
							Optional:    true,
							Description: "The VPC ID to deploy the cluster in. If not provided, VPC will be provisioned dynamically.",
						},
						"control_plane_lb": {
							Type:         schema.TypeString,
							ForceNew:     true,
							Default:      "",
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"", "Internet-facing", "internal"}, false),
							Description:  "Control plane load balancer type. Valid values are `Internet-facing` and `internal`. Defaults to `` (empty string).",
						},
					},
				},
			},
			"machine_pool": {
				Type:     schema.TypeSet,
				Required: true,
				Set:      resourceMachinePoolAwsHash,
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
							Type:        schema.TypeString,
							Required:    true,
							Description: "The name of the machine pool.",
						},
						"count": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Number of nodes in the machine pool.",
						},
						"instance_type": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The instance type to use for the machine pool nodes.",
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
						"node_repave_interval": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     0,
							Description: "Minimum number of seconds node should be Ready, before the next node is selected for repave. Default value is `0`, Applicable only for worker pools.",
						},
						"capacity_type": {
							Type:         schema.TypeString,
							Default:      "on-demand",
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"on-demand", "spot"}, false),
							Description:  "Capacity type is an instance type,  can be 'on-demand' or 'spot'. Defaults to 'on-demand'.",
						},
						"max_price": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Maximum price to bid for spot instances. Only applied when instance type is 'spot'.",
						},
						"update_strategy": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "RollingUpdateScaleOut",
							Description:  "Update strategy for the machine pool. Valid values are `RollingUpdateScaleOut` and `RollingUpdateScaleIn`.",
							ValidateFunc: validation.StringInSlice([]string{"RollingUpdateScaleOut", "RollingUpdateScaleIn"}, false),
						},
						"disk_size_gb": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     65,
							Description: "The disk size in GB for the machine pool nodes.",
						},
						"azs": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "Mutually exclusive with `az_subnets`. Use `azs` for Dynamic provisioning.",
							MinItems:    1,
							Set:         schema.HashString,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"az_subnets": {
							Type:        schema.TypeMap,
							Optional:    true,
							Description: "Mutually exclusive with `azs`. Use `az_subnets` for Static provisioning.",
							Elem: &schema.Schema{
								Type:     schema.TypeString,
								Required: true,
							},
						},
						"additional_security_groups": {
							Type: schema.TypeSet,
							Set:  schema.HashString,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Optional:    true,
							Description: "Additional security groups to attach to the instance.",
						},
						"node": schemas.NodeSchema(),
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

func resourceClusterAwsCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	cluster, err := toAwsCluster(c, d)
	if err != nil {
		return diag.FromErr(err)
	}

	uid, err := c.CreateClusterAws(cluster)
	if err != nil {
		return diag.FromErr(err)
	}

	diagnostics, isError := waitForClusterCreation(ctx, d, uid, diags, c, true)
	if isError {
		return diagnostics
	}

	resourceClusterAwsRead(ctx, d, m)

	return diags
}

//goland:noinspection GoUnhandledErrorResult
func resourceClusterAwsRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
	err = ValidateCloudType("spectrocloud_cluster_aws", cluster)
	if err != nil {
		return diag.FromErr(err)
	}

	diagnostics, done := readCommonFields(c, d, cluster)
	if done {
		return diagnostics
	}

	return flattenCloudConfigAws(cluster.Spec.CloudConfigRef.UID, d, c)
}

func flattenCloudConfigAws(configUID string, d *schema.ResourceData, c *client.V1Client) diag.Diagnostics {
	var diags diag.Diagnostics
	if err := d.Set("cloud_config_id", configUID); err != nil {
		return diag.FromErr(err)
	}
	if err := ReadCommonAttributes(d); err != nil {
		return diag.FromErr(err)
	}

	if config, err := c.GetCloudConfigAws(configUID); err != nil {
		return diag.FromErr(err)
	} else {
		if config.Spec != nil && config.Spec.CloudAccountRef != nil {
			if err := d.Set("cloud_account_id", config.Spec.CloudAccountRef.UID); err != nil {
				return diag.FromErr(err)
			}
		}
		if err := d.Set("cloud_config", flattenClusterConfigsAws(config)); err != nil {
			return diag.FromErr(err)
		}
		mp := flattenMachinePoolConfigsAws(config.Spec.MachinePoolConfig)
		mp, err := flattenNodeMaintenanceStatus(c, d, c.GetNodeStatusMapAws, mp, configUID)
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

func flattenClusterConfigsAws(config *models.V1AwsCloudConfig) []interface{} {
	if config == nil || config.Spec == nil || config.Spec.ClusterConfig == nil {
		return make([]interface{}, 0)
	}

	m := make(map[string]interface{})

	if config.Spec.ClusterConfig.SSHKeyName != "" {
		m["ssh_key_name"] = config.Spec.ClusterConfig.SSHKeyName
	}
	if config.Spec.ClusterConfig.Region != nil {
		m["region"] = *config.Spec.ClusterConfig.Region
	}
	if config.Spec.ClusterConfig.VpcID != "" {
		m["vpc_id"] = config.Spec.ClusterConfig.VpcID
	}
	if config.Spec.ClusterConfig.ControlPlaneLoadBalancer != "" {
		m["control_plane_lb"] = config.Spec.ClusterConfig.ControlPlaneLoadBalancer
	}

	return []interface{}{m}
}

func flattenMachinePoolConfigsAws(machinePools []*models.V1AwsMachinePoolConfig) []interface{} {

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

		oi["min"] = int(machinePool.MinSize)
		oi["max"] = int(machinePool.MaxSize)
		oi["instance_type"] = machinePool.InstanceType
		if machinePool.CapacityType != nil {
			oi["capacity_type"] = machinePool.CapacityType
		}
		if machinePool.SpotMarketOptions != nil {
			oi["max_price"] = machinePool.SpotMarketOptions.MaxPrice
		}
		oi["disk_size_gb"] = int(machinePool.RootDeviceSize)
		if machinePool.SubnetIds != nil {
			oi["az_subnets"] = machinePool.SubnetIds
		} else {
			oi["azs"] = machinePool.Azs
		}

		if machinePool.AdditionalSecurityGroups != nil && len(machinePool.AdditionalSecurityGroups) > 0 {
			additionalSecuritygroup := make([]string, 0)
			for _, sg := range machinePool.AdditionalSecurityGroups {
				additionalSecuritygroup = append(additionalSecuritygroup, sg.ID)
			}
			oi["additional_security_groups"] = additionalSecuritygroup
		}
		ois[i] = oi
	}

	sort.SliceStable(ois, func(i, j int) bool {
		var controlPlaneI, controlPlaneJ bool
		if ois[i].(map[string]interface{})["control_plane"] != nil {
			controlPlaneI = ois[i].(map[string]interface{})["control_plane"].(bool)
		}
		if ois[j].(map[string]interface{})["control_plane"] != nil {
			controlPlaneJ = ois[j].(map[string]interface{})["control_plane"].(bool)
		}

		// If both are control planes or both are not, sort by name
		if controlPlaneI == controlPlaneJ {
			return ois[i].(map[string]interface{})["name"].(string) < ois[j].(map[string]interface{})["name"].(string)
		}

		// Otherwise, control planes come first
		return controlPlaneI && !controlPlaneJ
	})

	return ois
}

func resourceClusterAwsUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	err := validateSystemRepaveApproval(d, c)
	if err != nil {
		return diag.FromErr(err)
	}
	cloudConfigId := d.Get("cloud_config_id").(string)
	//ClusterContext := d.Get("context").(string)
	CloudConfig, err := c.GetCloudConfigAws(cloudConfigId)
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
				if name != "" {
					hash := resourceMachinePoolAwsHash(machinePoolResource)
					vpcId := d.Get("cloud_config").([]interface{})[0].(map[string]interface{})["vpc_id"]

					var err error
					machinePool, err := toMachinePoolAws(machinePoolResource, vpcId.(string))
					if err != nil {
						return diag.FromErr(err)
					}

					if oldMachinePool, ok := osMap[name]; !ok {
						log.Printf("Create machine pool %s", name)
						err = c.CreateMachinePoolAws(cloudConfigId, machinePool)
					} else if hash != resourceMachinePoolAwsHash(oldMachinePool) {
						log.Printf("Change in machine pool %s", name)
						err = c.UpdateMachinePoolAws(cloudConfigId, machinePool)
						// Node Maintenance Actions
						err := resourceNodeAction(c, ctx, nsMap[name], c.GetNodeMaintenanceStatusAws, CloudConfig.Kind, cloudConfigId, name)
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
		}

		// Deleted old machine pools
		for _, mp := range osMap {
			machinePool := mp.(map[string]interface{})
			name := machinePool["name"].(string)
			log.Printf("Deleted machine pool %s", name)
			if err := c.DeleteMachinePoolAws(cloudConfigId, name); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	diagnostics, done := updateCommonFields(d, c)
	if done {
		return diagnostics
	}
	resourceClusterAwsRead(ctx, d, m)
	return diags
}

func toAwsCluster(c *client.V1Client, d *schema.ResourceData) (*models.V1SpectroAwsClusterEntity, error) {
	// gnarly, I know! =/
	cloudConfig := d.Get("cloud_config").([]interface{})[0].(map[string]interface{})

	clusterContext := d.Get("context").(string)
	profiles, err := toProfiles(c, d, clusterContext)
	if err != nil {
		return nil, err
	}
	cluster := &models.V1SpectroAwsClusterEntity{
		Metadata: getClusterMetadata(d),
		Spec: &models.V1SpectroAwsClusterEntitySpec{
			CloudAccountUID: ptr.To(d.Get("cloud_account_id").(string)),
			Profiles:        profiles,
			Policies:        toPolicies(d),
			CloudConfig: &models.V1AwsClusterConfig{
				SSHKeyName:               cloudConfig["ssh_key_name"].(string),
				Region:                   ptr.To(cloudConfig["region"].(string)),
				VpcID:                    cloudConfig["vpc_id"].(string),
				ControlPlaneLoadBalancer: cloudConfig["control_plane_lb"].(string),
			},
		},
	}

	machinePoolConfigs := make([]*models.V1AwsMachinePoolConfigEntity, 0)
	for _, machinePool := range d.Get("machine_pool").(*schema.Set).List() {
		mp, err := toMachinePoolAws(machinePool, cluster.Spec.CloudConfig.VpcID)
		if err != nil {
			return nil, err
		}
		machinePoolConfigs = append(machinePoolConfigs, mp)
	}

	sort.SliceStable(machinePoolConfigs, func(i, j int) bool {
		controlPlaneI := machinePoolConfigs[i].PoolConfig.IsControlPlane
		controlPlaneJ := machinePoolConfigs[j].PoolConfig.IsControlPlane

		// If both are control planes or both are not, sort by name
		if controlPlaneI == controlPlaneJ {
			return *machinePoolConfigs[i].PoolConfig.Name < *machinePoolConfigs[j].PoolConfig.Name
		}

		// Otherwise, control planes come first
		return controlPlaneI && !controlPlaneJ
	})

	cluster.Spec.Machinepoolconfig = machinePoolConfigs
	cluster.Spec.ClusterConfig = toClusterConfig(d)

	return cluster, nil
}

func toMachinePoolAws(machinePool interface{}, vpcId string) (*models.V1AwsMachinePoolConfigEntity, error) {
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
	capacityType := "on-demand" // on-demand by default.
	if m["capacity_type"] != nil && len(m["capacity_type"].(string)) > 0 {
		capacityType = m["capacity_type"].(string)
	}
	azSubnetsConfigs := make([]*models.V1AwsSubnetEntity, 0)
	if m["az_subnets"] != nil && len(m["az_subnets"].(map[string]interface{})) > 0 && vpcId != "" {
		for key, azSubnet := range m["az_subnets"].(map[string]interface{}) {
			azs = append(azs, key)
			azSubnetsConfigs = append(azSubnetsConfigs, &models.V1AwsSubnetEntity{
				ID: azSubnet.(string),
				Az: key,
			})
		}
	}
	if len(azs) == 0 {
		for _, az := range m["azs"].(*schema.Set).List() {
			azs = append(azs, az.(string))
		}
	}
	min := int32(m["count"].(int))
	max := int32(m["count"].(int))

	if m["min"] != nil {
		min = int32(m["min"].(int))
	}

	if m["max"] != nil {
		max = int32(m["max"].(int))
	}

	mp := &models.V1AwsMachinePoolConfigEntity{
		CloudConfig: &models.V1AwsMachinePoolCloudConfigEntity{
			Azs:            azs,
			InstanceType:   ptr.To(m["instance_type"].(string)),
			CapacityType:   &capacityType,
			RootDeviceSize: int64(m["disk_size_gb"].(int)),
			Subnets:        azSubnetsConfigs,
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
			MinSize:                 min,
			MaxSize:                 max,
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

	if capacityType == "spot" {
		maxPrice := "0.0" // default value
		if m["max_price"] != nil && len(m["max_price"].(string)) > 0 {
			maxPrice = m["max_price"].(string)
		}

		mp.CloudConfig.SpotMarketOptions = &models.V1SpotMarketOptions{
			MaxPrice: maxPrice,
		}
	}

	if m["additional_security_groups"] != nil {
		mp.CloudConfig.AdditionalSecurityGroups = setAdditionalSecurityGroups(m)
	}

	return mp, nil
}

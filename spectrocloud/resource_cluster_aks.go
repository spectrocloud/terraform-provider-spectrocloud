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

func resourceClusterAks() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterAksCreate,
		ReadContext:   resourceClusterAksRead,
		UpdateContext: resourceClusterAksUpdate,
		DeleteContext: resourceClusterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceClusterAksImport,
		},
		Description: "Resource for managing AKS clusters in Spectro Cloud through Palette.",

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
				Description: "",
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"", "project", "tenant"}, false),
				Description: "The context of the AKS cluster. Allowed values are `project` or `tenant`. " +
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
						"subscription_id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"resource_group": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"region": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"ssh_key": {
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
							Description: "Public SSH key to be used for the cluster nodes.",
						},
						"private_cluster": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							ForceNew:    true,
							Description: "Whether to create a private cluster(API endpoint). Default is `false`.",
						},

						// fields for static placement are having flat structure as backend currently doesn't support multiple subnets.
						"vnet_name": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},

						"vnet_resource_group": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},

						"vnet_cidr_block": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},

						"worker_subnet_name": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"worker_cidr": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"worker_subnet_security_group_name": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"control_plane_subnet_name": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"control_plane_cidr": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"control_plane_subnet_security_group_name": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},
			"machine_pool": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
							//ForceNew: true,
						},
						"additional_labels": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"node":   schemas.NodeSchema(),
						"taints": schemas.ClusterTaintsSchema(),
						"count": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Number of nodes in the machine pool.",
						},
						"update_strategy": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "RollingUpdateScaleOut",
							Description:  "Update strategy for the machine pool. Valid values are `RollingUpdateScaleOut` and `RollingUpdateScaleIn`.",
							ValidateFunc: validation.StringInSlice([]string{"RollingUpdateScaleOut", "RollingUpdateScaleIn"}, false),
						},
						"instance_type": {
							Type:     schema.TypeString,
							Required: true,
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
						"disk_size_gb": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"is_system_node_pool": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"storage_account_type": {
							Type:     schema.TypeString,
							Required: true,
							//ExactlyOneOf: []string{"Standard_LRS", "Standard_GRS", "Standard_RAGRS", "Standard_ZRS", "Premium_LRS", "Premium_ZRS", "Standard_GZRS", "Standard_RAGZRS"},
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

func resourceClusterAksCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	cluster, err := toAksCluster(c, d)
	if err != nil {
		return diag.FromErr(err)
	}

	uid, err := c.CreateClusterAks(cluster)
	if err != nil {
		return diag.FromErr(err)
	}

	diagnostics, isError := waitForClusterCreation(ctx, d, uid, diags, c, true)
	if isError {
		return diagnostics
	}

	resourceClusterAksRead(ctx, d, m)

	return diags
}

//goland:noinspection GoUnhandledErrorResult
func resourceClusterAksRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
	err = ValidateCloudType("spectrocloud_cluster_aks", cluster)
	if err != nil {
		return diag.FromErr(err)
	}

	configUID := cluster.Spec.CloudConfigRef.UID
	if err := d.Set("cloud_config_id", configUID); err != nil {
		return diag.FromErr(err)
	}
	if err := ReadCommonAttributes(d); err != nil {
		return diag.FromErr(err)
	}
	//ClusterContext := d.Get("context").(string)
	if config, err := c.GetCloudConfigAks(configUID); err != nil {
		return diag.FromErr(err)
	} else {
		if err := d.Set("cloud_account_id", config.Spec.CloudAccountRef.UID); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("cloud_config", flattenClusterConfigsAks(config)); err != nil {
			return diag.FromErr(err)
		}
		mp := flattenMachinePoolConfigsAks(config.Spec.MachinePoolConfig)
		mp, err := flattenNodeMaintenanceStatus(c, d, c.GetNodeStatusMapAks, mp, configUID)
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
	generalWarningForRepave(&diags)
	return diags
}

func ReadCommonAttributes(d *schema.ResourceData) error {
	ForceDelete := d.Get("force_delete").(bool)
	if err := d.Set("force_delete", ForceDelete); err != nil {
		return err
	}

	ForceDeleteDelay := d.Get("force_delete_delay").(int)
	if ForceDeleteDelay == 0 {
		ForceDeleteDelay = 20 // set default value
	}
	if err := d.Set("force_delete_delay", ForceDeleteDelay); err != nil {
		return err
	}

	OsPatchOnBoot := d.Get("os_patch_on_boot").(bool)
	if err := d.Set("os_patch_on_boot", OsPatchOnBoot); err != nil {
		return err
	}

	SkipCompletion := d.Get("skip_completion").(bool)
	if err := d.Set("skip_completion", SkipCompletion); err != nil {
		return err
	}

	ApplySetting := d.Get("apply_setting").(string)
	if ApplySetting == "" {
		ApplySetting = "DownloadAndInstall" // set default value
	}
	if err := d.Set("apply_setting", ApplySetting); err != nil {
		return err
	}

	return nil
}

func flattenClusterConfigsAks(config *models.V1AzureCloudConfig) []interface{} {
	if config == nil || config.Spec == nil || config.Spec.ClusterConfig == nil {
		return make([]interface{}, 0)
	}

	m := make(map[string]interface{})

	if config.Spec.ClusterConfig.SubscriptionID != nil {
		m["subscription_id"] = config.Spec.ClusterConfig.SubscriptionID
	}
	if config.Spec.ClusterConfig.ResourceGroup != "" {
		m["resource_group"] = config.Spec.ClusterConfig.ResourceGroup
	}
	if config.Spec.ClusterConfig.Location != nil {
		m["region"] = *config.Spec.ClusterConfig.Location
	}
	if config.Spec.ClusterConfig.SSHKey != nil {
		m["ssh_key"] = *config.Spec.ClusterConfig.SSHKey
	}
	m["private_cluster"] = config.Spec.ClusterConfig.APIServerAccessProfile.EnablePrivateCluster
	if config.Spec.ClusterConfig.VnetName != "" {
		m["vnet_name"] = config.Spec.ClusterConfig.VnetName
	}
	if config.Spec.ClusterConfig.VnetResourceGroup != "" {
		m["vnet_resource_group"] = config.Spec.ClusterConfig.VnetResourceGroup
	}
	if config.Spec.ClusterConfig.VnetCidrBlock != "" {
		m["vnet_cidr_block"] = config.Spec.ClusterConfig.VnetCidrBlock
	}
	if config.Spec.ClusterConfig.WorkerSubnet != nil {
		m["worker_subnet_name"] = config.Spec.ClusterConfig.WorkerSubnet.Name
		m["worker_cidr"] = config.Spec.ClusterConfig.WorkerSubnet.CidrBlock
		if config.Spec.ClusterConfig.WorkerSubnet.SecurityGroupName != "" {
			m["worker_subnet_security_group_name"] = config.Spec.ClusterConfig.WorkerSubnet.SecurityGroupName
		}
	}
	if config.Spec.ClusterConfig.ControlPlaneSubnet != nil {
		m["control_plane_subnet_name"] = config.Spec.ClusterConfig.ControlPlaneSubnet.Name
		m["control_plane_cidr"] = config.Spec.ClusterConfig.ControlPlaneSubnet.CidrBlock
		if config.Spec.ClusterConfig.ControlPlaneSubnet.SecurityGroupName != "" {
			m["control_plane_subnet_security_group_name"] = config.Spec.ClusterConfig.ControlPlaneSubnet.SecurityGroupName
		}
	}

	return []interface{}{m}
}

func flattenMachinePoolConfigsAks(machinePools []*models.V1AzureMachinePoolConfig) []interface{} {
	if machinePools == nil {
		return make([]interface{}, 0)
	}

	ois := make([]interface{}, 0)
	for _, machinePool := range machinePools {
		oi := make(map[string]interface{})

		FlattenAdditionalLabelsAndTaints(machinePool.AdditionalLabels, machinePool.Taints, oi)

		if machinePool.IsControlPlane != nil && *machinePool.IsControlPlane {
			continue
		}

		oi["name"] = machinePool.Name
		oi["count"] = int(machinePool.Size)
		oi["min"] = int(machinePool.MinSize)
		oi["max"] = int(machinePool.MaxSize)
		flattenUpdateStrategy(machinePool.UpdateStrategy, oi)

		oi["instance_type"] = machinePool.InstanceType
		oi["disk_size_gb"] = int(machinePool.OsDisk.DiskSizeGB)
		oi["is_system_node_pool"] = machinePool.IsSystemNodePool
		oi["storage_account_type"] = machinePool.OsDisk.ManagedDisk.StorageAccountType
		oi["min"] = int(machinePool.MinSize)
		oi["max"] = int(machinePool.MaxSize)
		ois = append(ois, oi)
	}
	return ois
}

func resourceClusterAksUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	err := validateSystemRepaveApproval(d, c)
	if err != nil {
		return diag.FromErr(err)
	}
	cloudConfigId := d.Get("cloud_config_id").(string)
	CloudConfig, err := c.GetCloudConfigAks(cloudConfigId)
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
			// since known issue in TF SDK: https://github.com/hashicorp/terraform-plugin-sdk/issues/588
			if machinePoolResource["name"].(string) != "" {
				name := machinePoolResource["name"].(string)
				hash := resourceMachinePoolAksHash(machinePoolResource)

				machinePool := toMachinePoolAks(machinePoolResource)

				var err error
				if oldMachinePool, ok := osMap[name]; !ok {
					log.Printf("Create machine pool %s", name)
					err = c.CreateMachinePoolAks(cloudConfigId, machinePool)
				} else if hash != resourceMachinePoolAksHash(oldMachinePool) {
					log.Printf("Change in machine pool %s", name)
					err = c.UpdateMachinePoolAks(cloudConfigId, machinePool)
					// Node Maintenance Actions
					err := resourceNodeAction(c, ctx, nsMap[name], c.GetNodeMaintenanceStatusAks, CloudConfig.Kind, cloudConfigId, name)
					if err != nil {
						return diag.FromErr(err)
					}
				}
				if err != nil {
					return diag.FromErr(err)
				}
				delete(osMap, name)
			}
		}

		// Deleted old machine pools
		for _, mp := range osMap {
			machinePool := mp.(map[string]interface{})
			name := machinePool["name"].(string)
			log.Printf("Deleted machine pool %s", name)
			if err := c.DeleteMachinePoolAks(cloudConfigId, name); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	diagnostics, done := updateCommonFields(d, c)
	if done {
		return diagnostics
	}

	resourceClusterAksRead(ctx, d, m)

	return diags
}

func toAksCluster(c *client.V1Client, d *schema.ResourceData) (*models.V1SpectroAzureClusterEntity, error) {
	config := d.Get("cloud_config").([]interface{})
	cloudConfig := config[0]
	cloudConfigMap := cloudConfig.(map[string]interface{})

	// static placement support
	var vnetname string
	if cloudConfigMap["vnet_name"] != nil {
		vnetname = cloudConfigMap["vnet_name"].(string)
	}

	var vnetResourceGroup string
	if cloudConfigMap["vnet_resource_group"] != nil {
		vnetResourceGroup = cloudConfigMap["vnet_resource_group"].(string)
	}

	var vnetcidr string
	if cloudConfigMap["vnet_cidr_block"] != nil {
		vnetcidr = cloudConfigMap["vnet_cidr_block"].(string)
	}

	var workerSubnet *models.V1Subnet
	if cloudConfigMap["worker_subnet_name"] != nil && cloudConfigMap["worker_cidr"] != nil {
		workerSubnet = &models.V1Subnet{
			Name:              cloudConfigMap["worker_subnet_name"].(string),
			CidrBlock:         cloudConfigMap["worker_cidr"].(string),
			SecurityGroupName: cloudConfigMap["worker_subnet_security_group_name"].(string),
		}
	}

	var controlPlaneSubnet *models.V1Subnet
	if cloudConfigMap["control_plane_subnet_name"] != "" && cloudConfigMap["control_plane_cidr"] != "" {
		controlPlaneSubnet = &models.V1Subnet{
			Name:              cloudConfigMap["control_plane_subnet_name"].(string),
			CidrBlock:         cloudConfigMap["control_plane_cidr"].(string),
			SecurityGroupName: cloudConfigMap["control_plane_subnet_security_group_name"].(string),
		}
	}

	clusterContext := d.Get("context").(string)
	profiles, err := toProfiles(c, d, clusterContext)
	if err != nil {
		return nil, err
	}
	cluster := &models.V1SpectroAzureClusterEntity{
		Metadata: getClusterMetadata(d),
		Spec: &models.V1SpectroAzureClusterEntitySpec{
			CloudAccountUID: ptr.To(d.Get("cloud_account_id").(string)),
			Profiles:        profiles,
			Policies:        toPolicies(d),
			CloudConfig: &models.V1AzureClusterConfig{
				Location:      ptr.To(cloudConfigMap["region"].(string)),
				ResourceGroup: cloudConfigMap["resource_group"].(string),
				SSHKey:        ptr.To(cloudConfigMap["ssh_key"].(string)),
				APIServerAccessProfile: &models.V1APIServerAccessProfile{
					EnablePrivateCluster: cloudConfigMap["private_cluster"].(bool),
				},
				SubscriptionID:     ptr.To(cloudConfigMap["subscription_id"].(string)),
				VnetName:           vnetname,
				VnetResourceGroup:  vnetResourceGroup,
				VnetCidrBlock:      vnetcidr,
				ControlPlaneSubnet: controlPlaneSubnet,
				WorkerSubnet:       workerSubnet,
			},
		},
	}

	machinePoolConfigs := make([]*models.V1AzureMachinePoolConfigEntity, 0)
	for _, machinePool := range d.Get("machine_pool").([]interface{}) {
		mp := toMachinePoolAks(machinePool)
		machinePoolConfigs = append(machinePoolConfigs, mp)
	}
	cluster.Spec.Machinepoolconfig = machinePoolConfigs
	cluster.Spec.ClusterConfig = toClusterConfig(d)

	return cluster, nil
}

func toMachinePoolAks(machinePool interface{}) *models.V1AzureMachinePoolConfigEntity {
	m := machinePool.(map[string]interface{})

	labels := make([]string, 0)
	controlPlane, _ := m["control_plane"].(bool)
	if controlPlane {
		labels = append(labels, "control-plane")
	} else {
		labels = append(labels, "worker")
	}

	min := int32(m["count"].(int))
	max := int32(m["count"].(int))

	if m["min"] != nil {
		min = int32(m["min"].(int))
	}

	if m["max"] != nil {
		max = int32(m["max"].(int))
	}

	mp := &models.V1AzureMachinePoolConfigEntity{
		CloudConfig: &models.V1AzureMachinePoolCloudConfigEntity{
			InstanceType: m["instance_type"].(string),
			OsDisk: &models.V1AzureOSDisk{
				DiskSizeGB: int32(m["disk_size_gb"].(int)),
				ManagedDisk: &models.V1ManagedDisk{
					StorageAccountType: m["storage_account_type"].(string),
				},
				OsType: "",
			},
			IsSystemNodePool: m["is_system_node_pool"].(bool),
		},
		ManagedPoolConfig: &models.V1AzureManagedMachinePoolConfig{
			IsSystemNodePool: m["is_system_node_pool"].(bool),
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
			MinSize: min,
			MaxSize: max,
		},
	}

	return mp
}

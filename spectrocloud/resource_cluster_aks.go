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

func resourceClusterAks() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterAksCreate,
		ReadContext:   resourceClusterAksRead,
		UpdateContext: resourceClusterAksUpdate,
		DeleteContext: resourceClusterDelete,
		Description:   "Resource for managing AKS clusters in Spectro Cloud through Palette.",

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
				Description:  "The context of the AKS cluster. Can be `project` or `tenant`. Default is `project`.",
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
						},
						"resource_group": {
							Type:     schema.TypeString,
							Required: true,
						},
						"region": {
							Type:     schema.TypeString,
							Required: true,
						},
						"ssh_key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"private_cluster": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Whether to create a private cluster(API endpoint). Default is `false`.",
						},

						// fields for static placement are having flat structure as backend currently doesn't support multiple subnets.
						"vnet_name": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"vnet_resource_group": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"vnet_cidr_block": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"worker_subnet_name": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"worker_cidr": {
							Type:     schema.TypeString,
							Optional: true,
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
	c := m.(*client.V1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	cluster, err := toAksCluster(c, d)
	if err != nil {
		return diag.FromErr(err)
	}

	ClusterContext := d.Get("context").(string)
	uid, err := c.CreateClusterAks(cluster, ClusterContext)
	if err != nil {
		return diag.FromErr(err)
	}

	diagnostics, isError := waitForClusterCreation(ctx, d, ClusterContext, uid, diags, c, true)
	if isError {
		return diagnostics
	}

	resourceClusterAksRead(ctx, d, m)

	return diags
}

//goland:noinspection GoUnhandledErrorResult
func resourceClusterAksRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	configUID := cluster.Spec.CloudConfigRef.UID
	if err := d.Set("cloud_config_id", configUID); err != nil {
		return diag.FromErr(err)
	}
	ClusterContext := d.Get("context").(string)
	if config, err := c.GetCloudConfigAks(configUID, ClusterContext); err != nil {
		return diag.FromErr(err)
	} else {
		mp := flattenMachinePoolConfigsAks(config.Spec.MachinePoolConfig)
		if err := d.Set("machine_pool", mp); err != nil {
			return diag.FromErr(err)
		}
	}

	diagnostics, done := readCommonFields(c, d, cluster)
	if done {
		return diagnostics
	}

	return diags
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
		flattenUpdateStrategy(machinePool.UpdateStrategy, oi)

		oi["instance_type"] = machinePool.InstanceType
		oi["disk_size_gb"] = int(machinePool.OsDisk.DiskSizeGB)
		oi["is_system_node_pool"] = machinePool.IsSystemNodePool
		oi["storage_account_type"] = machinePool.OsDisk.ManagedDisk.StorageAccountType
		ois = append(ois, oi)
	}
	return ois
}

func resourceClusterAksUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

		os := oraw.([]interface{})
		ns := nraw.([]interface{})

		osMap := make(map[string]interface{})
		for _, mp := range os {
			machinePool := mp.(map[string]interface{})
			osMap[machinePool["name"].(string)] = machinePool
		}

		for _, mp := range ns {
			machinePoolResource := mp.(map[string]interface{})
			// since known issue in TF SDK: https://github.com/hashicorp/terraform-plugin-sdk/issues/588
			if machinePoolResource["name"].(string) != "" {
				name := machinePoolResource["name"].(string)
				hash := resourceMachinePoolAksHash(machinePoolResource)

				machinePool := toMachinePoolAks(machinePoolResource)

				var err error
				if oldMachinePool, ok := osMap[name]; !ok {
					log.Printf("Create machine pool %s", name)
					err = c.CreateMachinePoolAks(cloudConfigId, machinePool, ClusterContext)
				} else if hash != resourceMachinePoolAksHash(oldMachinePool) {
					log.Printf("Change in machine pool %s", name)
					err = c.UpdateMachinePoolAks(cloudConfigId, machinePool, ClusterContext)
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
			if err := c.DeleteMachinePoolAks(cloudConfigId, name, ClusterContext); err != nil {
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
			Name:      cloudConfigMap["worker_subnet_name"].(string),
			CidrBlock: cloudConfigMap["worker_cidr"].(string),
		}
	}

	profiles, err := toProfiles(c, d)
	if err != nil {
		return nil, err
	}
	cluster := &models.V1SpectroAzureClusterEntity{
		Metadata: &models.V1ObjectMeta{
			Name:   d.Get("name").(string),
			UID:    d.Id(),
			Labels: toTags(d),
		},
		Spec: &models.V1SpectroAzureClusterEntitySpec{
			CloudAccountUID: types.Ptr(d.Get("cloud_account_id").(string)),
			Profiles:        profiles,
			Policies:        toPolicies(d),
			CloudConfig: &models.V1AzureClusterConfig{
				Location:      types.Ptr(cloudConfigMap["region"].(string)),
				ResourceGroup: cloudConfigMap["resource_group"].(string),
				SSHKey:        types.Ptr(cloudConfigMap["ssh_key"].(string)),
				APIServerAccessProfile: &models.V1APIServerAccessProfile{
					EnablePrivateCluster: cloudConfigMap["private_cluster"].(bool),
				},
				SubscriptionID:     types.Ptr(cloudConfigMap["subscription_id"].(string)),
				VnetName:           vnetname,
				VnetResourceGroup:  vnetResourceGroup,
				VnetCidrBlock:      vnetcidr,
				ControlPlaneSubnet: workerSubnet,
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
		labels = append(labels, "master")
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
			Name:             types.Ptr(m["name"].(string)),
			Size:             types.Ptr(int32(m["count"].(int))),
			UpdateStrategy: &models.V1UpdateStrategy{
				Type: getUpdateStrategy(m),
			},
			MinSize: min,
			MaxSize: max,
		},
	}

	return mp
}

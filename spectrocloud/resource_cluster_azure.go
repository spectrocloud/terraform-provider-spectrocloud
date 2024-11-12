package spectrocloud

import (
	"context"
	"fmt"
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

func resourceClusterAzure() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterAzureCreate,
		ReadContext:   resourceClusterAzureRead,
		UpdateContext: resourceClusterAzureUpdate,
		DeleteContext: resourceClusterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceClusterAzureImport,
		},
		Description: "Resource for managing Azure clusters in Spectro Cloud through Palette.",

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the cluster. This name will be used to create the cluster in Azure.",
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"", "project", "tenant"}, false),
				Description: "The context of the Azure cluster. Allowed values are `project` or `tenant`. " +
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
				Required:    true,
				ForceNew:    true,
				Description: "ID of the cloud account to be used for the cluster. This cloud account must be of type `azure`.",
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
						"subscription_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Azure subscription ID. This can be found in the Azure portal under `Subscriptions`.",
						},
						"resource_group": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Azure resource group. This can be found in the Azure portal under `Resource groups`.",
						},
						"region": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Azure region. This can be found in the Azure portal under `Resource groups`.",
						},
						"ssh_key": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Public SSH key to be used for the cluster nodes.",
						},
						"storage_account_name": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "Azure storage account name.",
						},
						"container_name": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "Container name within your azure storage account.",
						},
						"network_resource_group": {
							Type:         schema.TypeString,
							Optional:     true,
							RequiredWith: []string{"cloud_config.0.virtual_network_name", "cloud_config.0.virtual_network_cidr_block", "cloud_config.0.control_plane_subnet", "cloud_config.0.worker_node_subnet"},
							Description:  "Azure network resource group in which the cluster is to be provisioned.",
						},
						"virtual_network_name": {
							Type:         schema.TypeString,
							Optional:     true,
							RequiredWith: []string{"cloud_config.0.network_resource_group", "cloud_config.0.virtual_network_cidr_block", "cloud_config.0.control_plane_subnet", "cloud_config.0.worker_node_subnet"},
							Description:  "Azure virtual network in which the cluster is to be provisioned.",
						},
						"virtual_network_cidr_block": {
							Type:         schema.TypeString,
							Optional:     true,
							RequiredWith: []string{"cloud_config.0.network_resource_group", "cloud_config.0.virtual_network_name", "cloud_config.0.control_plane_subnet", "cloud_config.0.worker_node_subnet"},
							Description:  "Azure virtual network cidr block in which the cluster is to be provisioned.",
						},
						"control_plane_subnet": schemas.SubnetSchema(),
						"worker_node_subnet":   schemas.SubnetSchema(),
					},
				},
			},
			"machine_pool": {
				Type:     schema.TypeSet,
				Required: true,
				Set:      resourceMachinePoolAzureHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"additional_labels": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
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
							Description: "Name of the machine pool. This must be unique within the cluster.",
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
							Type:        schema.TypeString,
							Required:    true,
							Description: "Azure instance type from the Azure portal.",
						},
						"update_strategy": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "RollingUpdateScaleOut",
							Description:  "Update strategy for the machine pool. Valid values are `RollingUpdateScaleOut` and `RollingUpdateScaleIn`.",
							ValidateFunc: validation.StringInSlice([]string{"RollingUpdateScaleOut", "RollingUpdateScaleIn"}, false),
						},
						"disk": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							// Unfortunately can't do any defaulting
							// https://github.com/hashicorp/terraform-plugin-sdk/issues/142
							//DefaultFunc: func() (interface{}, error) {
							//	disk := map[string]interface{}{
							//		"size_gb": 55,
							//		"type" : "Standard_LRS",
							//	}
							//	//return "us-west", nil
							//	return []interface{}{disk}, nil
							//},
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"size_gb": {
										Type:        schema.TypeInt,
										Required:    true,
										Description: "Size of the disk in GB.",
									},
									"type": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Type of the disk. Valid values are `Standard_LRS`, `StandardSSD_LRS`, `Premium_LRS`.",
									},
								},
							},
							Description: "Disk configuration for the machine pool.",
						},
						"azs": {
							Type:     schema.TypeSet,
							Optional: true,
							Set:      schema.HashString,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "Availability zones for the machine pool. Check if your region provides availability zones on [the Azure documentation](https://learn.microsoft.com/en-us/azure/reliability/availability-zones-service-support#azure-regions-with-availability-zone-support). Default value is `[\"\"]`.",
							DefaultFunc: func() (any, error) {
								return []string{""}, nil
							},
						},
						"is_system_node_pool": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Whether this machine pool is a system node pool. Default value is `false'.",
						},
						"os_type": {
							Type:     schema.TypeString,
							Optional: true,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								return false
							},
							Default:      "Linux",
							ValidateFunc: validation.StringInSlice([]string{"Linux", "Windows"}, false),
							Description:  "Operating system type for the machine pool. Valid values are `Linux` and `Windows`. Defaults to `Linux`.",
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

func resourceClusterAzureCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	cluster, err := toAzureCluster(c, d)
	if err != nil {
		return diag.FromErr(err)
	}
	diags = validateCPPoolCount(cluster.Spec.Machinepoolconfig)
	if diags != nil {
		return diags
	}

	uid, err := c.CreateClusterAzure(cluster)
	if err != nil {
		return diag.FromErr(err)
	}

	diagnostics, isError := waitForClusterCreation(ctx, d, uid, diags, c, true)
	if isError {
		return diagnostics
	}

	resourceClusterAzureRead(ctx, d, m)

	return diags
}

//goland:noinspection GoUnhandledErrorResult
func resourceClusterAzureRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
	err = ValidateCloudType("spectrocloud_cluster_azure", cluster)
	if err != nil {
		return diag.FromErr(err)
	}
	diagnostics, done := readCommonFields(c, d, cluster)
	if done {
		return diagnostics
	}

	return flattenCloudConfigAzure(cluster.Spec.CloudConfigRef.UID, d, c)
}
func flattenClusterConfigsAzure(config *models.V1AzureCloudConfig) []interface{} {
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
		m["region"] = config.Spec.ClusterConfig.Location
	}
	if config.Spec.ClusterConfig.SSHKey != nil {
		m["ssh_key"] = config.Spec.ClusterConfig.SSHKey
	}
	if config.Spec.ClusterConfig.StorageAccountName != "" {
		m["storage_account_name"] = config.Spec.ClusterConfig.StorageAccountName
	}
	if config.Spec.ClusterConfig.ContainerName != "" {
		m["container_name"] = config.Spec.ClusterConfig.ContainerName
	}
	if config.Spec.ClusterConfig.VnetResourceGroup != "" {
		m["network_resource_group"] = config.Spec.ClusterConfig.VnetResourceGroup
	}
	if config.Spec.ClusterConfig.VnetName != "" {
		m["virtual_network_name"] = config.Spec.ClusterConfig.VnetName
	}
	if config.Spec.ClusterConfig.VnetCidrBlock != "" {
		m["virtual_network_cidr_block"] = config.Spec.ClusterConfig.VnetCidrBlock
	}
	if config.Spec.ClusterConfig.VnetResourceGroup != "" && config.Spec.ClusterConfig.VnetName != "" && config.Spec.ClusterConfig.VnetCidrBlock != "" {
		if config.Spec.ClusterConfig.ControlPlaneSubnet != nil {
			cpSubnet := map[string]interface{}{
				"name":                config.Spec.ClusterConfig.ControlPlaneSubnet.Name,
				"cidr_block":          config.Spec.ClusterConfig.ControlPlaneSubnet.CidrBlock,
				"security_group_name": config.Spec.ClusterConfig.ControlPlaneSubnet.SecurityGroupName,
			}
			m["control_plane_subnet"] = []interface{}{cpSubnet}
		}
		if config.Spec.ClusterConfig.WorkerSubnet != nil {
			workerSubnet := map[string]interface{}{
				"name":                config.Spec.ClusterConfig.WorkerSubnet.Name,
				"cidr_block":          config.Spec.ClusterConfig.WorkerSubnet.CidrBlock,
				"security_group_name": config.Spec.ClusterConfig.WorkerSubnet.SecurityGroupName,
			}
			m["worker_node_subnet"] = []interface{}{workerSubnet}
		}
	}

	return []interface{}{m}
}
func flattenCloudConfigAzure(configUID string, d *schema.ResourceData, c *client.V1Client) diag.Diagnostics {
	var diags diag.Diagnostics
	if err := d.Set("cloud_config_id", configUID); err != nil {
		return diag.FromErr(err)
	}
	if config, err := c.GetCloudConfigAzure(configUID); err != nil {
		return diag.FromErr(err)
	} else {
		if err := d.Set("cloud_account_id", config.Spec.CloudAccountRef.UID); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("cloud_config", flattenClusterConfigsAzure(config)); err != nil {
			return diag.FromErr(err)
		}
		mp := flattenMachinePoolConfigsAzure(config.Spec.MachinePoolConfig)
		mp, err := flattenNodeMaintenanceStatus(c, d, c.GetNodeStatusMapAzure, mp, configUID)
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

func flattenMachinePoolConfigsAzure(machinePools []*models.V1AzureMachinePoolConfig) []interface{} {

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
		oi["count"] = machinePool.Size
		flattenUpdateStrategy(machinePool.UpdateStrategy, oi)

		oi["instance_type"] = machinePool.InstanceType
		oi["is_system_node_pool"] = machinePool.IsSystemNodePool

		oi["azs"] = machinePool.Azs
		oi["os_type"] = machinePool.OsType
		if machinePool.OsDisk != nil {
			d := make(map[string]interface{})
			d["size_gb"] = machinePool.OsDisk.DiskSizeGB
			d["type"] = machinePool.OsDisk.ManagedDisk.StorageAccountType

			oi["disk"] = []interface{}{d}
		}

		ois[i] = oi
	}

	return ois
}

func resourceClusterAzureUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
	CloudConfig, err := c.GetCloudConfigAzure(cloudConfigId)
	if err != nil {
		return diag.FromErr(err)
	}
	if d.HasChange("machine_pool") {
		cluster, err := toAzureCluster(c, d)
		if err != nil {
			return diag.FromErr(err)
		}
		diags = validateCPPoolCount(cluster.Spec.Machinepoolconfig)
		if diags != nil {
			return diags
		}
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
				hash := resourceMachinePoolAzureHash(machinePoolResource)
				var err error
				machinePool, err := toMachinePoolAzure(machinePoolResource)
				if err != nil {
					diag.FromErr(err)
				}

				if oldMachinePool, ok := osMap[name]; !ok {
					log.Printf("Create machine pool %s", name)
					err = c.CreateMachinePoolAzure(cloudConfigId, machinePool)
				} else if hash != resourceMachinePoolAzureHash(oldMachinePool) {
					log.Printf("Change in machine pool %s", name)
					err = c.UpdateMachinePoolAzure(cloudConfigId, machinePool)
					// Node Maintenance Actions
					err := resourceNodeAction(c, ctx, nsMap[name], c.GetNodeMaintenanceStatusAzure, CloudConfig.Kind, cloudConfigId, name)
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
			if err := c.DeleteMachinePoolAzure(cloudConfigId, name); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	diagnostics, done := updateCommonFields(d, c)
	if done {
		return diagnostics
	}

	resourceClusterAzureRead(ctx, d, m)

	return diags
}

func toAzureCluster(c *client.V1Client, d *schema.ResourceData) (*models.V1SpectroAzureClusterEntity, error) {
	// gnarly, I know! =/
	cloudConfig := d.Get("cloud_config").([]interface{})[0].(map[string]interface{})
	//clientSecret := strfmt.Password(d.Get("azure_client_secret").(string))
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
				Location:           ptr.To(cloudConfig["region"].(string)),
				SSHKey:             ptr.To(cloudConfig["ssh_key"].(string)),
				SubscriptionID:     ptr.To(cloudConfig["subscription_id"].(string)),
				ResourceGroup:      cloudConfig["resource_group"].(string),
				StorageAccountName: cloudConfig["storage_account_name"].(string),
				ContainerName:      cloudConfig["container_name"].(string),
			},
		},
	}
	// setting static placements
	toStaticPlacement(cluster, cloudConfig)
	//for _, machinePool := range d.Get("machine_pool").([]interface{}) {
	machinePoolConfigs := make([]*models.V1AzureMachinePoolConfigEntity, 0)
	for _, machinePool := range d.Get("machine_pool").(*schema.Set).List() {
		mp, err := toMachinePoolAzure(machinePool)
		if err != nil {
			return nil, err
		}
		machinePoolConfigs = append(machinePoolConfigs, mp)
	}

	cluster.Spec.Machinepoolconfig = machinePoolConfigs
	cluster.Spec.ClusterConfig = toClusterConfig(d)

	return cluster, nil
}

func toStaticPlacement(c *models.V1SpectroAzureClusterEntity, cloudConfig map[string]interface{}) {
	if v, ok := cloudConfig["network_resource_group"]; ok && v != "" {
		c.Spec.CloudConfig.VnetResourceGroup = cloudConfig["network_resource_group"].(string)
		c.Spec.CloudConfig.VnetName = cloudConfig["virtual_network_name"].(string)
		c.Spec.CloudConfig.VnetCidrBlock = cloudConfig["virtual_network_cidr_block"].(string)
		cpSubnet := cloudConfig["control_plane_subnet"].([]interface{})[0].(map[string]interface{})
		c.Spec.CloudConfig.ControlPlaneSubnet = &models.V1Subnet{
			CidrBlock:         cpSubnet["cidr_block"].(string),
			Name:              cpSubnet["name"].(string),
			SecurityGroupName: cpSubnet["security_group_name"].(string),
		}
		workerSubnet := cloudConfig["worker_node_subnet"].([]interface{})[0].(map[string]interface{})
		c.Spec.CloudConfig.WorkerSubnet = &models.V1Subnet{
			CidrBlock:         workerSubnet["cidr_block"].(string),
			Name:              workerSubnet["name"].(string),
			SecurityGroupName: workerSubnet["security_group_name"].(string),
		}
	}
}

func toMachinePoolAzure(machinePool interface{}) (*models.V1AzureMachinePoolConfigEntity, error) {
	m := machinePool.(map[string]interface{})

	labels := make([]string, 0)
	controlPlane := m["control_plane"].(bool)
	controlPlaneAsWorker := m["control_plane_as_worker"].(bool)
	if controlPlane {
		labels = append(labels, "control-plane")
	} else {
		labels = append(labels, "worker")
	}

	var diskSize, diskType = DefaultDiskSize, DefaultDiskType
	disks := m["disk"].([]interface{})
	if len(disks) > 0 {
		disk0 := disks[0].(map[string]interface{})
		diskSize = disk0["size_gb"].(int)
		diskType = disk0["type"].(string)
	}

	azs := make([]string, 0)
	for _, az := range m["azs"].(*schema.Set).List() {
		azs = append(azs, az.(string))
	}

	osType := models.V1OsTypeLinux

	if m["os_type"] != "" {
		os_type := m["os_type"].(string)
		if os_type == "Windows" {
			osType = models.V1OsTypeWindows
		}
	}

	mp := &models.V1AzureMachinePoolConfigEntity{
		CloudConfig: &models.V1AzureMachinePoolCloudConfigEntity{
			Azs:          azs,
			InstanceType: m["instance_type"].(string),
			OsDisk: &models.V1AzureOSDisk{
				DiskSizeGB: int32(diskSize),
				ManagedDisk: &models.V1ManagedDisk{
					StorageAccountType: diskType,
				},
				OsType: osType,
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

func validateCPPoolCount(machinePool []*models.V1AzureMachinePoolConfigEntity) diag.Diagnostics {
	for _, machineConfig := range machinePool {
		if machineConfig.PoolConfig.IsControlPlane {
			if *machineConfig.PoolConfig.Size%2 == 0 {
				return diag.FromErr(fmt.Errorf("The control-plane node pool size should be in an odd number. But it set to an even number '%d' in node name '%s' ", *machineConfig.PoolConfig.Size, *machineConfig.PoolConfig.Name))
			}
		}
	}
	return nil
}

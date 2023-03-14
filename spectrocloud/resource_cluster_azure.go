package spectrocloud

import (
	"context"
	"fmt"
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

func resourceClusterAzure() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterAzureCreate,
		ReadContext:   resourceClusterAzureRead,
		UpdateContext: resourceClusterAzureUpdate,
		DeleteContext: resourceClusterDelete,
		Description:   "Resource for managing Azure clusters in Spectro Cloud through Palette.",

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
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Apply setting for the cluster. This can be set to `on_create` or `on_update`.",
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
							Description: "SSH key to be used for the cluster nodes.",
						},
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
						"instance_type": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Azure instance type from the Azure portal.",
						},
						"update_strategy": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "RollingUpdateScaleOut",
							Description: "Update strategy for the machine pool. Valid values are `RollingUpdateScaleOut` and `RollingUpdateScaleIn`.",
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
							Required: true,
							Set:      schema.HashString,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "Availability zones for the machine pool.",
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
		},
	}
}

func resourceClusterAzureCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	cluster := toAzureCluster(c, d)
	diags = validateMasterPoolCount(cluster.Spec.Machinepoolconfig)
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
	c := m.(*client.V1Client)

	var diags diag.Diagnostics

	uid := d.Id()

	cluster, err := c.GetCluster(uid)
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

	return flattenCloudConfigAzure(cluster.Spec.CloudConfigRef.UID, d, c)
}

func flattenCloudConfigAzure(configUID string, d *schema.ResourceData, c *client.V1Client) diag.Diagnostics {
	if err := d.Set("cloud_config_id", configUID); err != nil {
		return diag.FromErr(err)
	}
	if config, err := c.GetCloudConfigAzure(configUID); err != nil {
		return diag.FromErr(err)
	} else {
		mp := flattenMachinePoolConfigsAzure(config.Spec.MachinePoolConfig)
		if err := d.Set("machine_pool", mp); err != nil {
			return diag.FromErr(err)
		}
	}

	return diag.Diagnostics{}
}

func flattenMachinePoolConfigsAzure(machinePools []*models.V1AzureMachinePoolConfig) []interface{} {

	if machinePools == nil {
		return make([]interface{}, 0)
	}

	ois := make([]interface{}, len(machinePools))

	for i, machinePool := range machinePools {
		oi := make(map[string]interface{})

		SetAdditionalLabelsAndTaints(machinePool.AdditionalLabels, machinePool.Taints, oi)

		oi["control_plane"] = machinePool.IsControlPlane
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
	c := m.(*client.V1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	cloudConfigId := d.Get("cloud_config_id").(string)

	if d.HasChange("machine_pool") {
		cluster := toAzureCluster(c, d)
		diags = validateMasterPoolCount(cluster.Spec.Machinepoolconfig)
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

		for _, mp := range ns.List() {
			machinePoolResource := mp.(map[string]interface{})
			name := machinePoolResource["name"].(string)
			hash := resourceMachinePoolAzureHash(machinePoolResource)

			machinePool := toMachinePoolAzure(machinePoolResource)

			var err error
			if oldMachinePool, ok := osMap[name]; !ok {
				log.Printf("Create machine pool %s", name)
				err = c.CreateMachinePoolAzure(cloudConfigId, machinePool)
			} else if hash != resourceMachinePoolAzureHash(oldMachinePool) {
				log.Printf("Change in machine pool %s", name)
				err = c.UpdateMachinePoolAzure(cloudConfigId, machinePool)
			}

			if err != nil {
				return diag.FromErr(err)
			}

			// Processed (if exists)
			delete(osMap, name)
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
	//TODO(saamalik) update for cluster as well
	//if err := waitForClusterU(ctx, c, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
	//	return diag.FromErr(err)
	//}

	diagnostics, done := updateCommonFields(d, c)
	if done {
		return diagnostics
	}

	resourceClusterAzureRead(ctx, d, m)

	return diags
}

func toAzureCluster(c *client.V1Client, d *schema.ResourceData) *models.V1SpectroAzureClusterEntity {
	// gnarly, I know! =/
	cloudConfig := d.Get("cloud_config").([]interface{})[0].(map[string]interface{})
	//clientSecret := strfmt.Password(d.Get("azure_client_secret").(string))

	cluster := &models.V1SpectroAzureClusterEntity{
		Metadata: &models.V1ObjectMeta{
			Name:   d.Get("name").(string),
			UID:    d.Id(),
			Labels: toTags(d),
		},
		Spec: &models.V1SpectroAzureClusterEntitySpec{
			CloudAccountUID: types.Ptr(d.Get("cloud_account_id").(string)),
			Profiles:        toProfiles(c, d),
			Policies:        toPolicies(d),
			CloudConfig: &models.V1AzureClusterConfig{
				Location:       types.Ptr(cloudConfig["region"].(string)),
				SSHKey:         types.Ptr(cloudConfig["ssh_key"].(string)),
				SubscriptionID: types.Ptr(cloudConfig["subscription_id"].(string)),
				ResourceGroup:  cloudConfig["resource_group"].(string),
			},
		},
	}

	//for _, machinePool := range d.Get("machine_pool").([]interface{}) {
	machinePoolConfigs := make([]*models.V1AzureMachinePoolConfigEntity, 0)
	for _, machinePool := range d.Get("machine_pool").(*schema.Set).List() {
		mp := toMachinePoolAzure(machinePool)
		machinePoolConfigs = append(machinePoolConfigs, mp)
	}

	cluster.Spec.Machinepoolconfig = machinePoolConfigs
	cluster.Spec.ClusterConfig = toClusterConfig(d)

	return cluster
}

func toMachinePoolAzure(machinePool interface{}) *models.V1AzureMachinePoolConfigEntity {
	m := machinePool.(map[string]interface{})

	labels := make([]string, 0)
	controlPlane := m["control_plane"].(bool)
	controlPlaneAsWorker := m["control_plane_as_worker"].(bool)
	if controlPlane {
		labels = append(labels, "master")
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
			Name:             types.Ptr(m["name"].(string)),
			Size:             types.Ptr(int32(m["count"].(int))),
			UpdateStrategy: &models.V1UpdateStrategy{
				Type: getUpdateStrategy(m),
			},
			UseControlPlaneAsWorker: controlPlaneAsWorker,
		},
	}
	return mp
}

func validateMasterPoolCount(machinePool []*models.V1AzureMachinePoolConfigEntity) diag.Diagnostics {
	for _, machineConfig := range machinePool {
		if machineConfig.PoolConfig.IsControlPlane == true {
			if *machineConfig.PoolConfig.Size%2 == 0 {
				return diag.FromErr(fmt.Errorf("The master node pool size should be in an odd number. But it set to an even number '%d' in node name '%s' ", *machineConfig.PoolConfig.Size, *machineConfig.PoolConfig.Name))
			}
		}
	}
	return nil
}

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

func resourceClusterAws() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterAwsCreate,
		ReadContext:   resourceClusterAwsRead,
		UpdateContext: resourceClusterAwsUpdate,
		DeleteContext: resourceClusterDelete,
		Description:   "Resource for managing AWS clusters in Spectro Cloud through Palette.",

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
						"ssh_key_name": {
							Type:     schema.TypeString,
							ForceNew: true,
							Required: true,
						},
						"region": {
							Type:     schema.TypeString,
							ForceNew: true,
							Required: true,
						},
						"vpc_id": {
							Type:     schema.TypeString,
							ForceNew: true,
							Optional: true,
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
							Type:     schema.TypeString,
							Required: true,
						},
						"count": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Number of nodes in the machine pool.",
						},
						"instance_type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"capacity_type": {
							Type:         schema.TypeString,
							Default:      "on-demand",
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"on-demand", "spot"}, false),
							Description:  "Capacity type is an instance type,  can be 'on-demand' or 'spot'. Defaults to 'on-demand'.",
						},
						"max_price": {
							Type:     schema.TypeString,
							Optional: true,
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

func resourceClusterAwsCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	cluster := toAwsCluster(c, d)

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
	c := m.(*client.V1Client)

	var diags diag.Diagnostics
	//
	uid := d.Id()
	//
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

	return flattenCloudConfigAws(cluster.Spec.CloudConfigRef.UID, d, c)
}

func flattenCloudConfigAws(configUID string, d *schema.ResourceData, c *client.V1Client) diag.Diagnostics {
	if err := d.Set("cloud_config_id", configUID); err != nil {
		return diag.FromErr(err)
	}
	if config, err := c.GetCloudConfigAws(configUID); err != nil {
		return diag.FromErr(err)
	} else {
		mp := flattenMachinePoolConfigsAws(config.Spec.MachinePoolConfig)
		if err := d.Set("machine_pool", mp); err != nil {
			return diag.FromErr(err)
		}
	}

	return diag.Diagnostics{}
}

func flattenMachinePoolConfigsAws(machinePools []*models.V1AwsMachinePoolConfig) []interface{} {

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
		oi["count"] = int(machinePool.Size)
		flattenUpdateStrategy(machinePool.UpdateStrategy, oi)

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
		ois[i] = oi
	}

	return ois
}

func resourceClusterAwsUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	cloudConfigId := d.Get("cloud_config_id").(string)

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
			name := machinePoolResource["name"].(string)
			if name != "" {
				hash := resourceMachinePoolAwsHash(machinePoolResource)
				vpcId := d.Get("cloud_config").([]interface{})[0].(map[string]interface{})["vpc_id"]
				machinePool := toMachinePoolAws(machinePoolResource, vpcId.(string))

				var err error
				if oldMachinePool, ok := osMap[name]; !ok {
					log.Printf("Create machine pool %s", name)
					err = c.CreateMachinePoolAws(cloudConfigId, machinePool)
				} else if hash != resourceMachinePoolAwsHash(oldMachinePool) {
					log.Printf("Change in machine pool %s", name)
					err = c.UpdateMachinePoolAws(cloudConfigId, machinePool)
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
			if err := c.DeleteMachinePoolAws(cloudConfigId, name); err != nil {
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

	resourceClusterAwsRead(ctx, d, m)

	return diags
}

func toAwsCluster(c *client.V1Client, d *schema.ResourceData) *models.V1SpectroAwsClusterEntity {
	// gnarly, I know! =/
	cloudConfig := d.Get("cloud_config").([]interface{})[0].(map[string]interface{})

	cluster := &models.V1SpectroAwsClusterEntity{
		Metadata: &models.V1ObjectMeta{
			Name:   d.Get("name").(string),
			UID:    d.Id(),
			Labels: toTags(d),
		},
		Spec: &models.V1SpectroAwsClusterEntitySpec{
			CloudAccountUID: types.Ptr(d.Get("cloud_account_id").(string)),
			Profiles:        toProfiles(c, d),
			Policies:        toPolicies(d),
			CloudConfig: &models.V1AwsClusterConfig{
				SSHKeyName: cloudConfig["ssh_key_name"].(string),
				Region:     types.Ptr(cloudConfig["region"].(string)),
				VpcID:      cloudConfig["vpc_id"].(string),
			},
		},
	}

	//for _, machinePool := range d.Get("machine_pool").([]interface{}) {
	machinePoolConfigs := make([]*models.V1AwsMachinePoolConfigEntity, 0)
	for _, machinePool := range d.Get("machine_pool").(*schema.Set).List() {
		mp := toMachinePoolAws(machinePool, cluster.Spec.CloudConfig.VpcID)
		machinePoolConfigs = append(machinePoolConfigs, mp)
	}

	cluster.Spec.Machinepoolconfig = machinePoolConfigs
	cluster.Spec.ClusterConfig = toClusterConfig(d)

	return cluster
}

func toMachinePoolAws(machinePool interface{}, vpcId string) *models.V1AwsMachinePoolConfigEntity {
	m := machinePool.(map[string]interface{})

	labels := make([]string, 0)
	controlPlane := m["control_plane"].(bool)
	controlPlaneAsWorker := m["control_plane_as_worker"].(bool)
	if controlPlane {
		labels = append(labels, "master")
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
	mp := &models.V1AwsMachinePoolConfigEntity{
		CloudConfig: &models.V1AwsMachinePoolCloudConfigEntity{
			Azs:            azs,
			InstanceType:   types.Ptr(m["instance_type"].(string)),
			CapacityType:   &capacityType,
			RootDeviceSize: int64(m["disk_size_gb"].(int)),
			Subnets:        azSubnetsConfigs,
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

	if capacityType == "spot" {
		maxPrice := "0.0" // default value
		if m["max_price"] != nil && len(m["max_price"].(string)) > 0 {
			maxPrice = m["max_price"].(string)
		}

		mp.CloudConfig.SpotMarketOptions = &models.V1SpotMarketOptions{
			MaxPrice: maxPrice,
		}
	}
	return mp
}

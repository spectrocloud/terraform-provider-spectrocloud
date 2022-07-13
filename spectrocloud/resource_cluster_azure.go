package spectrocloud

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/gomi/pkg/ptr"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"
)

func resourceClusterAzure() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterAzureCreate,
		ReadContext:   resourceClusterAzureRead,
		UpdateContext: resourceClusterAzureUpdate,
		DeleteContext: resourceClusterDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

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
			},
			"cluster_profile": {
				Type:          schema.TypeList,
				Optional:      true,
				ConflictsWith: []string{"pack"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
							//ForceNew: true,
						},
						"pack": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"registry_uid": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"tag": {
										Type:     schema.TypeString,
										Required: true,
									},
									"values": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
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
				Type:     schema.TypeString,
				Computed: true,
			},
			"os_patch_on_boot": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"os_patch_schedule": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateOsPatchSchedule,
			},
			"os_patch_after": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateOsPatchOnDemandAfter,
			},
			"kubeconfig": {
				Type:     schema.TypeString,
				Computed: true,
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
					},
				},
			},
			"pack": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"registry_uid": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"tag": {
							Type:     schema.TypeString,
							Required: true,
						},
						"values": {
							Type:     schema.TypeString,
							Required: true,
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
						"control_plane": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
							//ForceNew: true,
						},
						"control_plane_as_worker": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,

							//ForceNew: true,
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
							//ForceNew: true,
						},
						"count": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"instance_type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"update_strategy": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "RollingUpdateScaleOut",
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
										Type:     schema.TypeInt,
										Required: true,
									},
									"type": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"azs": {
							Type:     schema.TypeSet,
							Required: true,
							Set:      schema.HashString,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"is_system_node_pool": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"os_type": {
							Type:     schema.TypeString,
							Optional: true,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								return false
							},
						},
					},
				},
			},
			"backup_policy": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"prefix": {
							Type:     schema.TypeString,
							Required: true,
						},
						"backup_location_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"schedule": {
							Type:     schema.TypeString,
							Required: true,
						},
						"expiry_in_hour": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"include_disks": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"include_cluster_resources": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"namespaces": {
							Type:     schema.TypeSet,
							Optional: true,
							Set:      schema.HashString,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"scan_policy": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"configuration_scan_schedule": {
							Type:     schema.TypeString,
							Required: true,
						},
						"penetration_scan_schedule": {
							Type:     schema.TypeString,
							Required: true,
						},
						"conformance_scan_schedule": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			//"cloud_config": {
			//	Type:     schema.TypeString,
			//	Required: true,
			//	//DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			//	//	return false
			//	//},
			//	//StateFunc: func(val interface{}) string {
			//	//	return strings.ToLower(val.(string))
			//	//},
			//},
		},
	}
}

func resourceClusterAzureCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	cluster := toAzureCluster(c, d)

	uid, err := c.CreateClusterAzure(cluster)
	if err != nil {
		return diag.FromErr(err)
	}

	diagnostics, isError := waitForClusterCreation(ctx, d, uid, diags, c)
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

	if err := d.Set("tags", flattenTags(cluster.Metadata.Labels)); err != nil {
		return diag.FromErr(err)
	}

	kubecfg, err := c.GetClusterKubeConfig(uid)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("kubeconfig", kubecfg); err != nil {
		return diag.FromErr(err)
	}

	if policy, err := c.GetClusterBackupConfig(d.Id()); err != nil {
		return diag.FromErr(err)
	} else if policy != nil && policy.Spec.Config != nil {
		if err := d.Set("backup_policy", flattenBackupPolicy(policy.Spec.Config)); err != nil {
			return diag.FromErr(err)
		}
	}

	if policy, err := c.GetClusterScanConfig(d.Id()); err != nil {
		return diag.FromErr(err)
	} else if policy != nil && policy.Spec.DriverSpec != nil {
		if err := d.Set("scan_policy", flattenScanPolicy(policy.Spec.DriverSpec)); err != nil {
			return diag.FromErr(err)
		}
	}

	return flattenCloudConfigAzure(cluster.Spec.CloudConfigRef.UID, d, c)
}

func flattenCloudConfigAzure(configUID string, d *schema.ResourceData, c *client.V1Client) diag.Diagnostics {
	d.Set("cloud_config_id", configUID)
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

		oi["control_plane"] = machinePool.IsControlPlane
		oi["control_plane_as_worker"] = machinePool.UseControlPlaneAsWorker
		oi["name"] = machinePool.Name
		oi["count"] = machinePool.Size
		if machinePool.UpdateStrategy.Type != "" {
			oi["update_strategy"] = machinePool.UpdateStrategy.Type
		} else {
			oi["update_strategy"] = "RollingUpdateScaleOut"
		}
		oi["instance_type"] = machinePool.InstanceType
		oi["is_system_node_pool"] = machinePool.IsSystemNodePool

		oi["azs"] = machinePool.Azs

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

	if d.HasChanges("cluster_profile") {
		if err := updateProfiles(c, d); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("backup_policy") {
		if err := updateBackupPolicy(c, d); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("scan_policy") {
		if err := updateScanPolicy(c, d); err != nil {
			return diag.FromErr(err)
		}
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
			CloudAccountUID: ptr.StringPtr(d.Get("cloud_account_id").(string)),
			Profiles:        toProfiles(c, d),
			Policies:        toPolicies(d),
			CloudConfig: &models.V1AzureClusterConfig{
				Location:       ptr.StringPtr(cloudConfig["region"].(string)),
				SSHKey:         ptr.StringPtr(cloudConfig["ssh_key"].(string)),
				SubscriptionID: ptr.StringPtr(cloudConfig["subscription_id"].(string)),
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
			IsControlPlane: controlPlane,
			Labels:         labels,
			Name:           ptr.StringPtr(m["name"].(string)),
			Size:           ptr.Int32Ptr(int32(m["count"].(int))),
			UpdateStrategy: &models.V1UpdateStrategy{
				Type: m["update_strategy"].(string),
			},
			UseControlPlaneAsWorker: controlPlaneAsWorker,
		},
	}
	return mp
}

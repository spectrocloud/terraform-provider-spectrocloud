package spectrocloud

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/gomi/pkg/ptr"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"
)

func resourceClusterAks() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterAksCreate,
		ReadContext:   resourceClusterAksRead,
		UpdateContext: resourceClusterAksUpdate,
		DeleteContext: resourceClusterDelete,

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
			},
			"cluster_profile_id": {
				Type:       schema.TypeString,
				Optional:   true,
				Deprecated: "Switch to cluster_profile",
			},
			"cluster_profile": {
				Type:          schema.TypeList,
				Optional:      true,
				ConflictsWith: []string{"cluster_profile_id", "pack"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
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
			"cloud_account_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cloud_config_id": {
				Type:     schema.TypeString,
				Computed: true,
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
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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
		},
	}
}

func resourceClusterAksCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	cluster := toAksCluster(d)

	uid, err := c.CreateClusterAks(cluster)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(uid)

	stateConf := &resource.StateChangeConf{
		Pending:    resourceClusterCreatePendingStates,
		Target:     []string{"Running"},
		Refresh:    resourceClusterStateRefreshFunc(c, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutCreate) - 1*time.Minute,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	// Wait, catching any errors
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	resourceClusterAksRead(ctx, d, m)

	return diags
}

//goland:noinspection GoUnhandledErrorResult
func resourceClusterAksRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	configUID := cluster.Spec.CloudConfigRef.UID
	d.Set("cloud_config_id", configUID)

	if err := d.Set("tags", flattenTags(cluster.Metadata.Labels)); err != nil {
		return diag.FromErr(err)
	}

	var config *models.V1AzureCloudConfig
	if config, err = c.GetCloudConfigAks(configUID); err != nil {
		return diag.FromErr(err)
	}

	// Update the kubeconfig
	kubecfg, err := c.GetClusterKubeConfig(uid)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("kubeconfig", kubecfg); err != nil {
		return diag.FromErr(err)
	}

	mp := flattenMachinePoolConfigsAks(config.Spec.MachinePoolConfig)
	if err := d.Set("machine_pool", mp); err != nil {
		return diag.FromErr(err)
	}

	//read backup policy and scan policy
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
	return diags
}

func flattenMachinePoolConfigsAks(machinePools []*models.V1AzureMachinePoolConfig) []interface{} {
	if machinePools == nil {
		return make([]interface{}, 0)
	}

	ois := make([]interface{}, 0)
	for _, machinePool := range machinePools {
		oi := make(map[string]interface{})

		if *machinePool.IsControlPlane {
			continue
		}

		oi["name"] = machinePool.Name
		oi["count"] = int(machinePool.Size)
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
	_ = d.Get("machine_pool")
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
			}
			if err != nil {
				return diag.FromErr(err)
			}
			delete(osMap, name)
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

	resourceClusterAksRead(ctx, d, m)

	return diags
}

func toAksCluster(d *schema.ResourceData) *models.V1SpectroAzureClusterEntity {
	cloudConfig := d.Get("cloud_config").([]interface{})[0].(map[string]interface{})
	cluster := &models.V1SpectroAzureClusterEntity{
		Metadata: &models.V1ObjectMeta{
			Name:   d.Get("name").(string),
			UID:    d.Id(),
			Labels: toTags(d),
		},
		Spec: &models.V1SpectroAzureClusterEntitySpec{
			CloudAccountUID: ptr.StringPtr(d.Get("cloud_account_id").(string)),
			Profiles:        toProfiles(d),
			Policies:        toPolicies(d),
			CloudConfig: &models.V1AzureClusterConfig{
				ControlPlaneSubnet: nil,
				Location:           ptr.StringPtr(cloudConfig["region"].(string)),
				ResourceGroup:      cloudConfig["resource_group"].(string),
				SSHKey:             ptr.StringPtr(cloudConfig["ssh_key"].(string)),
				SubscriptionID:     ptr.StringPtr(cloudConfig["subscription_id"].(string)),
			},
		},
	}

	machinePoolConfigs := make([]*models.V1AzureMachinePoolConfigEntity, 0)
	for _, machinePool := range d.Get("machine_pool").([]interface{}) {
		mp := toMachinePoolAks(machinePool)
		machinePoolConfigs = append(machinePoolConfigs, mp)
	}
	cluster.Spec.Machinepoolconfig = machinePoolConfigs
	return cluster
}

func toMachinePoolAks(machinePool interface{}) *models.V1AzureMachinePoolConfigEntity {
	m := machinePool.(map[string]interface{})

	labels := make([]string, 0)
	controlPlane, _ := m["control_plane"].(bool)
	if controlPlane {
		labels = append(labels, "master")
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
			IsControlPlane: controlPlane,
			Labels:         labels,
			Name:           ptr.StringPtr(m["name"].(string)),
			Size:           ptr.Int32Ptr(int32(m["count"].(int))),
			MinSize:        int32(m["count"].(int)),
			MaxSize:        int32(m["count"].(int)),
		},
	}

	return mp
}

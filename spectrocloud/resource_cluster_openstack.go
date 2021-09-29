package spectrocloud

import (
	"context"
	"log"
	"sort"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/gomi/pkg/ptr"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"
)

func resourceClusterOpenStack() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterOpenStackCreate,
		ReadContext:   resourceClusterOpenStackRead,
		UpdateContext: resourceClusterOpenStackUpdate,
		DeleteContext: resourceClusterDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(180 * time.Minute),
			Update: schema.DefaultTimeout(180 * time.Minute),
			Delete: schema.DefaultTimeout(180 * time.Minute),
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
							Type:     schema.TypeString,
							Required: true,
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
						"control_plane": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"control_plane_as_worker": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"count": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"update_strategy": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "RollingUpdateScaleOut",
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

func resourceClusterOpenStackCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	cluster := toOpenStackCluster(d)

	uid, err := c.CreateClusterOpenStack(cluster)
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

	resourceClusterOpenStackRead(ctx, d, m)

	return diags
}

func toOpenStackCluster(d *schema.ResourceData) *models.V1SpectroOpenStackClusterEntity {

	cloudConfig := d.Get("cloud_config").([]interface{})[0].(map[string]interface{})

	cluster := &models.V1SpectroOpenStackClusterEntity{
		Metadata: &models.V1ObjectMeta{
			Name:   d.Get("name").(string),
			UID:    d.Id(),
			Labels: toTags(d),
		},
		Spec: &models.V1SpectroOpenStackClusterEntitySpec{
			CloudAccountUID: ptr.StringPtr(d.Get("cloud_account_id").(string)),
			Profiles:        toProfiles(d),
			Policies:        toPolicies(d),
			CloudConfig: &models.V1OpenStackClusterConfig{
				Region:     cloudConfig["region"].(string),
				SSHKeyName: cloudConfig["ssh_key"].(string),
				Domain: &models.V1OpenStackResource {
					ID: cloudConfig["domain"].(string),
					Name: cloudConfig["domain"].(string),
				},
				Network: &models.V1OpenStackResource {
					ID: cloudConfig["network_id"].(string),
				},
				Project: &models.V1OpenStackResource {
					Name: cloudConfig["project"].(string),
				},
				Subnet: &models.V1OpenStackResource {
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

	for _, machinePool := range d.Get("machine_pool").([]interface{}) {
		mp := toMachinePoolOpenStack(machinePool)
		machinePoolConfigs = append(machinePoolConfigs, mp)
	}

	// sort
	sort.SliceStable(machinePoolConfigs, func(i, j int) bool {
		return machinePoolConfigs[i].PoolConfig.IsControlPlane
	})

	cluster.Spec.Machinepoolconfig = machinePoolConfigs
	cluster.Spec.ClusterConfig = toClusterConfig(d)

	return cluster
}


//goland:noinspection GoUnhandledErrorResult
func resourceClusterOpenStackRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	var config *models.V1OpenStackCloudConfig
	if config, err = c.GetCloudConfigOpenStack(configUID); err != nil {
		return diag.FromErr(err)
	}

	kubecfg, err := c.GetClusterKubeConfig(uid)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("kubeconfig", kubecfg); err != nil {
		return diag.FromErr(err)
	}

	mp := flattenMachinePoolConfigsOpenStack(config.Spec.MachinePoolConfig)
	if err := d.Set("machine_pool", mp); err != nil {
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

	return diags
}

func flattenMachinePoolConfigsOpenStack(machinePools []*models.V1OpenStackMachinePoolConfig) []interface{} {

	if machinePools == nil {
		return make([]interface{}, 0)
	}

	ois := make([]interface{}, 0)

	for _, machinePool := range machinePools {
		oi := make(map[string]interface{})

		oi["control_plane"] = machinePool.IsControlPlane
		oi["control_plane_as_worker"] = machinePool.UseControlPlaneAsWorker
		oi["name"] = machinePool.Name
		oi["count"] = int(machinePool.Size)
		oi["update_strategy"] = machinePool.UpdateStrategy.Type

		oi["subnet_id"] = machinePool.Subnet.ID
		oi["azs"] = machinePool.Azs
		oi["instance_type"] = machinePool.FlavorConfig.Name

		ois = append(ois, oi)
	}

	return ois
}

func resourceClusterOpenStackUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
			hash := resourceMachinePoolOpenStackHash(machinePoolResource)

			machinePool := toMachinePoolOpenStack(machinePoolResource)

			var err error
			if oldMachinePool, ok := osMap[name]; !ok {
				log.Printf("Create machine pool %s", name)
				err = c.CreateMachinePoolOpenStack(cloudConfigId, machinePool)
			} else if hash != resourceMachinePoolOpenStackHash(oldMachinePool) {
				log.Printf("Change in machine pool %s", name)
				err = c.UpdateMachinePoolOpenStack(cloudConfigId, machinePool)
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
			if err := c.DeleteMachinePoolOpenStack(cloudConfigId, name); err != nil {
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

	resourceClusterOpenStackRead(ctx, d, m)

	return diags
}


func toMachinePoolOpenStack(machinePool interface{}) *models.V1OpenStackMachinePoolConfigEntity {
	m := machinePool.(map[string]interface{})

	labels := make([]string, 0)
	controlPlane := m["control_plane"].(bool)
	controlPlaneAsWorker := m["control_plane_as_worker"].(bool)
	if controlPlane {
		labels = append(labels, "master")
	}

	azs := make([]string, 0)
	for _, val := range m["azs"].(*schema.Set).List() {
		azs = append(azs, val.(string))
	}

	mp := &models.V1OpenStackMachinePoolConfigEntity{
		CloudConfig: &models.V1OpenStackMachinePoolCloudConfigEntity{
			Azs: azs,
			Subnet: &models.V1OpenStackResource {
				ID: m["subnet_id"].(string),
			},
			FlavorConfig: &models.V1OpenstackFlavorConfig {
				Name: ptr.StringPtr(m["instance_type"].(string)),
			},
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



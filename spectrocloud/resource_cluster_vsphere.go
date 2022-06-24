package spectrocloud

import (
	"context"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/gomi/pkg/ptr"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"
)

func resourceClusterVsphere() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterVsphereCreate,
		ReadContext:   resourceClusterVsphereRead,
		UpdateContext: resourceClusterVsphereUpdate,
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
						"datacenter": {
							Type:     schema.TypeString,
							Required: true,
						},
						"folder": {
							Type:     schema.TypeString,
							Required: true,
						},

						"ssh_key": {
							Type:     schema.TypeString,
							Required: true,
						},

						"static_ip": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},

						// DHCP Properties
						"network_type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"network_search_domain": {
							Type:     schema.TypeString,
							Optional: true,
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
				Set:      resourceMachinePoolVsphereHash,
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
						"taints": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"key": {
										Type:     schema.TypeString,
										Required: true,
									},
									"value": {
										Type:     schema.TypeString,
										Required: true,
									},
									"effect": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
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
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"disk_size_gb": {
										Type:     schema.TypeInt,
										Required: true,
									},
									"memory_mb": {
										Type:     schema.TypeInt,
										Required: true,
									},
									"cpu": {
										Type:     schema.TypeInt,
										Required: true,
									},
								},
							},
						},
						"placement": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"cluster": {
										Type:     schema.TypeString,
										Required: true,
									},
									"resource_pool": {
										Type:     schema.TypeString,
										Required: true,
									},
									"datastore": {
										Type:     schema.TypeString,
										Required: true,
									},
									"network": {
										Type:     schema.TypeString,
										Required: true,
									},
									"static_ip_pool_id": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
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
			"cluster_rbac_binding": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"namespace": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"role": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"subjects": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:     schema.TypeString,
										Required: true,
									},
									"name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"namespace": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"namespaces": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"resource_allocation": {
							Type:     schema.TypeMap,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}
}

func resourceClusterVsphereCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	cluster := toVsphereCluster(c, d)

	uid, err := c.CreateClusterVsphere(cluster)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(uid)

	if _, found := toTags(d)["skip_completion"]; found {
		return diags
	}

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

	resourceClusterVsphereRead(ctx, d, m)

	return diags
}

//goland:noinspection GoUnhandledErrorResult
func resourceClusterVsphereRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
	if err := d.Set("cloud_config_id", configUID); err != nil {
		return diag.FromErr(err)
	}
	if config, err := c.GetCloudConfigVsphere(configUID); err != nil {
		return diag.FromErr(err)
	} else {
		mp := flattenMachinePoolConfigsVsphere(config.Spec.MachinePoolConfig)
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

func flattenCloudConfigVsphere(configUID string, d *schema.ResourceData, c *client.V1Client) diag.Diagnostics {
	d.Set("cloud_config_id", configUID)
	if config, err := c.GetCloudConfigVsphere(configUID); err != nil {
		return diag.FromErr(err)
	} else {
		cloudConfig, err := c.GetVsphereClouldConfigValues(configUID)
		if err != nil {
			return diag.FromErr(err)
		}
		cloudConfigFlatten := flattenClusterConfigsVsphere(cloudConfig)
		if err := d.Set("cloud_config", cloudConfigFlatten); err != nil {
			return diag.FromErr(err)
		}
		mp := flattenMachinePoolConfigsVsphere(config.Spec.MachinePoolConfig)
		if err := d.Set("machine_pool", mp); err != nil {
			return diag.FromErr(err)
		}
	}

	return diag.Diagnostics{}
}

func flattenClusterConfigsVsphere(cloudConfig *models.V1VsphereCloudConfig) interface{} {

	cloudConfigFlatten := make([]interface{}, 0)
	if cloudConfig == nil {
		return cloudConfigFlatten
	}

	ret := make(map[string]interface{})

	cpEndpoint := cloudConfig.Spec.ClusterConfig.ControlPlaneEndpoint
	placement := cloudConfig.Spec.ClusterConfig.Placement
	ret["datacenter"] = placement.Datacenter
	ret["folder"] = placement.Folder
	ret["ssh_key"] = strings.TrimSpace(cloudConfig.Spec.ClusterConfig.SSHKeys[0])
	ret["static_ip"] = cloudConfig.Spec.ClusterConfig.StaticIP

	if cpEndpoint.Type != "" {
		ret["network_type"] = cpEndpoint.Type
	}

	if cpEndpoint.DdnsSearchDomain != "" {
		ret["network_search_domain"] = cpEndpoint.DdnsSearchDomain
	}

	cloudConfigFlatten = append(cloudConfigFlatten, ret)

	return cloudConfigFlatten
}

func flattenMachinePoolConfigsVsphere(machinePools []*models.V1VsphereMachinePoolConfig) []interface{} {

	if machinePools == nil {
		return make([]interface{}, 0)
	}

	ois := make([]interface{}, len(machinePools))

	for i, machinePool := range machinePools {
		oi := make(map[string]interface{})

		if machinePool.AdditionalLabels == nil || len(machinePool.AdditionalLabels) == 0 {
			oi["additional_labels"] = make(map[string]interface{})
		} else {
			oi["additional_labels"] = machinePool.AdditionalLabels
		}

		taints := flattenClusterTaints(machinePool.Taints)
		if len(taints) > 0 {
			oi["taints"] = taints
		}

		oi["control_plane"] = machinePool.IsControlPlane
		oi["control_plane_as_worker"] = machinePool.UseControlPlaneAsWorker
		oi["name"] = machinePool.Name
		oi["count"] = machinePool.Size
		if machinePool.UpdateStrategy.Type != "" {
			oi["update_strategy"] = machinePool.UpdateStrategy.Type
		} else {
			oi["update_strategy"] = "RollingUpdateScaleOut"
		}

		if machinePool.InstanceType != nil {
			s := make(map[string]interface{})
			s["disk_size_gb"] = int(*machinePool.InstanceType.DiskGiB)
			s["memory_mb"] = int(*machinePool.InstanceType.MemoryMiB)
			s["cpu"] = int(*machinePool.InstanceType.NumCPUs)

			oi["instance_type"] = []interface{}{s}
		}

		placements := make([]interface{}, len(machinePool.Placements))
		for j, p := range machinePool.Placements {
			pj := make(map[string]interface{})
			pj["id"] = p.UID
			pj["cluster"] = p.Cluster
			pj["resource_pool"] = p.ResourcePool
			pj["datastore"] = p.Datastore
			pj["network"] = p.Network.NetworkName

			poolID := ""
			if p.Network.ParentPoolRef != nil {
				poolID = p.Network.ParentPoolRef.UID
			}
			pj["static_ip_pool_id"] = poolID

			placements[j] = pj
		}
		oi["placement"] = placements

		ois[i] = oi
	}

	return ois
}

func resourceClusterVsphereUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	cloudConfigId := d.Get("cloud_config_id").(string)

	if d.HasChange("cloud_config") {
		cloudConfig := toCloudConfigUpdate(d.Get("cloud_config").([]interface{})[0].(map[string]interface{}))
		c.UpdateVsphereCloudConfigValues(cloudConfigId, cloudConfig)
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

		for _, mp := range ns.List() {
			machinePoolResource := mp.(map[string]interface{})
			name := machinePoolResource["name"].(string)
			hash := resourceMachinePoolVsphereHash(machinePoolResource)

			machinePool := toMachinePoolVsphere(machinePoolResource)

			var err error
			if oldMachinePool, ok := osMap[name]; !ok {
				log.Printf("Create machine pool %s", name)
				err = c.CreateMachinePoolVsphere(cloudConfigId, machinePool)
			} else if hash != resourceMachinePoolVsphereHash(oldMachinePool) {
				log.Printf("Change in machine pool %s", name)
				oldMachinePool := toMachinePoolVsphere(oldMachinePool)
				oldPlacements := oldMachinePool.CloudConfig.Placements

				// set the placement ids
				for i, p := range machinePool.CloudConfig.Placements {
					if len(oldPlacements) > i {
						p.UID = oldPlacements[i].UID
					}
				}

				err = c.UpdateMachinePoolVsphere(cloudConfigId, machinePool)
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
			if err := c.DeleteMachinePoolVsphere(cloudConfigId, name); err != nil {
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

	resourceClusterVsphereRead(ctx, d, m)

	return diags
}

func toVsphereCluster(c *client.V1Client, d *schema.ResourceData) *models.V1SpectroVsphereClusterEntity {
	// gnarly, I know! =/
	cloudConfig := d.Get("cloud_config").([]interface{})[0].(map[string]interface{})
	//clientSecret := strfmt.Password(d.Get("azure_client_secret").(string))

	cluster := &models.V1SpectroVsphereClusterEntity{
		Metadata: &models.V1ObjectMeta{
			Name:   d.Get("name").(string),
			UID:    d.Id(),
			Labels: toTags(d),
		},
		Spec: &models.V1SpectroVsphereClusterEntitySpec{
			CloudAccountUID: d.Get("cloud_account_id").(string),
			Profiles:        toProfiles(c, d),
			Policies:        toPolicies(d),
			CloudConfig:     toCloudConfigCreate(cloudConfig),
		},
	}

	machinePoolConfigs := make([]*models.V1VsphereMachinePoolConfigEntity, 0)
	for _, machinePool := range d.Get("machine_pool").(*schema.Set).List() {
		mp := toMachinePoolVsphere(machinePool)
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

func toCloudConfigCreate(cloudConfig map[string]interface{}) *models.V1VsphereClusterConfigEntity {
	staticIP := cloudConfig["static_ip"].(bool)

	V1VsphereClusterConfigEntity := &models.V1VsphereClusterConfigEntity{
		NtpServers: nil,
		Placement: &models.V1VspherePlacementConfigEntity{
			Datacenter: cloudConfig["datacenter"].(string),
			Folder:     cloudConfig["folder"].(string),
		},
		// ssh key trim is needed for UI to display key correctly.
		SSHKeys:  []string{strings.TrimSpace(cloudConfig["ssh_key"].(string))},
		StaticIP: staticIP,
	}

	if !staticIP {
		V1VsphereClusterConfigEntity.ControlPlaneEndpoint = &models.V1ControlPlaneEndPoint{
			DdnsSearchDomain: cloudConfig["network_search_domain"].(string),
			Type:             cloudConfig["network_type"].(string),
		}
	}

	return V1VsphereClusterConfigEntity
}

func toCloudConfigUpdate(cloudConfig map[string]interface{}) *models.V1VsphereCloudClusterConfigEntity {
	return &models.V1VsphereCloudClusterConfigEntity{
		toCloudConfigCreate(cloudConfig),
	}
}

func toMachinePoolVsphere(machinePool interface{}) *models.V1VsphereMachinePoolConfigEntity {
	m := machinePool.(map[string]interface{})

	labels := make([]string, 0)
	controlPlane := m["control_plane"].(bool)
	controlPlaneAsWorker := m["control_plane_as_worker"].(bool)
	if controlPlane {
		labels = append(labels, "master")
	}

	placements := make([]*models.V1VspherePlacementConfigEntity, 0)
	for _, pos := range m["placement"].([]interface{}) {
		p := pos.(map[string]interface{})
		poolID := p["static_ip_pool_id"].(string)
		staticIP := false
		if len(poolID) > 0 {
			staticIP = true
		}

		placements = append(placements, &models.V1VspherePlacementConfigEntity{
			UID:          p["id"].(string),
			Cluster:      p["cluster"].(string),
			ResourcePool: p["resource_pool"].(string),
			Datastore:    p["datastore"].(string),
			Network: &models.V1VsphereNetworkConfigEntity{
				NetworkName:   ptr.StringPtr(p["network"].(string)),
				ParentPoolUID: poolID,
				StaticIP:      staticIP,
			},
		})

	}

	ins := m["instance_type"].([]interface{})[0].(map[string]interface{})
	instanceType := models.V1VsphereInstanceType{
		DiskGiB:   ptr.Int32Ptr(int32(ins["disk_size_gb"].(int))),
		MemoryMiB: ptr.Int64Ptr(int64(ins["memory_mb"].(int))),
		NumCPUs:   ptr.Int32Ptr(int32(ins["cpu"].(int))),
	}

	mp := &models.V1VsphereMachinePoolConfigEntity{
		CloudConfig: &models.V1VsphereMachinePoolCloudConfigEntity{
			Placements:   placements,
			InstanceType: &instanceType,
		},
		PoolConfig: &models.V1MachinePoolConfigEntity{
			AdditionalLabels: toAdditionalNodePoolLabels(m),
			Taints:           toClusterTaints(m),
			IsControlPlane:   controlPlane,
			Labels:           labels,
			Name:             ptr.StringPtr(m["name"].(string)),
			Size:             ptr.Int32Ptr(int32(m["count"].(int))),
			UpdateStrategy: &models.V1UpdateStrategy{
				Type: m["update_strategy"].(string),
			},
			UseControlPlaneAsWorker: controlPlaneAsWorker,
		},
	}
	return mp
}

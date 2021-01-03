package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/gomi/pkg/ptr"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"
	"log"
	"time"
)

func resourceClusterVsphere() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterVsphereCreate,
		ReadContext:   resourceClusterVsphereRead,
		UpdateContext: resourceClusterVsphereUpdate,
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
			"cluster_profile_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
						"datacenter": {
							Type:     schema.TypeString,
							Required: true,
						},
						"folder": {
							Type:     schema.TypeString,
							Required: true,
						},
						"network_type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"network_search_domain": {
							Type:     schema.TypeString,
							Required: true,
							// TODO(saamalik) required only when not static IP

						},
						"ssh_key": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"pack" : {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      resourcePackHash,
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
				Type:     schema.TypeSet,
				Required: true,
				Set : resourceMachinePoolVsphereHash,
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
								},
							},
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

func resourceClusterVsphereCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1alpha1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	cluster := toVsphereCluster(d)

	uid, err := c.CreateClusterVsphere(cluster)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(uid)

	stateConf := &resource.StateChangeConf{
		Pending:    resourceClusterCreatePendingStates,
		Target:     []string{"Running"},
		Refresh:    resourceClusterStateRefreshFunc(c, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	// Wait, catching any errors
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		diag.FromErr(err)
	}

	resourceClusterVsphereRead(ctx, d, m)

	return diags
}

//goland:noinspection GoUnhandledErrorResult
func resourceClusterVsphereRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1alpha1Client)

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

	var config *models.V1alpha1VsphereCloudConfig
	if config, err = c.GetCloudConfigVsphere(configUID); err != nil {
		return diag.FromErr(err)
	}


	mp := flattenMachinePoolConfigsVsphere(config.Spec.MachinePoolConfig)
	if err := d.Set("machine_pool", mp); err != nil {
		return diag.FromErr(err)
	}

	// Update the kubeconfig
	kubeconfig, err := c.GetClusterKubeConfig(uid)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("kubeconfig", kubeconfig); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func flattenMachinePoolConfigsVsphere(machinePools []*models.V1alpha1VsphereMachinePoolConfig) []interface{} {

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
		oi["update_strategy"] = machinePool.UpdateStrategy.Type

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
			pj["cluster"] = p.Cluster
			pj["resource_pool"] = p.ResourcePool
			pj["datastore"] = p.Datastore
			pj["network"] = p.Network.NetworkName
			// TODO(saamalik) Static ip

			placements[j] = pj
		}
		oi["placement"] = placements

		ois[i] = oi
	}

	return ois
}

func resourceClusterVsphereUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1alpha1Client)

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
			hash := resourceMachinePoolVsphereHash(machinePoolResource)

			machinePool := toMachinePoolVsphere(machinePoolResource)

			var err error
			if oldMachinePool, ok := osMap[name]; !ok {
				log.Printf("Create machine pool %s", name)
				err = c.CreateMachinePoolVsphere(cloudConfigId, machinePool)
			} else if hash != resourceMachinePoolVsphereHash(oldMachinePool) {
				log.Printf("Change in machine pool %s", name)
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

	if d.HasChanges("pack") {
		log.Printf("Updating packs")
		cluster := toVsphereCluster(d)
		if err := c.UpdateClusterVsphere(cluster); err != nil {
			return diag.FromErr(err)
		}
	}

	resourceClusterVsphereRead(ctx, d, m)

	return diags
}

func toVsphereCluster(d *schema.ResourceData) *models.V1alpha1SpectroVsphereClusterEntity {
	// gnarly, I know! =/
	cloudConfig := d.Get("cloud_config").([]interface{})[0].(map[string]interface{})
	//clientSecret := strfmt.Password(d.Get("azure_client_secret").(string))

	cluster := &models.V1alpha1SpectroVsphereClusterEntity{
		Metadata: &models.V1ObjectMeta{
			Name: d.Get("name").(string),
			UID:  d.Id(),
		},
		Spec: &models.V1alpha1SpectroVsphereClusterEntitySpec{
			CloudAccountUID: ptr.StringPtr(d.Get("cloud_account_id").(string)),
			ProfileUID:      ptr.StringPtr(d.Get("cluster_profile_id").(string)),
			CloudConfig: &models.V1alpha1VsphereClusterConfigEntity{
				ControlPlaneEndpoint: &models.V1alpha1ControlPlaneEndPoint{
					DdnsSearchDomain: cloudConfig["network_search_domain"].(string),
					Type:             cloudConfig["network_type"].(string),
				},
				NtpServers: nil,
				Placement: &models.V1alpha1VspherePlacementConfigEntity{
					Datacenter:          cloudConfig["datacenter"].(string),
					Folder:              cloudConfig["folder"].(string),
					//ImageTemplateFolder: "",
					//Network: &models.V1alpha1VsphereNetworkConfigEntity{
					//	// TODO(saamalik)
					//	StaticIP:      false,
					//},
					//UID:          "",
				},
				SSHKeys:  []string{cloudConfig["ssh_key"].(string)},
				// TODO(saamalik)
				StaticIP: false,
			},
		},
	}

	machinePoolConfigs := make([]*models.V1alpha1VsphereMachinePoolConfigEntity, 0)
	for _, machinePool := range d.Get("machine_pool").(*schema.Set).List() {
		mp := toMachinePoolVsphere(machinePool)
		machinePoolConfigs = append(machinePoolConfigs, mp)
	}

	cluster.Spec.Machinepoolconfig = machinePoolConfigs

	packValues := make([]*models.V1alpha1PackValuesEntity, 0)
	for _, pack := range d.Get("pack").(*schema.Set).List() {
		p := toPack(pack)
		packValues = append(packValues, p)
	}
	cluster.Spec.PackValues = packValues

	return cluster
}

func toMachinePoolVsphere(machinePool interface{}) *models.V1alpha1VsphereMachinePoolConfigEntity {
	m := machinePool.(map[string]interface{})

	labels := make([]string, 0)
	controlPlane := m["control_plane"].(bool)
	controlPlaneAsWorker := m["control_plane_as_worker"].(bool)
	if controlPlane {
		labels = append(labels, "master")
	}

	placements := make([]*models.V1alpha1VspherePlacementConfigEntity, 0)
	for _, pos := range m["placement"].([]interface{}) {
		p := pos.(map[string]interface{})
		placements = append(placements, &models.V1alpha1VspherePlacementConfigEntity{
			Cluster:             p["cluster"].(string),
			ResourcePool:        p["resource_pool"].(string),
			Datastore:           p["datastore"].(string),
			Network:             &models.V1alpha1VsphereNetworkConfigEntity{
				NetworkName: ptr.StringPtr(p["network"].(string)),
			},
		})

	}

	ins := m["instance_type"].([]interface{})[0].(map[string]interface{})
	instanceType := models.V1alpha1VsphereInstanceType{
		DiskGiB:   ptr.Int32Ptr(int32(ins["disk_size_gb"].(int))),
		MemoryMiB: ptr.Int64Ptr(int64(ins["memory_mb"].(int))),
		NumCPUs:   ptr.Int32Ptr(int32(ins["cpu"].(int))),
	}

	mp := &models.V1alpha1VsphereMachinePoolConfigEntity{
		CloudConfig: &models.V1alpha1VsphereMachinePoolCloudConfigEntity{
			Placements: placements,
			InstanceType: &instanceType,
		},
		PoolConfig: &models.V1alpha1MachinePoolConfigEntity{
			IsControlPlane: controlPlane,
			Labels:         labels,
			Name:           ptr.StringPtr(m["name"].(string)),
			Size:           ptr.Int32Ptr(int32(m["count"].(int))),
			UpdateStrategy: &models.V1alpha1UpdateStrategy{
				Type: m["update_strategy"].(string),
			},
			UseControlPlaneAsWorker: controlPlaneAsWorker,
		},
	}
	return mp
}


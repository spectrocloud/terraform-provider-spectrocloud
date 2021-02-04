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

func resourceClusterAws() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterAwsCreate,
		ReadContext:   resourceClusterAwsRead,
		UpdateContext: resourceClusterAwsUpdate,
		DeleteContext: resourceClusterDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

		SchemaVersion: 2,
		Schema: map[string]*schema.Schema{
			name: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			cluster_prrofile_id: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			cloud_account_id: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			os_patch_on_boot: {
				Type:     schema.TypeBool,
				Optional: true,
			},
			os_patch_schedule: {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateOsPatchSchedule,
			},
			os_patch_after: {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateOsPatchOnDemandAfter,
			},
			cloud_config_id: {
				Type:     schema.TypeString,
				Computed: true,
			},
			kubeconfig: {
				Type:     schema.TypeString,
				Computed: true,
			},
			cloud_config: {
				Type:     schema.TypeList,
				ForceNew: true,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ssh_key_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"region": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			pack: {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      resourcePackHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						name: {
							Type:     schema.TypeString,
							Required: true,
						},
						tag: {
							Type:     schema.TypeString,
							Required: true,
						},
						values: {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			machine_pool: {
				Type:     schema.TypeSet,
				Required: true,
				Set:      resourceMachinePoolAwsHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						control_plane: {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
							//ForceNew: true,
						},
						control_plane_as_worker: {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,

							//ForceNew: true,
						},
						name: {
							Type:     schema.TypeString,
							Required: true,
							//ForceNew: true,
						},
						count: {
							Type:     schema.TypeInt,
							Required: true,
						},
						instance_type: {
							Type:     schema.TypeString,
							Required: true,
						},
						update_strategy: {
							Type:     schema.TypeString,
							Optional: true,
							Default:  rolling_update_scale_out,
						},
						disk_size_in_gb: {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  65,
						},
						availability_zones: {
							Type:     schema.TypeSet,
							Required: true,
							MinItems: 1,
							Set:      schema.HashString,
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

func resourceClusterAwsCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1alpha1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	cluster := toAwsCluster(d)

	uid, err := c.CreateClusterAws(cluster)
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

	resourceClusterAwsRead(ctx, d, m)

	return diags
}

//goland:noinspection GoUnhandledErrorResult
func resourceClusterAwsRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1alpha1Client)

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

	configUID := cluster.Spec.CloudConfigRef.UID
	d.Set("cloud_config_id", configUID)

	var config *models.V1alpha1AwsCloudConfig
	if config, err = c.GetCloudConfigAws(configUID); err != nil {
		return diag.FromErr(err)
	}

	mp := flattenMachinePoolConfigsAws(config.Spec.MachinePoolConfig)
	if err := d.Set("machine_pool", mp); err != nil {
		return diag.FromErr(err)
	}

	// Update the kubeconfig
	if cluster.Status != nil && cluster.Status.ClusterImport != nil && cluster.Status.ClusterImport.IsBrownfield {
		if err := d.Set("cluster_import_manifest_url", cluster.Status.ClusterImport.ImportLink); err != nil {
			return diag.FromErr(err)
		}
	} else {
		kubeconfig, err := c.GetClusterKubeConfig(uid)
		if err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("kubeconfig", kubeconfig); err != nil {
			return diag.FromErr(err)
		}
	}
	return diags
}

func flattenMachinePoolConfigsAws(machinePools []*models.V1alpha1AwsMachinePoolConfig) []interface{} {

	if machinePools == nil {
		return make([]interface{}, 0)
	}

	ois := make([]interface{}, len(machinePools))

	for i, machinePool := range machinePools {
		oi := make(map[string]interface{})

		oi[control_plane] = machinePool.IsControlPlane
		oi[control_plane_as_worker] = machinePool.UseControlPlaneAsWorker
		oi["name"] = machinePool.Name
		oi["count"] = int(machinePool.Size)
		oi["update_strategy"] = machinePool.UpdateStrategy.Type
		oi["instance_type"] = machinePool.InstanceType

		oi["disk_size_gb"] = int(machinePool.RootDeviceSize)

		oi["azs"] = machinePool.Azs

		ois[i] = oi
	}

	return ois
}

func resourceClusterAwsUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
			hash := resourceMachinePoolAwsHash(machinePoolResource)

			machinePool := toMachinePoolAws(machinePoolResource)

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

	if d.HasChanges("pack") {
		log.Printf("Updating packs")
		cluster := toAwsCluster(d)
		if err := c.UpdateClusterAws(cluster); err != nil {
			return diag.FromErr(err)
		}
	}

	resourceClusterAwsRead(ctx, d, m)

	return diags
}

func toAwsCluster(d *schema.ResourceData) *models.V1alpha1SpectroAwsClusterEntity {
	// gnarly, I know! =/
	cloudConfig := d.Get("cloud_config").([]interface{})[0].(map[string]interface{})

	cluster := &models.V1alpha1SpectroAwsClusterEntity{
		Metadata: &models.V1ObjectMeta{
			Name: d.Get("name").(string),
			UID:  d.Id(),
		},
		Spec: &models.V1alpha1SpectroAwsClusterEntitySpec{
			CloudAccountUID: ptr.StringPtr(d.Get(cloud_account_id).(string)),
			CloudConfig: &models.V1alpha1AwsClusterConfig{
				SSHKeyName: cloudConfig["ssh_key_name"].(string),
				Region:     ptr.StringPtr(cloudConfig["region"].(string)),
			},
		},
	}

	machinePoolConfigs := make([]*models.V1alpha1AwsMachinePoolConfigEntity, 0)
	for _, machinePool := range d.Get("machine_pool").(*schema.Set).List() {
		mp := toMachinePoolAws(machinePool)
		machinePoolConfigs = append(machinePoolConfigs, mp)
	}

	cluster.Spec.Machinepoolconfig = machinePoolConfigs

	packValues := make([]*models.V1alpha1PackValuesEntity, 0)
	for _, pack := range d.Get("pack").(*schema.Set).List() {
		p := toPack(pack)
		packValues = append(packValues, p)
	}
	cluster.Spec.PackValues = packValues
	cluster.Spec.ClusterConfig = getClusterConfig(d)
	return cluster
}

func toMachinePoolAws(machinePool interface{}) *models.V1alpha1AwsMachinePoolConfigEntity {
	m := machinePool.(map[string]interface{})

	labels := make([]string, 0)
	controlPlane := m[control_plane].(bool)
	controlPlaneAsWorker := m[control_plane_as_worker].(bool)
	if controlPlane {
		labels = append(labels, "master")
	}

	azs := make([]string, 0)
	for _, az := range m["azs"].(*schema.Set).List() {
		azs = append(azs, az.(string))
	}

	mp := &models.V1alpha1AwsMachinePoolConfigEntity{
		CloudConfig: &models.V1alpha1AwsMachinePoolCloudConfigEntity{
			Azs:            azs,
			InstanceType:   ptr.StringPtr(m["instance_type"].(string)),
			RootDeviceSize: int64(m["disk_size_gb"].(int)),
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

//brownfield
func resourceClusterAzureImport(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1alpha1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	meta := toClusterMeta(d)

	uid, err := c.ImportClusterAzure(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(uid)

	stateConf := &resource.StateChangeConf{
		Target:     []string{string(pending)},
		Refresh:    resourceClusterStateRefreshFunc(c, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutCreate) - 1*time.Minute,
		MinTimeout: 1 * time.Second,
		Delay:      5 * time.Second,
	}

	// Wait, catching any errors
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	resourceClusterGcpRead(ctx, d, m)

	if profiles := getCloudClusterProfiles(d); profiles != nil {
		if err := c.UpdateBrownfieldCluster(uid, profiles); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

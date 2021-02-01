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

func resourceClusterGcp() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterGcpCreate,
		ReadContext:   resourceClusterGcpRead,
		UpdateContext: resourceClusterGcpUpdate,
		DeleteContext: resourceClusterDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

		SchemaVersion: 2,
		Schema: map[string]*schema.Schema{
			Name: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			ClusterProfileId: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			CloudAccountId: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			CloudConfigId: {
				Type:     schema.TypeString,
				Computed: true,
			},
			Kubeconfig: {
				Type:     schema.TypeString,
				Computed: true,
			},
			CloudConfig: {
				Type:     schema.TypeList,
				ForceNew: true,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						Network: {
							Type:     schema.TypeString,
							Optional: true,
						},
						Project: {
							Type:     schema.TypeString,
							Required: true,
						},
						Region: {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			Pack: {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      resourcePackHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						Name: {
							Type:     schema.TypeString,
							Required: true,
						},
						Tag: {
							Type:     schema.TypeString,
							Required: true,
						},
						Values: {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			MachinePool: {
				Type:     schema.TypeSet,
				Required: true,
				Set:      resourceMachinePoolGcpHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						ControlPlane: {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
							//ForceNew: true,
						},
						ControlPlaneAsWorker: {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,

							//ForceNew: true,
						},
						Name: {
							Type:     schema.TypeString,
							Required: true,
							//ForceNew: true,
						},
						Count: {
							Type:     schema.TypeInt,
							Required: true,
						},
						InstanceType: {
							Type:     schema.TypeString,
							Required: true,
						},
						UpdateStrategy: {
							Type:     schema.TypeString,
							Optional: true,
							Default:  RollingUpdateScaleOut,
						},
						DiskSizeInGb: {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  65,
						},
						AvailabilityZones: {
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

func resourceClusterGcpCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1alpha1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	cluster := toGcpCluster(d)

	uid, err := c.CreateClusterGcp(cluster)
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

	resourceClusterGcpRead(ctx, d, m)

	return diags
}

//goland:noinspection GoUnhandledErrorResult
func resourceClusterGcpRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
	d.Set(CloudConfigId, configUID)

	var config *models.V1alpha1GcpCloudConfig
	if config, err = c.GetCloudConfigGcp(configUID); err != nil {
		return diag.FromErr(err)
	}

	//for brownfield cluster
	if cluster.Status != nil && cluster.Status.ClusterImport != nil && cluster.Status.ClusterImport.IsBrownfield {
		importManifest, err := c.GetClusterImportManifest(uid)
		if err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set(ClusterImportManifest, importManifest); err != nil {
			return diag.FromErr(err)
		}

		if err := d.Set(ClusterImportManifestUrl, cluster.Status.ClusterImport.ImportLink); err != nil {
			return diag.FromErr(err)
		}

	} else {
		kubeconfig, err := c.GetClusterKubeConfig(uid)
		if err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set(Kubeconfig, kubeconfig); err != nil {
			return diag.FromErr(err)
		}
	}

	//for brownfield, untill cluster is not in running state, don't get machine pool
	if cluster.Status.ClusterImport == nil || cluster.Status.State == string(Running) {
		mp := flattenMachinePoolConfigsGcp(config.Spec.MachinePoolConfig)
		if err := d.Set(MachinePool, mp); err != nil {
			return diag.FromErr(err)
		}
	}
	return diags
}

func flattenMachinePoolConfigsGcp(machinePools []*models.V1alpha1GcpMachinePoolConfig) []interface{} {

	if machinePools == nil {
		return make([]interface{}, 0)
	}

	ois := make([]interface{}, len(machinePools))

	for i, machinePool := range machinePools {
		oi := make(map[string]interface{})

		oi[ControlPlane] = machinePool.IsControlPlane
		oi[ControlPlaneAsWorker] = machinePool.UseControlPlaneAsWorker
		oi[Name] = machinePool.Name
		oi[Count] = int(machinePool.Size)
		oi[DiskSizeInGb] = int(machinePool.RootDeviceSize)

		if machinePool.UpdateStrategy != nil {
			oi[UpdateStrategy] = machinePool.UpdateStrategy.Type
		}
		if machinePool.InstanceType != nil {
			oi[InstanceType] = *machinePool.InstanceType
		}
		if len(machinePool.Azs) > 0 {
			oi[AvailabilityZones] = machinePool.Azs
		}

		ois[i] = oi
	}

	return ois
}

func resourceClusterGcpUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
			hash := resourceMachinePoolGcpHash(machinePoolResource)

			machinePool := toMachinePoolGcp(machinePoolResource)

			var err error
			if oldMachinePool, ok := osMap[name]; !ok {
				log.Printf("Create machine pool %s", name)
				err = c.CreateMachinePoolGcp(cloudConfigId, machinePool)
			} else if hash != resourceMachinePoolGcpHash(oldMachinePool) {
				log.Printf("Change in machine pool %s", name)
				err = c.UpdateMachinePoolGcp(cloudConfigId, machinePool)
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
			if err := c.DeleteMachinePoolGcp(cloudConfigId, name); err != nil {
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
		cluster := toGcpCluster(d)
		if err := c.UpdateClusterGcp(cluster); err != nil {
			return diag.FromErr(err)
		}
	}

	resourceClusterGcpRead(ctx, d, m)

	return diags
}

func toGcpCluster(d *schema.ResourceData) *models.V1alpha1SpectroGcpClusterEntity {
	// gnarly, I know! =/
	cloudConfig := d.Get("cloud_config").([]interface{})[0].(map[string]interface{})
	//clientSecret := strfmt.Password(d.Get("gcp_client_secret").(string))

	cluster := &models.V1alpha1SpectroGcpClusterEntity{
		Metadata: &models.V1ObjectMeta{
			Name: d.Get("name").(string),
			UID:  d.Id(),
		},
		Spec: toGcpClusterSpec(d, cloudConfig),
	}

	//for _, machinePool := range d.Get("machine_pool").([]interface{}) {
	machinePoolConfigs := make([]*models.V1alpha1GcpMachinePoolConfigEntity, 0)
	for _, machinePool := range d.Get("machine_pool").(*schema.Set).List() {
		mp := toMachinePoolGcp(machinePool)
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

func toGcpClusterSpec(d *schema.ResourceData, cloudConfig map[string]interface{}) *models.V1alpha1SpectroGcpClusterEntitySpec {
	clusterSpec := &models.V1alpha1SpectroGcpClusterEntitySpec{
		ProfileUID:  d.Get(ClusterProfileId).(string),
		CloudConfig: toGcpClusterConfig(cloudConfig),
	}

	//for brownfield, there will be no cloud account
	if d.Get(CloudAccountId) != nil {
		clusterSpec.CloudAccountUID = ptr.StringPtr(d.Get(CloudAccountId).(string))
	}
	return clusterSpec
}

func toGcpClusterConfig(cloudConfig map[string]interface{}) *models.V1alpha1GcpClusterConfig {
	clusterConfig := &models.V1alpha1GcpClusterConfig{}
	if cloudConfig["network"] != nil {
		clusterConfig.Network = cloudConfig["network"].(string)
	}
	if cloudConfig["project"] != nil {
		clusterConfig.Project = ptr.StringPtr(cloudConfig["project"].(string))
	}
	if cloudConfig["region"] != nil {
		clusterConfig.Region = ptr.StringPtr(cloudConfig["region"].(string))
	}
	return clusterConfig
}

func toMachinePoolGcp(machinePool interface{}) *models.V1alpha1GcpMachinePoolConfigEntity {
	m := machinePool.(map[string]interface{})

	labels := make([]string, 0)
	controlPlane := m[ControlPlane].(bool)
	controlPlaneAsWorker := m[ControlPlaneAsWorker].(bool)
	if controlPlane {
		labels = append(labels, "master")
	}

	azs := make([]string, 0)
	for _, az := range m["azs"].(*schema.Set).List() {
		azs = append(azs, az.(string))
	}

	mp := &models.V1alpha1GcpMachinePoolConfigEntity{
		CloudConfig: &models.V1alpha1GcpMachinePoolCloudConfigEntity{
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
func resourceClusterGcpImport(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1alpha1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	meta := toClusterMeta(d)

	uid, err := c.ImportClusterGcp(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(uid)

	stateConf := &resource.StateChangeConf{
		//Pending:    resourceClusterCreatePendingStates,
		Target:     []string{string(Pending)},
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

	if profiles := resourceCloudClusterProfilesGet(d); profiles != nil {
		if err := c.UpdateBrownfieldCluster(uid, profiles); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

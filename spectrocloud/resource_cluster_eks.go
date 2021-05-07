package spectrocloud

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/go-cty/cty"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/gomi/pkg/ptr"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"
)

func resourceClusterEks() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterEksCreate,
		ReadContext:   resourceClusterEksRead,
		UpdateContext: resourceClusterEksUpdate,
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
						"ssh_key_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"vpc_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"region": {
							Type:     schema.TypeString,
							Required: true,
						},
						"endpoint_access": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: validateEndpointAccessType,
							Default:          "public",
						},
						"public_access_cidrs": {
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
				Type:     schema.TypeSet,
				Required: true,
				Set:      resourceMachinePoolAwsHash,
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
						"disk_size_gb": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  65,
						},
						"azs": {
							Type:     schema.TypeSet,
							Optional: true,
							MinItems: 1,
							Set:      schema.HashString,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"subnets": {
							Type:     schema.TypeMap,
							Optional: true,
							//Set:      schema.HashString,
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

func resourceClusterEksCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1alpha1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	cluster := toEksCluster(d)

	uid, err := c.CreateClusterEks(cluster)
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

	resourceClusterEksRead(ctx, d, m)

	return diags
}

//goland:noinspection GoUnhandledErrorResult
func resourceClusterEksRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	var config *models.V1alpha1EksCloudConfig
	if config, err = c.GetCloudConfigEks(configUID); err != nil {
		return diag.FromErr(err)
	}

	// Update the kubeconfig
	if cluster.Status != nil && cluster.Status.ClusterImport != nil && cluster.Status.ClusterImport.IsBrownfield {
		if err := d.Set("cluster_import_manifest_apply_command", cluster.Status.ClusterImport.ImportLink); err != nil {
			return diag.FromErr(err)
		}

		importManifest, err := c.GetClusterImportManifest(uid)
		if err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("cluster_import_manifest", importManifest); err != nil {
			return diag.FromErr(err)
		}
	} else {
		kubecfg, err := c.GetClusterKubeConfig(uid)
		if err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("kubeconfig", kubecfg); err != nil {
			return diag.FromErr(err)
		}
	}

	//for brownfield, until cluster is not in running state, don't get machine pool
	if cluster.Status.ClusterImport == nil || cluster.Status.State == "Running" {
		mp := flattenMachinePoolConfigsEks(config.Spec.MachinePoolConfig)
		if err := d.Set("machine_pool", mp); err != nil {
			return diag.FromErr(err)
		}
	}
	return diags
}

func flattenMachinePoolConfigsEks(machinePools []*models.V1alpha1EksMachinePoolConfig) []interface{} {

	if machinePools == nil {
		return make([]interface{}, 0)
	}

	ois := make([]interface{}, len(machinePools))

	for i, machinePool := range machinePools {
		oi := make(map[string]interface{})

		oi["control_plane"] = machinePool.IsControlPlane
		oi["control_plane_as_worker"] = machinePool.UseControlPlaneAsWorker
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

func resourceClusterEksUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

			machinePool := toMachinePoolEks(machinePoolResource)

			var err error
			if oldMachinePool, ok := osMap[name]; !ok {
				log.Printf("Create machine pool %s", name)
				err = c.CreateMachinePoolEks(cloudConfigId, machinePool)
			} else if hash != resourceMachinePoolAwsHash(oldMachinePool) {
				log.Printf("Change in machine pool %s", name)
				err = c.UpdateMachinePoolEks(cloudConfigId, machinePool)
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
			if err := c.DeleteMachinePoolEks(cloudConfigId, name); err != nil {
				return diag.FromErr(err)
			}
		}
	}
	//TODO(saamalik) update for cluster as well
	//if err := waitForClusterU(ctx, c, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
	//	return diag.FromErr(err)
	//}

	if d.HasChanges("pack") {
		if err := updatePacks(c, d); err != nil {
			return diag.FromErr(err)
		}
	}

	resourceClusterEksRead(ctx, d, m)

	return diags
}

func toEksCluster(d *schema.ResourceData) *models.V1alpha1SpectroEksClusterEntity {
	// gnarly, I know! =/
	cloudConfig := d.Get("cloud_config").([]interface{})[0].(map[string]interface{})
	//clientSecret := strfmt.Password(d.Get("Eks_client_secret").(string))

	cluster := &models.V1alpha1SpectroEksClusterEntity{
		Metadata: &models.V1ObjectMeta{
			Name: d.Get("name").(string),
			UID:  d.Id(),
		},
		Spec: &models.V1alpha1SpectroEksClusterEntitySpec{
			CloudAccountUID: ptr.StringPtr(d.Get("cloud_account_id").(string)),
			ProfileUID:      d.Get("cluster_profile_id").(string),
			CloudConfig: &models.V1alpha1EksClusterConfig{
				Region:     ptr.StringPtr(cloudConfig["region"].(string)),
				SSHKeyName: cloudConfig["ssh_key_name"].(string),
			},
		},
	}

	if cloudConfig["vpc_id"] != nil && len(cloudConfig["vpc_id"].(string)) > 0 {
		cluster.Spec.CloudConfig.VpcID = cloudConfig["vpc_id"].(string)
	}

	if cloudConfig["endpoint_access"] != nil && len(cloudConfig["endpoint_access"].(string)) > 0 {
		access := models.V1alpha1EksClusterConfigEndpointAccess{}
		if cloudConfig["endpoint_access"].(string) == "public" {
			//"public", "private", "private_and_public"
			access.Public = true
			access.Private = false
		} else if cloudConfig["endpoint_access"].(string) == "private" {
			//"public", "private", "private_and_public"
			access.Public = false
			access.Private = false
		} else if cloudConfig["endpoint_access"].(string) == "private_and_public" {
			//"public", "private", "private_and_public"
			access.Public = true
			access.Private = true
		}

		if cloudConfig["public_access_cidrs"] != nil {
			cidrs := make([]string, 0, 1)
			for _, cidr := range cloudConfig["public_access_cidrs"].(*schema.Set).List() {
				cidrs = append(cidrs, cidr.(string))
			}
			access.PublicCIDRs = cidrs
		}
	}

	//for _, machinePool := range d.Get("machine_pool").([]interface{}) {
	machinePoolConfigs := make([]*models.V1alpha1EksMachinePoolConfigEntity, 0)
	for _, machinePool := range d.Get("machine_pool").(*schema.Set).List() {
		mp := toMachinePoolEks(machinePool)
		machinePoolConfigs = append(machinePoolConfigs, mp)
	}

	cluster.Spec.Machinepoolconfig = machinePoolConfigs

	packValues := make([]*models.V1alpha1PackValuesEntity, 0)
	for _, pack := range d.Get("pack").([]interface{}) {
		p := toPack(pack)
		packValues = append(packValues, p)
	}
	cluster.Spec.PackValues = packValues
	//cluster.Spec.ClusterConfig = toClusterConfig(d)

	return cluster
}

func toMachinePoolEks(machinePool interface{}) *models.V1alpha1EksMachinePoolConfigEntity {
	m := machinePool.(map[string]interface{})

	labels := make([]string, 0)
	controlPlane := m["control_plane"].(bool)
	controlPlaneAsWorker := m["control_plane_as_worker"].(bool)
	if controlPlane {
		labels = append(labels, "master")
	}

	azs := make([]string, 0)
	for _, az := range m["azs"].(*schema.Set).List() {
		azs = append(azs, az.(string))
	}

	mp := &models.V1alpha1EksMachinePoolConfigEntity{
		CloudConfig: &models.V1alpha1EksMachineCloudConfigEntity{
			Azs:            azs,
			InstanceType:   m["instance_type"].(string),
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

	if v, ok := m["subnets"]; ok {
		//m := make(map[string]string)
		subnets := make([]*models.V1alpha1EksSubnetEntity, 0, 1)
		for k, val := range v.(map[string]interface{}) {
			//m[k] = val.(string)
			subnets = append(subnets, &models.V1alpha1EksSubnetEntity{
				Az: k,
				ID: val.(string),
			})
		}
		mp.CloudConfig.Subnets = subnets
	}

	// add subnet in machine pool
	return mp
}

func validateEndpointAccessType(data interface{}, path cty.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	andpointAccess := data.(string)
	for _, accessType := range []string{"public", "private", "private_and_public"} {
		if accessType == andpointAccess {
			return diags
		}
	}
	return diag.FromErr(fmt.Errorf("endpoint access type '%s' is invalid. valid endpoint access types are 'public', 'private' and 'private_and_public'", andpointAccess))
}

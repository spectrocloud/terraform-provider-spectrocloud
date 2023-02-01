package spectrocloud

import (
	"context"
	schemas "github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/schemas"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"
)

func resourceClusterEdgeNative() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterEdgeNativeCreate,
		ReadContext:   resourceClusterEdgeNativeRead,
		UpdateContext: resourceClusterEdgeNativeUpdate,
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
			"cluster_profile": schemas.ClusterProfileSchema(),
			"apply_setting": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"cloud_account_id": {
				Type:     schema.TypeString,
				Optional: true,
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
						"ssh_key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"vip": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"ntp_servers": {
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
			"machine_pool": {
				Type:     schema.TypeSet,
				Required: true,
				Set:      resourceMachinePoolEdgeNativeHash,
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
						"update_strategy": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "RollingUpdateScaleOut",
						},
						"host_uids": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
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
			"location_config":      schemas.ClusterLocationSchema(),
			"skip_completion": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func resourceClusterEdgeNativeCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	cluster := toEdgeNativeCluster(c, d)

	uid, err := c.CreateClusterEdgeNative(cluster)
	if err != nil {
		return diag.FromErr(err)
	}

	diagnostics, isError := waitForClusterCreation(ctx, d, uid, diags, c, true)
	if isError {
		return diagnostics
	}

	diags = resourceClusterEdgeNativeRead(ctx, d, m)

	return diags
}

//goland:noinspection GoUnhandledErrorResult
func resourceClusterEdgeNativeRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	// Update the kubeconfig
	diagnostics, errorSet := readCommonFields(c, d, cluster)
	if errorSet {
		return diagnostics
	}

	diags = flattenCloudConfigEdgeNative(cluster.Spec.CloudConfigRef.UID, d, c)
	return diags
}

func flattenCloudConfigEdgeNative(configUID string, d *schema.ResourceData, c *client.V1Client) diag.Diagnostics {
	if err := d.Set("cloud_config_id", configUID); err != nil {
		return diag.FromErr(err)
	}
	if config, err := c.GetCloudConfigEdgeNative(configUID); err != nil {
		return diag.FromErr(err)
	} else {
		mp := flattenMachinePoolConfigsEdgeNative(config.Spec.MachinePoolConfig)
		if err := d.Set("machine_pool", mp); err != nil {
			return diag.FromErr(err)
		}
	}

	return diag.Diagnostics{}
}

func flattenMachinePoolConfigsEdgeNative(machinePools []*models.V1EdgeNativeMachinePoolConfig) []interface{} {

	if machinePools == nil {
		return make([]interface{}, 0)
	}

	ois := make([]interface{}, 0)

	for _, machinePool := range machinePools {
		oi := make(map[string]interface{})

		SetAdditionalLabelsAndTaints(machinePool.AdditionalLabels, machinePool.Taints, oi)

		oi["control_plane"] = machinePool.IsControlPlane
		oi["control_plane_as_worker"] = machinePool.UseControlPlaneAsWorker
		oi["name"] = machinePool.Name
		hosts := make([]string, 0)
		for _, host := range machinePool.Hosts {
			hosts = append(hosts, *host.HostUID)
		}
		oi["host_uids"] = hosts
		flattenUpdateStrategy(machinePool.UpdateStrategy, oi)

		ois = append(ois, oi)
	}

	return ois
}

func resourceClusterEdgeNativeUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
			if name == "" {
				continue
			}
			hash := resourceMachinePoolEdgeNativeHash(machinePoolResource)

			machinePool := toMachinePoolEdgeNative(machinePoolResource)

			var err error
			if oldMachinePool, ok := osMap[name]; !ok {
				log.Printf("Create machine pool %s", name)
				err = c.CreateMachinePoolEdgeNative(cloudConfigId, machinePool)
			} else if hash != resourceMachinePoolEdgeNativeHash(oldMachinePool) {
				log.Printf("Change in machine pool %s", name)
				err = c.UpdateMachinePoolEdgeNative(cloudConfigId, machinePool)
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
			if err := c.DeleteMachinePoolEdgeNative(cloudConfigId, name); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	diagnostics, errorSet := updateCommonFields(d, c)
	if errorSet {
		return diagnostics
	}

	diags = resourceClusterEdgeNativeRead(ctx, d, m)

	return diags
}

func toEdgeNativeCluster(c *client.V1Client, d *schema.ResourceData) *models.V1SpectroEdgeNativeClusterEntity {
	cloudConfig := d.Get("cloud_config").([]interface{})[0].(map[string]interface{})

	controlPlaneEndpoint := &models.V1EdgeNativeControlPlaneEndPoint{}
	if cloudConfig["vip"] != nil {
		vip := cloudConfig["vip"].(string)
		controlPlaneEndpoint =
			&models.V1EdgeNativeControlPlaneEndPoint{
				//DdnsSearchDomain: cloudConfig["network_search_domain"].(string),
				Host: vip,
				Type: "IP", // only IP type for now no DDNS
			}
	}
	cluster := &models.V1SpectroEdgeNativeClusterEntity{
		Metadata: &models.V1ObjectMeta{
			Name:   d.Get("name").(string),
			UID:    d.Id(),
			Labels: toTags(d),
		},
		Spec: &models.V1SpectroEdgeNativeClusterEntitySpec{
			Profiles: toProfiles(c, d),
			Policies: toPolicies(d),
			CloudConfig: &models.V1EdgeNativeClusterConfig{
				NtpServers:           toNtpServers(cloudConfig),
				SSHKeys:              []string{cloudConfig["ssh_key"].(string)},
				ControlPlaneEndpoint: controlPlaneEndpoint,
			},
		},
	}

	machinePoolConfigs := make([]*models.V1EdgeNativeMachinePoolConfigEntity, 0)
	for _, machinePool := range d.Get("machine_pool").(*schema.Set).List() {
		mp := toMachinePoolEdgeNative(machinePool)
		machinePoolConfigs = append(machinePoolConfigs, mp)
	}
	cluster.Spec.Machinepoolconfig = machinePoolConfigs
	cluster.Spec.ClusterConfig = toClusterConfig(d)

	return cluster
}

func toMachinePoolEdgeNative(machinePool interface{}) *models.V1EdgeNativeMachinePoolConfigEntity {
	m := machinePool.(map[string]interface{})

	labels := make([]string, 0)
	controlPlane := m["control_plane"].(bool)
	controlPlaneAsWorker := m["control_plane_as_worker"].(bool)

	cloudConfig := toEdgeHosts(m)
	mp := &models.V1EdgeNativeMachinePoolConfigEntity{
		CloudConfig: cloudConfig,
		PoolConfig: &models.V1MachinePoolConfigEntity{
			AdditionalLabels: toAdditionalNodePoolLabels(m),
			Taints:           toClusterTaints(m),
			IsControlPlane:   controlPlane,
			Labels:           labels,
			Name:             types.Ptr(m["name"].(string)),
			Size:             types.Ptr(int32(len(cloudConfig.EdgeHosts))),
			UpdateStrategy: &models.V1UpdateStrategy{
				Type: getUpdateStrategy(m),
			},
			UseControlPlaneAsWorker: controlPlaneAsWorker,
		},
	}
	return mp
}

func toEdgeHosts(m map[string]interface{}) *models.V1EdgeNativeMachinePoolCloudConfigEntity {
	if m["host_uids"] == nil {
		return nil
	}
	edgeHosts := make([]*models.V1EdgeNativeMachinePoolHostEntity, 0)
	for _, host := range m["host_uids"].([]interface{}) {
		edgeHosts = append(edgeHosts, &models.V1EdgeNativeMachinePoolHostEntity{
			HostUID: types.Ptr(host.(string)),
		})
	}

	return &models.V1EdgeNativeMachinePoolCloudConfigEntity{
		EdgeHosts: edgeHosts,
	}
}

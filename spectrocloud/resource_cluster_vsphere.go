package spectrocloud

import (
	"context"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/palette-sdk-go/client"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/schemas"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

func resourceClusterVsphere() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterVsphereCreate,
		ReadContext:   resourceClusterVsphereRead,
		UpdateContext: resourceClusterVsphereUpdate,
		DeleteContext: resourceClusterDelete,
		Description:   "A resource to manage a vSphere cluster in Palette.",

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(180 * time.Minute),
			Update: schema.DefaultTimeout(180 * time.Minute),
			Delete: schema.DefaultTimeout(180 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the cluster.",
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
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "DownloadAndInstall",
				ValidateFunc: validation.StringInSlice([]string{"DownloadAndInstall", "DownloadAndInstallLater"}, false),
				Description: "The setting to apply the cluster profile. `DownloadAndInstall` will download and install packs in one action. " +
					"`DownloadAndInstallLater` will only download artifact and postpone install for later. " +
					"Default value is `DownloadAndInstall`.",
			},
			"cloud_account_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the cloud account to be used for the cluster. This cloud account must be of type `vsphere`.",
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
				Description:      "The date and time after which to patch the cluster. Prefix the time value with the respective RFC. Ex: `RFC3339: 2006-01-02T15:04:05Z07:00`",
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
						"datacenter": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The name of the datacenter in vSphere. This is the name of the datacenter as it appears in vSphere.",
						},
						"folder": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The name of the folder in vSphere. This is the name of the folder as it appears in vSphere.",
						},
						"image_template_folder": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The name of the image template folder in vSphere. This is the name of the folder as it appears in vSphere.",
						},

						"ssh_key": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The SSH key to be used for the cluster. This is the public key that will be used to access the cluster.",
						},

						"static_ip": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
							Description: "Whether to use static IP addresses for the cluster. If `true`, the cluster will use static IP addresses. " +
								"If `false`, the cluster will use DDNS. Default is `false`.",
						},

						// DHCP Properties
						"network_type": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The type of network to use for the cluster. This can be `VIP` or `DDNS`.",
						},
						"network_search_domain": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The search domain to use for the cluster in case of DHCP.",
						},
						"ntp_servers": {
							Type:     schema.TypeSet,
							Optional: true,
							Set:      schema.HashString,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "A list of NTP servers to be used by the cluster.",
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
							Description: "The name of the machine pool. This is used to identify the machine pool in the cluster.",
						},
						"additional_labels": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"taints": schemas.ClusterTaintsSchema(),
						"control_plane": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
							//ForceNew: true,
							Description: "Whether this machine pool is a control plane. Defaults to `false`.",
						},
						"control_plane_as_worker": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
							//ForceNew: true,
							Description: "Whether this machine pool is a control plane and a worker. Defaults to `false`.",
						},
						"count": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Number of nodes in the machine pool.",
						},
						"update_strategy": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "RollingUpdateScaleOut",
							Description:  "Update strategy for the machine pool. Valid values are `RollingUpdateScaleOut` and `RollingUpdateScaleIn`.",
							ValidateFunc: validation.StringInSlice([]string{"RollingUpdateScaleOut", "RollingUpdateScaleIn"}, false),
						},
						"instance_type": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"disk_size_gb": {
										Type:        schema.TypeInt,
										Required:    true,
										Description: "The size of the disk in GB.",
									},
									"memory_mb": {
										Type:        schema.TypeInt,
										Required:    true,
										Description: "The amount of memory in MB.",
									},
									"cpu": {
										Type:        schema.TypeInt,
										Required:    true,
										Description: "The number of CPUs.",
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
										Type:        schema.TypeString,
										Required:    true,
										Description: "The name of the cluster to use for the machine pool. As it appears in the vSphere.",
									},
									"resource_pool": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The name of the resource pool to use for the machine pool. As it appears in the vSphere.",
									},
									"datastore": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The name of the datastore to use for the machine pool. As it appears in the vSphere.",
									},
									"network": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The name of the network to use for the machine pool. As it appears in the vSphere.",
									},
									"static_ip_pool_id": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "The ID of the static IP pool to use for the machine pool in case of static cluster placement.",
									},
								},
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
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "If `true`, the cluster will be created asynchronously. Default value is `false`.",
			},
			"force_delete": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "If set to `true`, the cluster will be force deleted and user has to manually clean up the provisioned cloud resources.",
			},
			"force_delete_delay": {
				Type:             schema.TypeInt,
				Optional:         true,
				Default:          20,
				Description:      "Delay duration in minutes to before invoking cluster force delete. Default and minimum is 20 & maximum is 180 minutes.",
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(20, 180)),
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

	diagnostics, isError := waitForClusterCreation(ctx, d, uid, diags, c, true)
	if isError {
		return diagnostics
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
	if err := d.Set("cloud_config_id", configUID); err != nil {
		return diag.FromErr(err)
	}
	if config, err := c.GetCloudConfigVsphere(configUID); err != nil {
		return diag.FromErr(err)
	} else {
		cloudConfig, err := c.GetCloudConfigVsphereValues(configUID)
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

	if cloudConfig.Spec.ClusterConfig.NtpServers != nil {
		ret["ntp_servers"] = cloudConfig.Spec.ClusterConfig.NtpServers
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

		SetAdditionalLabelsAndTaints(machinePool.AdditionalLabels, machinePool.Taints, oi)

		oi["control_plane"] = machinePool.IsControlPlane
		oi["control_plane_as_worker"] = machinePool.UseControlPlaneAsWorker
		oi["name"] = machinePool.Name
		oi["count"] = machinePool.Size
		flattenUpdateStrategy(machinePool.UpdateStrategy, oi)

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

func ValidateMachinePoolChange(oraw interface{}, nraw interface{}) bool {
	hasChange := false
	for i, nm := range nraw.(*schema.Set).List() {
		oldPlacement := oraw.(*schema.Set).List()[i].(map[string]interface{})["placement"].([]interface{})[0].(map[string]interface{})
		newPlacement := nm.(map[string]interface{})["placement"].([]interface{})[0].(map[string]interface{})
		if (oldPlacement["cluster"] != newPlacement["cluster"]) || (oldPlacement["datastore"] != newPlacement["datastore"]) || (oldPlacement["network"] != newPlacement["network"]) {
			hasChange = true
			return hasChange
		}
	}
	return hasChange
}

func resourceClusterVsphereUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	cloudConfigId := d.Get("cloud_config_id").(string)

	if d.HasChange("cloud_config") {
		occ, ncc := d.GetChange("cloud_config")
		if occ.([]interface{})[0].(map[string]interface{})["datacenter"] != ncc.([]interface{})[0].(map[string]interface{})["datacenter"] {
			return diag.Errorf("Validation error: %s", "Datacenter value cannot be updated after cluster provisioning. Kindly destroy and recreate with updated Datacenter attribute.")
		}
		cloudConfig := toCloudConfigUpdate(d.Get("cloud_config").([]interface{})[0].(map[string]interface{}))
		if err := c.UpdateCloudConfigVsphereValues(cloudConfigId, cloudConfig); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("machine_pool") {
		oraw, nraw := d.GetChange("machine_pool")
		if oraw != nil && nraw != nil {
			if ValidateMachinePoolChange(oraw, nraw) {
				return diag.Errorf("Validation error: %s", "Datastore, Network, and ComputeCluster values cannot be updated after cluster provisioning. Kindly destroy and recreate with updated placement attributes")
			}
		}
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

	sort.SliceStable(machinePoolConfigs, func(i, j int) bool {
		return machinePoolConfigs[i].PoolConfig.IsControlPlane
	})

	cluster.Spec.Machinepoolconfig = machinePoolConfigs
	cluster.Spec.ClusterConfig = toClusterConfig(d)

	return cluster
}

func toCloudConfigCreate(cloudConfig map[string]interface{}) *models.V1VsphereClusterConfigEntity {
	staticIP := cloudConfig["static_ip"].(bool)

	V1VsphereClusterConfigEntity := getClusterConfigEntity(cloudConfig)

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
		ClusterConfig: toCloudConfigCreate(cloudConfig),
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
				NetworkName:   types.Ptr(p["network"].(string)),
				ParentPoolUID: poolID,
				StaticIP:      staticIP,
			},
		})

	}

	ins := m["instance_type"].([]interface{})[0].(map[string]interface{})
	instanceType := models.V1VsphereInstanceType{
		DiskGiB:   types.Ptr(int32(ins["disk_size_gb"].(int))),
		MemoryMiB: types.Ptr(int64(ins["memory_mb"].(int))),
		NumCPUs:   types.Ptr(int32(ins["cpu"].(int))),
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
			Name:             types.Ptr(m["name"].(string)),
			Size:             types.Ptr(int32(m["count"].(int))),
			UpdateStrategy: &models.V1UpdateStrategy{
				Type: getUpdateStrategy(m),
			},
			UseControlPlaneAsWorker: controlPlaneAsWorker,
		},
	}
	return mp
}

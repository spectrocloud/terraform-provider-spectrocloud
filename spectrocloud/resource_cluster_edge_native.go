package spectrocloud

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	schemas "github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/schemas"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/palette-sdk-go/client"
)

func resourceClusterEdgeNative() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterEdgeNativeCreate,
		ReadContext:   resourceClusterEdgeNativeRead,
		UpdateContext: resourceClusterEdgeNativeUpdate,
		DeleteContext: resourceClusterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceClusterEdgeNativeImport,
		},
		Description: "Resource for managing Edge Native clusters in Spectro Cloud through Palette.",

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
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"", "project", "tenant"}, false),
				Description: "The context of the Edge cluster. Allowed values are `project` or `tenant`. " +
					"Default is `project`. " + PROJECT_NAME_NUANCE,
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
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The description of the cluster. Default value is empty string.",
			},
			"cluster_meta_attribute": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "`cluster_meta_attribute` can be used to set additional cluster metadata information, eg `{'nic_name': 'test', 'env': 'stage'}`",
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
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"cloud_config_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the cloud config used for the cluster. This cloud config must be of type `azure`.",
				Deprecated:  "This field is deprecated and will be removed in the future. Use `cloud_config` instead.",
			},
			"review_repave_state": {
				Type:         schema.TypeString,
				Default:      "",
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"", "Approved", "Pending"}, false),
				Description:  "To authorize the cluster repave, set the value to `Approved` for approval and `\"\"` to decline. Default value is `\"\"`.",
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
				Description:      "Date and time after which to patch cluster `RFC3339: 2006-01-02T15:04:05Z07:00`",
			},
			"kubeconfig": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Kubeconfig for the cluster. This can be used to connect to the cluster using `kubectl`.",
			},
			"admin_kube_config": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Admin Kube-config for the cluster. This can be used to connect to the cluster using `kubectl`, With admin privilege.",
			},
			"cloud_config": {
				Type:     schema.TypeList,
				ForceNew: true,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ssh_keys": {
							Type:     schema.TypeSet,
							Optional: true,
							Set:      schema.HashString,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "List of public SSH (Secure Shell) to establish, administer, and communicate with remote clusters.",
						},
						"vip": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The control plane endpoint can be specified as either an IP address or a fully qualified domain name (FQDN).",
						},
						"overlay_cidr_range": {
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
							Description: "The `overlay_cidr_range` parameter configures the overlay network. When `overlay_cidr_range` is set, it enables the overlay network. For example, `100.64.192.0/24`",
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
						"node":   schemas.NodeSchema(),
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
						"node_repave_interval": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     0,
							Description: "Minimum number of seconds node should be Ready, before the next node is selected for repave. Default value is `0`, Applicable only for worker pools.",
						},
						"update_strategy": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "RollingUpdateScaleOut",
							Description:  "Update strategy for the machine pool. Valid values are `RollingUpdateScaleOut` and `RollingUpdateScaleIn`.",
							ValidateFunc: validation.StringInSlice([]string{"RollingUpdateScaleOut", "RollingUpdateScaleIn"}, false),
						},
						"edge_host": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"host_name": {
										Type:        schema.TypeString,
										Optional:    true,
										Default:     "",
										Description: "Edge host name",
									},
									"host_uid": {
										Type:        schema.TypeString,
										Description: "Edge host id",
										Required:    true,
									},
									"static_ip": {
										Type:        schema.TypeString,
										Description: "Edge host static IP",
										Optional:    true,
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
				Description:      "Delay duration in minutes to before invoking cluster force delete. Default and minimum is 20.",
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(20)),
			},
		},
	}
}

func resourceClusterEdgeNativeCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	cluster, err := toEdgeNativeCluster(c, d)
	if err != nil {
		return diag.FromErr(err)
	}

	ClusterContext := d.Get("context").(string)
	uid, err := c.CreateClusterEdgeNative(cluster, ClusterContext)
	if err != nil {
		return diag.FromErr(err)
	}

	diagnostics, isError := waitForClusterCreation(ctx, d, ClusterContext, uid, diags, c, true)
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

	cluster, err := resourceClusterRead(d, c, diags)
	if err != nil {
		return diag.FromErr(err)
	} else if cluster == nil {
		// Deleted - Terraform will recreate it
		d.SetId("")
		return diags
	}

	// verify cluster type
	err = ValidateCloudType("spectrocloud_cluster_edge_native", cluster)
	if err != nil {
		return diag.FromErr(err)
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
	ClusterContext := d.Get("context").(string)
	if err := d.Set("cloud_config_id", configUID); err != nil {
		return diag.FromErr(err)
	}
	if err := ReadCommonAttributes(d); err != nil {
		return diag.FromErr(err)
	}

	if config, err := c.GetCloudConfigEdgeNative(configUID, ClusterContext); err != nil {
		return diag.FromErr(err)
	} else {
		cloudConfig := d.Get("cloud_config").([]interface{})[0].(map[string]interface{})
		if err := d.Set("cloud_config", flattenClusterConfigsEdgeNative(cloudConfig, config)); err != nil {
			return diag.FromErr(err)
		}
		mp := flattenMachinePoolConfigsEdgeNative(config.Spec.MachinePoolConfig)
		mp, err := flattenNodeMaintenanceStatus(c, d, c.GetNodeStatusMapEdgeNative, mp, configUID, ClusterContext)
		if err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("machine_pool", mp); err != nil {
			return diag.FromErr(err)
		}
	}

	return diag.Diagnostics{}
}

func flattenClusterConfigsEdgeNative(cloudConfig map[string]interface{}, config *models.V1EdgeNativeCloudConfig) []interface{} {
	if config == nil || config.Spec == nil || config.Spec.ClusterConfig == nil {
		return make([]interface{}, 0)
	}

	m := make(map[string]interface{})
	if config.Spec.ClusterConfig.SSHKeys != nil {
		m["ssh_keys"] = config.Spec.ClusterConfig.SSHKeys
	}
	if config.Spec.ClusterConfig.ControlPlaneEndpoint.Host != "" {
		if v, ok := cloudConfig["vip"]; ok && v.(string) != "" {
			m["vip"] = config.Spec.ClusterConfig.ControlPlaneEndpoint.Host
		}
	}
	if config.Spec.ClusterConfig.NtpServers != nil {
		m["ntp_servers"] = config.Spec.ClusterConfig.NtpServers
	}
	if config.Spec.ClusterConfig.OverlayNetworkConfiguration.Cidr != "" {
		m["overlay_cidr_range"] = config.Spec.ClusterConfig.OverlayNetworkConfiguration.Cidr
	}

	return []interface{}{m}
}

func flattenMachinePoolConfigsEdgeNative(machinePools []*models.V1EdgeNativeMachinePoolConfig) []interface{} {

	if machinePools == nil {
		return make([]interface{}, 0)
	}

	ois := make([]interface{}, 0)

	for _, machinePool := range machinePools {
		oi := make(map[string]interface{})

		FlattenAdditionalLabelsAndTaints(machinePool.AdditionalLabels, machinePool.Taints, oi)
		FlattenControlPlaneAndRepaveInterval(&machinePool.IsControlPlane, oi, machinePool.NodeRepaveInterval)
		oi["control_plane"] = machinePool.IsControlPlane
		oi["control_plane_as_worker"] = machinePool.UseControlPlaneAsWorker
		oi["name"] = machinePool.Name
		var hosts []map[string]string
		for _, host := range machinePool.Hosts {
			hosts = append(hosts, map[string]string{
				"host_name": host.HostName,
				"host_uid":  *host.HostUID,
				"static_ip": host.StaticIP,
			})
		}
		oi["edge_host"] = hosts
		flattenUpdateStrategy(machinePool.UpdateStrategy, oi)

		ois = append(ois, oi)
	}

	return ois
}

func resourceClusterEdgeNativeUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	err := validateSystemRepaveApproval(d, c)
	if err != nil {
		return diag.FromErr(err)
	}

	cloudConfigId := d.Get("cloud_config_id").(string)
	ClusterContext := d.Get("context").(string)
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

		nsMap := make(map[string]interface{})

		for _, mp := range ns.List() {
			machinePoolResource := mp.(map[string]interface{})
			nsMap[machinePoolResource["name"].(string)] = machinePoolResource
			// since known issue in TF SDK: https://github.com/hashicorp/terraform-plugin-sdk/issues/588
			if machinePoolResource["name"].(string) != "" {
				name := machinePoolResource["name"].(string)
				if name == "" {
					continue
				}
				hash := resourceMachinePoolEdgeNativeHash(machinePoolResource)
				var err error
				machinePool, err := toMachinePoolEdgeNative(machinePoolResource)
				if err != nil {
					return diag.FromErr(err)
				}

				if oldMachinePool, ok := osMap[name]; !ok {
					log.Printf("Create machine pool %s", name)
					err = c.CreateMachinePoolEdgeNative(cloudConfigId, ClusterContext, machinePool)
				} else if hash != resourceMachinePoolEdgeNativeHash(oldMachinePool) {
					log.Printf("Change in machine pool %s", name)
					err = c.UpdateMachinePoolEdgeNative(cloudConfigId, ClusterContext, machinePool)
					err := resourceNodeAction(c, ctx, nsMap[name], c.GetNodeMaintenanceStatusEdgeNative, "edge-native", ClusterContext, cloudConfigId, name)
					if err != nil {
						return diag.FromErr(err)
					}
				}

				if err != nil {
					return diag.FromErr(err)
				}

				// Processed (if exists)
				delete(osMap, name)
			}
		}

		// Deleted old machine pools
		for _, mp := range osMap {
			machinePool := mp.(map[string]interface{})
			name := machinePool["name"].(string)
			log.Printf("Deleted machine pool %s", name)
			if err := c.DeleteMachinePoolEdgeNative(cloudConfigId, name, ClusterContext); err != nil {
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

func toEdgeNativeCluster(c *client.V1Client, d *schema.ResourceData) (*models.V1SpectroEdgeNativeClusterEntity, error) {
	cloudConfig := d.Get("cloud_config").([]interface{})[0].(map[string]interface{})
	sshKeys, _ := toSSHKeys(cloudConfig)

	clusterContext := d.Get("context").(string)
	profiles, err := toProfiles(c, d, clusterContext)
	if err != nil {
		return nil, err
	}

	controlPlaneEndpoint, overlayConfig, err := toOverlayNetworkConfigAndVip(cloudConfig)
	if err != nil {
		return nil, err
	}

	cluster := &models.V1SpectroEdgeNativeClusterEntity{
		Metadata: getClusterMetadata(d),
		Spec: &models.V1SpectroEdgeNativeClusterEntitySpec{
			Profiles: profiles,
			Policies: toPolicies(d),
			CloudConfig: &models.V1EdgeNativeClusterConfig{
				NtpServers:                  toNtpServers(cloudConfig),
				SSHKeys:                     sshKeys,
				ControlPlaneEndpoint:        controlPlaneEndpoint,
				OverlayNetworkConfiguration: overlayConfig,
			},
		},
	}

	machinePoolConfigs := make([]*models.V1EdgeNativeMachinePoolConfigEntity, 0)
	for _, machinePool := range d.Get("machine_pool").(*schema.Set).List() {
		mp, err := toMachinePoolEdgeNative(machinePool)
		if err != nil {
			return nil, err
		}
		machinePoolConfigs = append(machinePoolConfigs, mp)
	}
	cluster.Spec.Machinepoolconfig = machinePoolConfigs
	cluster.Spec.ClusterConfig = toClusterConfig(d)

	return cluster, nil
}

func toMachinePoolEdgeNative(machinePool interface{}) (*models.V1EdgeNativeMachinePoolConfigEntity, error) {
	m := machinePool.(map[string]interface{})

	labels := make([]string, 0)
	controlPlane := m["control_plane"].(bool)
	controlPlaneAsWorker := m["control_plane_as_worker"].(bool)
	if controlPlane {
		labels = append(labels, "master")
	}

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

	nodeRepaveInterval := 0
	if m["node_repave_interval"] != nil {
		nodeRepaveInterval = m["node_repave_interval"].(int)
	}
	if !controlPlane {
		mp.PoolConfig.NodeRepaveInterval = int32(nodeRepaveInterval)
	} else {
		err := ValidationNodeRepaveIntervalForControlPlane(nodeRepaveInterval)
		if err != nil {
			return mp, err
		}
	}

	return mp, nil
}

func toEdgeHosts(m map[string]interface{}) *models.V1EdgeNativeMachinePoolCloudConfigEntity {
	edgeHostIdsLen := len(m["edge_host"].([]interface{}))
	edgeHosts := make([]*models.V1EdgeNativeMachinePoolHostEntity, 0)
	if m["edge_host"] == nil || edgeHostIdsLen == 0 {
		return nil
	}
	for _, host := range m["edge_host"].([]interface{}) {
		hostName := ""
		if v, ok := host.(map[string]interface{})["host_name"].(string); ok {
			hostName = v
		}
		hostId := host.(map[string]interface{})["host_uid"].(string)
		edgeHosts = append(edgeHosts, &models.V1EdgeNativeMachinePoolHostEntity{
			HostName: hostName,
			HostUID:  &hostId,
			StaticIP: host.(map[string]interface{})["static_ip"].(string),
		})
	}
	return &models.V1EdgeNativeMachinePoolCloudConfigEntity{
		EdgeHosts: edgeHosts,
	}
}

func toOverlayNetworkConfigAndVip(cloudConfig map[string]interface{}) (*models.V1EdgeNativeControlPlaneEndPoint, *models.V1EdgeNativeOverlayNetworkConfiguration, error) {
	controlPlaneEndpoint := &models.V1EdgeNativeControlPlaneEndPoint{}
	overlayConfig := &models.V1EdgeNativeOverlayNetworkConfiguration{}
	if (cloudConfig["overlay_cidr_range"] != nil) && (cloudConfig["overlay_cidr_range"].(string) != "") {
		overlayConfig.Cidr = cloudConfig["overlay_cidr_range"].(string)
		overlayConfig.Enable = true
	} else {
		overlayConfig.Cidr = ""
		overlayConfig.Enable = false
	}

	if (cloudConfig["vip"] != nil) && (cloudConfig["vip"].(string) != "") {
		vip := cloudConfig["vip"].(string)
		controlPlaneEndpoint =
			&models.V1EdgeNativeControlPlaneEndPoint{
				Host: vip,
				Type: "IP",
			}
	} else {
		if overlayConfig.Enable {
			autoGenVip, err := getFirstIPRange(overlayConfig.Cidr)
			if err != nil {
				return nil, nil, err
			}
			controlPlaneEndpoint =
				&models.V1EdgeNativeControlPlaneEndPoint{
					Host: autoGenVip,
					Type: "IP",
				}
		}
	}

	return controlPlaneEndpoint, overlayConfig, nil
}

func getFirstIPRange(cidr string) (string, error) {
	// Parse the CIDR string
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", err
	}

	// Get the network address from the parsed CIDR
	networkIP := ipNet.IP

	// Ensure that the subnet mask is applied correctly
	firstIP := make(net.IP, len(networkIP))
	copy(firstIP, networkIP)
	for i := range firstIP {
		firstIP[i] &= ipNet.Mask[i]
	}

	// Increment the last octet to get the first usable IP
	firstIP[len(firstIP)-1]++

	// Convert the IP address to a string
	firstIPString := firstIP.String()

	return firstIPString, nil
}

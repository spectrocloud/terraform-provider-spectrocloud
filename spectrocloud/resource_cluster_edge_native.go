package spectrocloud

import (
	"context"
	"log"
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
		Description:   "Resource for managing Edge Native clusters in Spectro Cloud through Palette.",

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
				Type:     schema.TypeString,
				Optional: true,
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
			"cloud_config": {
				Type:     schema.TypeList,
				ForceNew: true,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ssh_key": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "SSH Key (Secure Shell) to establish, administer, and communicate with remote clusters, `ssh_key & ssh_keys` are mutually exclusive.",
						},
						"ssh_keys": {
							Type:          schema.TypeSet,
							Optional:      true,
							Set:           schema.HashString,
							ConflictsWith: []string{"cloud_config.0.ssh_key"},
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "List of SSH (Secure Shell) to establish, administer, and communicate with remote clusters, `ssh_key & ssh_keys` are mutually exclusive.",
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
						"update_strategy": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "RollingUpdateScaleOut",
							Description:  "Update strategy for the machine pool. Valid values are `RollingUpdateScaleOut` and `RollingUpdateScaleIn`.",
							ValidateFunc: validation.StringInSlice([]string{"RollingUpdateScaleOut", "RollingUpdateScaleIn"}, false),
						},
						"host_uids": {
							Type:       schema.TypeList,
							Optional:   true,
							Deprecated: "This field is deprecated from provider 0.13.0. Use `edge_host` instead.",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"edge_host": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
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
	if config, err := c.GetCloudConfigEdgeNative(configUID, ClusterContext); err != nil {
		return diag.FromErr(err)
	} else {
		mp := flattenMachinePoolConfigsEdgeNative(config.Spec.MachinePoolConfig)
		mp, err := flattenNodeMaintenanceStatus(c, c.GetMachinesItemsActionsEdgeNative, mp, configUID, ClusterContext)
		if err != nil {
			return diag.FromErr(err)
		}
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

		FlattenAdditionalLabelsAndTaints(machinePool.AdditionalLabels, machinePool.Taints, oi)

		oi["control_plane_as_worker"] = machinePool.UseControlPlaneAsWorker
		oi["name"] = machinePool.Name
		var hosts []map[string]string
		for _, host := range machinePool.Hosts {
			hosts = append(hosts, map[string]string{
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

	profiles, err := toProfiles(c, d)
	if err != nil {
		return nil, err
	}
	cluster := &models.V1SpectroEdgeNativeClusterEntity{
		Metadata: &models.V1ObjectMeta{
			Name:   d.Get("name").(string),
			UID:    d.Id(),
			Labels: toTags(d),
		},
		Spec: &models.V1SpectroEdgeNativeClusterEntitySpec{
			Profiles: profiles,
			Policies: toPolicies(d),
			CloudConfig: &models.V1EdgeNativeClusterConfig{
				NtpServers:           toNtpServers(cloudConfig),
				SSHKeys:              sshKeys,
				ControlPlaneEndpoint: controlPlaneEndpoint,
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

	return mp, nil
}

func toEdgeHosts(m map[string]interface{}) *models.V1EdgeNativeMachinePoolCloudConfigEntity {
	edgeHostIdsLen := len(m["edge_host"].([]interface{}))
	edgeHosts := make([]*models.V1EdgeNativeMachinePoolHostEntity, 0)
	if m["edge_host"] == nil || edgeHostIdsLen == 0 {
		return nil
	}
	for _, host := range m["edge_host"].([]interface{}) {
		hostId := host.(map[string]interface{})["host_uid"].(string)
		edgeHosts = append(edgeHosts, &models.V1EdgeNativeMachinePoolHostEntity{
			HostUID:  &hostId,
			StaticIP: host.(map[string]interface{})["static_ip"].(string),
		})
	}
	return &models.V1EdgeNativeMachinePoolCloudConfigEntity{
		EdgeHosts: edgeHosts,
	}
}

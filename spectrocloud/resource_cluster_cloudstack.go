package spectrocloud

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/schemas"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

// safeInt32Conversion safely converts int to int32 with overflow protection
// Returns the converted value and true if conversion is safe, or defaultVal and false if overflow would occur
func safeInt32Conversion(value int, defaultVal int32) int32 {
	if value < math.MinInt32 || value > math.MaxInt32 {
		return defaultVal
	}
	return int32(value)
}

func resourceClusterCloudStack() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterCloudStackCreate,
		ReadContext:   resourceClusterCloudStackRead,
		UpdateContext: resourceClusterCloudStackUpdate,
		DeleteContext: resourceClusterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceClusterCloudStackImport,
		},
		Description: "Resource for managing CloudStack clusters in Spectro Cloud through Palette.",

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

		SchemaVersion: 2,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the cluster.",
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"", "project", "tenant"}, false),
				Description: "The context of the CloudStack configuration. Allowed values are `project` or `tenant`. " +
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
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the CloudStack cloud account used for the cluster. This cloud account must be of type `cloudstack`.",
			},
			"cloud_config_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the cloud config used for the cluster. This cloud config must be of type `cloudstack`.",
				Deprecated:  "This field is deprecated and will be removed in the future. Use `cloud_config` instead.",
			},
			"review_repave_state": {
				Type:         schema.TypeString,
				Default:      "",
				Optional:     true,
				ValidateFunc: validateReviewRepaveValue,
				Description:  "To authorize the cluster repave, set the value to `Approved` for approval and `\"\"` to decline. Default value is `\"\"`.",
			},
			"pause_agent_upgrades": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "unlock",
				ValidateFunc: validation.StringInSlice([]string{"lock", "unlock"}, false),
				Description:  "The pause agent upgrades setting allows to control the automatic upgrade of the Palette component and agent for an individual cluster. The default value is `unlock`, meaning upgrades occur automatically. Setting it to `lock` pauses automatic agent upgrades for the cluster.",
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
				Description:      "Cron schedule for OS patching. This must be in the form of `0 0 * * *`.",
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
			"admin_kube_config": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Admin Kube-config for the cluster. This can be used to connect to the cluster using `kubectl`, With admin privilege.",
			},
			"cloud_config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"domain": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "CloudStack domain name in which the cluster will be provisioned.",
						},
						"project": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "CloudStack project name (optional). If not specified, the cluster will be created in the domain's default project.",
						},
						"ssh_key_name": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "SSH key name for accessing cluster nodes.",
						},
						"control_plane_endpoint": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Endpoint IP to be used for the API server. Should only be set for static CloudStack networks.",
						},
						"sync_with_cks": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Determines if an external managed CKS (CloudStack Kubernetes Service) cluster should be created. Default is `false`.",
						},
						"zone": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "CloudStack zone name where the cluster will be deployed.",
									},
									"network": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "Network name in this zone.",
												},
												"type": {
													Type:        schema.TypeString,
													Optional:    true,
													Description: "Network type: Isolated, Shared, etc.",
												},
												"gateway": {
													Type:        schema.TypeString,
													Optional:    true,
													Description: "Gateway IP address for the network.",
												},
												"netmask": {
													Type:        schema.TypeString,
													Optional:    true,
													Description: "Network mask for the network.",
												},
												"offering": {
													Type:        schema.TypeString,
													Optional:    true,
													Description: "Network offering name to use when creating the network. Optional for advanced network configurations.",
												},
												"routing_mode": {
													Type:        schema.TypeString,
													Optional:    true,
													Description: "Routing mode for the network (e.g., Static, Dynamic). Optional, defaults to CloudStack's default routing mode.",
												},
												"vpc": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"name": {
																Type:        schema.TypeString,
																Required:    true,
																Description: "VPC name.",
															},
															"cidr": {
																Type:        schema.TypeString,
																Optional:    true,
																Description: "CIDR block for the VPC (e.g., 10.0.0.0/16).",
															},
															"offering": {
																Type:        schema.TypeString,
																Optional:    true,
																Description: "VPC offering name.",
															},
														},
													},
													Description: "VPC configuration for VPC-based network deployments. Optional, only needed when deploying in a VPC.",
												},
											},
										},
										Description: "Network configuration for this zone.",
									},
								},
							},
							Description: "List of CloudStack zones for multi-AZ deployments. If only one zone is specified, it will be treated as single-zone deployment.",
						},
					},
				},
				Description: "CloudStack cluster configuration.",
			},
			"machine_pool": {
				Type:     schema.TypeSet,
				Required: true,
				Set:      resourceMachinePoolCloudStackHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"additional_labels": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "Additional labels to be applied to the machine pool. Labels must be in the form of `key:value`.",
						},
						"node":   schemas.NodeSchema(),
						"taints": schemas.ClusterTaintsSchema(),
						"control_plane": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Whether this machine pool is a control plane. Defaults to `false`.",
						},
						"control_plane_as_worker": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Whether this machine pool is a control plane and a worker. Defaults to `false`.",
						},
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Name of the machine pool.",
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
						"min": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Minimum number of nodes in the machine pool. This is used for autoscaling.",
						},
						"max": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Maximum number of nodes in the machine pool. This is used for autoscaling.",
						},
						"offering": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "CloudStack compute offering (instance type/size) name.",
						},
						"disk_offering": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "CloudStack disk offering name for root disk (optional).",
						},
						"root_disk_size_gb": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Root disk size in GB (optional).",
						},
						"affinity_group_ids": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "List of affinity group IDs for VM placement (optional).",
						},
						"details": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "Additional details for instance creation as key-value pairs.",
						},
						"network": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"network_name": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Network name to attach to the machine pool.",
									},
									"ip_address": {
										Type:        schema.TypeString,
										Optional:    true,
										Deprecated:  "This field is no longer supported by the CloudStack API and will be ignored.",
										Description: "Static IP address to assign. **DEPRECATED**: This field is no longer supported by CloudStack and will be ignored.",
									},
								},
							},
							Description: "Network configuration for the machine pool instances.",
						},
					},
				},
				Description: "Machine pool configuration for the cluster.",
			},
			"backup_policy":        schemas.BackupPolicySchema(),
			"scan_policy":          schemas.ScanPolicySchema(),
			"cluster_rbac_binding": schemas.ClusterRbacBindingSchema(),
			"namespaces":           schemas.ClusterNamespacesSchema(),
			"host_config":          schemas.ClusterHostConfigSchema(),
			"location_config":      schemas.ClusterLocationSchemaComputed(),
		},
	}
}

func resourceClusterCloudStackCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "")
	var diags diag.Diagnostics

	cluster, err := toCloudStackCluster(c, d)
	if err != nil {
		return diag.FromErr(err)
	}

	ClusterContext := d.Get("context").(string)
	if ClusterContext == "" {
		ClusterContext = "project"
	}
	c = getV1ClientWithResourceContext(m, ClusterContext)
	uid, err := c.CreateClusterCloudStack(cluster)
	if err != nil {
		return diag.FromErr(err)
	}

	diags, done := waitForClusterCreation(ctx, d, uid, diags, c, false)
	if done {
		return diags
	}

	resourceClusterCloudStackRead(ctx, d, m)

	return diags
}

func resourceClusterCloudStackRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "")
	var diags diag.Diagnostics

	cluster, err := c.GetCluster(d.Id())
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

	diagnostics, done := readCommonFields(c, d, cluster)
	if done {
		return diagnostics
	}

	return diags
}

func resourceClusterCloudStackUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	cloudConfigId := d.Get("cloud_config_id").(string)
	ClusterContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, ClusterContext)

	if err := validateSystemRepaveApproval(d, c); err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("machine_pool") {
		if err := updateMachinePoolCloudStack(ctx, c, d, cloudConfigId); err != nil {
			return diag.FromErr(err)
		}
	}

	// Check common updates
	diagnostics, done := updateCommonFields(d, c)
	if done {
		return diagnostics
	}

	resourceClusterCloudStackRead(ctx, d, m)

	return diags
}

func toCloudStackCluster(c *client.V1Client, d *schema.ResourceData) (*models.V1SpectroCloudStackClusterEntity, error) {
	cloudConfig := toCloudStackCloudConfig(d)

	clusterContext := d.Get("context").(string)
	profiles, err := toProfiles(c, d, clusterContext)
	if err != nil {
		return nil, err
	}

	// Convert metadata to input entity type
	metadata := getClusterMetadata(d)
	cluster := &models.V1SpectroCloudStackClusterEntity{
		Metadata: &models.V1ObjectMetaInputEntity{
			Name:        metadata.Name,
			Labels:      metadata.Labels,
			Annotations: metadata.Annotations,
		},
		Spec: &models.V1SpectroCloudStackClusterEntitySpec{
			CloudAccountUID: types.Ptr(d.Get("cloud_account_id").(string)),
			Profiles:        profiles,
			Policies:        toPolicies(d),
			CloudConfig:     cloudConfig,
		},
	}

	machinePoolConfigs := make([]*models.V1CloudStackMachinePoolConfigEntity, 0)
	for _, machinePool := range d.Get("machine_pool").(*schema.Set).List() {
		mp := toMachinePoolCloudStack(machinePool)
		machinePoolConfigs = append(machinePoolConfigs, mp)
	}
	cluster.Spec.Machinepoolconfig = machinePoolConfigs
	cluster.Spec.ClusterConfig = toClusterConfig(d)

	return cluster, nil
}

func toCloudStackCloudConfig(d *schema.ResourceData) *models.V1CloudStackClusterConfig {
	cloudConfig := d.Get("cloud_config").([]interface{})[0].(map[string]interface{})

	config := &models.V1CloudStackClusterConfig{
		Domain:               cloudConfig["domain"].(string),
		Project:              cloudConfig["project"].(string),
		SSHKeyName:           cloudConfig["ssh_key_name"].(string),
		ControlPlaneEndpoint: cloudConfig["control_plane_endpoint"].(string),
		SyncWithCKS:          cloudConfig["sync_with_cks"].(bool),
	}

	// Process zones
	if zones, ok := cloudConfig["zone"].([]interface{}); ok && len(zones) > 0 {
		config.Zones = make([]*models.V1CloudStackZoneSpec, 0, len(zones))
		for _, z := range zones {
			zone := z.(map[string]interface{})
			zoneSpec := &models.V1CloudStackZoneSpec{
				Name: zone["name"].(string),
			}

			// Process network configuration for the zone
			if networks, ok := zone["network"].([]interface{}); ok && len(networks) > 0 {
				network := networks[0].(map[string]interface{})
				zoneSpec.Network = &models.V1CloudStackNetworkSpec{
					Name:        network["name"].(string),
					Type:        network["type"].(string),
					Gateway:     network["gateway"].(string),
					Netmask:     network["netmask"].(string),
					Offering:    network["offering"].(string),
					RoutingMode: network["routing_mode"].(string),
				}

				// Process VPC configuration if present
				if vpcs, ok := network["vpc"].([]interface{}); ok && len(vpcs) > 0 {
					vpc := vpcs[0].(map[string]interface{})
					zoneSpec.Network.Vpc = &models.V1CloudStackVPCSpec{
						Name:     vpc["name"].(string),
						Cidr:     vpc["cidr"].(string),
						Offering: vpc["offering"].(string),
					}
				}
			}

			config.Zones = append(config.Zones, zoneSpec)
		}
	}

	return config
}

func toMachinePoolCloudStack(machinePool interface{}) *models.V1CloudStackMachinePoolConfigEntity {
	mp := machinePool.(map[string]interface{})

	labels := make([]string, 0)
	controlPlane := mp["control_plane"].(bool)
	controlPlaneAsWorker := mp["control_plane_as_worker"].(bool)
	if controlPlane {
		labels = append(labels, "master")
	}

	cloudConfig := &models.V1CloudStackMachinePoolCloudConfigEntity{
		Offering:       types.Ptr(mp["offering"].(string)),
		DiskOffering:   mp["disk_offering"].(string),
		RootDiskSizeGB: safeInt32Conversion(mp["root_disk_size_gb"].(int), 0),
	}

	// Process affinity groups
	if affinityGroups, ok := mp["affinity_group_ids"].(*schema.Set); ok {
		cloudConfig.AffinityGroupIds = make([]string, 0)
		for _, ag := range affinityGroups.List() {
			cloudConfig.AffinityGroupIds = append(cloudConfig.AffinityGroupIds, ag.(string))
		}
	}

	// Process details
	if details, ok := mp["details"].(map[string]interface{}); ok && len(details) > 0 {
		cloudConfig.Details = make(map[string]string)
		for k, v := range details {
			cloudConfig.Details[k] = v.(string)
		}
	}

	// Process networks
	if networks, ok := mp["network"].([]interface{}); ok && len(networks) > 0 {
		cloudConfig.Networks = make([]*models.V1CloudStackNetworkConfig, 0, len(networks))
		for _, n := range networks {
			network := n.(map[string]interface{})
			netConfig := &models.V1CloudStackNetworkConfig{
				Name: network["network_name"].(string),
				// Note: IP address assignment moved to different level in new SDK
			}
			cloudConfig.Networks = append(cloudConfig.Networks, netConfig)
		}
	}

	poolConfig := &models.V1MachinePoolConfigEntity{
		AdditionalLabels: toAdditionalNodePoolLabels(mp),
		Taints:           toClusterTaints(mp),
		IsControlPlane:   controlPlane,
		Labels:           labels,
		Name:             types.Ptr(mp["name"].(string)),
		Size:             types.Ptr(safeInt32Conversion(mp["count"].(int), 1)),
		UpdateStrategy: &models.V1UpdateStrategy{
			Type: getUpdateStrategy(mp),
		},
		UseControlPlaneAsWorker: controlPlaneAsWorker,
	}

	// Safe conversion for min size
	if mp["min"] != nil {
		minSize := mp["min"].(int)
		if minSize > 0 {
			poolConfig.MinSize = safeInt32Conversion(minSize, 0)
		}
	}

	// Safe conversion for max size
	if mp["max"] != nil {
		maxSize := mp["max"].(int)
		if maxSize > 0 {
			poolConfig.MaxSize = safeInt32Conversion(maxSize, 0)
		}
	}

	mpEntity := &models.V1CloudStackMachinePoolConfigEntity{
		CloudConfig: cloudConfig,
		PoolConfig:  poolConfig,
	}

	return mpEntity
}

func resourceMachinePoolCloudStackHash(v interface{}) int {
	var buf string
	m := v.(map[string]interface{})
	buf = fmt.Sprintf("%s-%t-%t-%s",
		m["name"].(string),
		m["control_plane"].(bool),
		m["control_plane_as_worker"].(bool),
		m["offering"].(string),
	)
	return schema.HashString(buf)
}

func updateMachinePoolCloudStack(ctx context.Context, c *client.V1Client, d *schema.ResourceData, cloudConfigId string) error {
	log.Printf("Updating CloudStack machine pools")

	old, new := d.GetChange("machine_pool")
	oldMachinePools := old.(*schema.Set)
	newMachinePools := new.(*schema.Set)

	// Delete removed machine pools
	for _, old := range oldMachinePools.List() {
		if !newMachinePools.Contains(old) {
			oldMachinePool := old.(map[string]interface{})
			machinePoolName := oldMachinePool["name"].(string)
			log.Printf("Deleting machine pool: %s", machinePoolName)
			if err := c.DeleteMachinePoolCloudStack(cloudConfigId, machinePoolName); err != nil {
				return err
			}
		}
	}

	// Create new machine pools
	for _, new := range newMachinePools.List() {
		if !oldMachinePools.Contains(new) {
			newMachinePool := toMachinePoolCloudStack(new)
			log.Printf("Creating machine pool: %s", *newMachinePool.PoolConfig.Name)
			if err := c.CreateMachinePoolCloudStack(cloudConfigId, newMachinePool); err != nil {
				return err
			}
		}
	}

	// Update existing machine pools
	for _, new := range newMachinePools.List() {
		if oldMachinePools.Contains(new) {
			newMachinePool := toMachinePoolCloudStack(new)
			log.Printf("Updating machine pool: %s", *newMachinePool.PoolConfig.Name)
			if err := c.UpdateMachinePoolCloudStack(cloudConfigId, newMachinePool); err != nil {
				return err
			}
		}
	}

	return nil
}

func resourceClusterCloudStackImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	_, err := GetCommonCluster(d, m)
	if err != nil {
		return nil, err
	}

	diags := resourceClusterCloudStackRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read cluster for import: %v", diags)
	}

	return []*schema.ResourceData{d}, nil
}

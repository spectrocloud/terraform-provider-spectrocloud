package spectrocloud

import (
	"context"
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
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

func resourceClusterApacheCloudStack() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterApacheCloudStackCreate,
		ReadContext:   resourceClusterApacheCloudStackRead,
		UpdateContext: resourceClusterApacheCloudStackUpdate,
		DeleteContext: resourceClusterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceClusterApacheCloudStackImport,
		},
		Description: "Resource for managing Apache CloudStack clusters in Spectro Cloud through Palette.",

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
			"cluster_profile":  schemas.ClusterProfileSchema(),
			"cluster_template": schemas.ClusterTemplateSchema(),
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
			"update_worker_pools_in_parallel": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Controls whether worker pool updates occur in parallel or sequentially. When set to `true`, all worker pools are updated simultaneously. When `false` (default), worker pools are updated one at a time, reducing cluster disruption but taking longer to complete updates.",
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
						"project": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "CloudStack project configuration (optional). If not specified, the cluster will be created in the domain's default project.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "CloudStack project ID.",
									},
									"name": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "CloudStack project name.",
									},
								},
							},
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
									"id": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "CloudStack zone ID. Either `id` or `name` can be used to identify the zone. If both are specified, `id` takes precedence.",
									},
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
												"id": {
													Type:        schema.TypeString,
													Optional:    true,
													Description: "Network ID in CloudStack. Either `id` or `name` can be used to identify the network. If both are specified, `id` takes precedence.",
												},
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
															"id": {
																Type:        schema.TypeString,
																Optional:    true,
																Description: "VPC ID. Either `id` or `name` can be used to identify the VPC. If both are specified, `id` takes precedence.",
															},
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
				Set:      resourceMachinePoolApacheCloudStackHash,
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
							Description: "Apache CloudStack compute offering (instance type/size) name.",
						},
						"instance_config": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Instance configuration details returned by the CloudStack API. This is a computed field based on the selected offering.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"disk_gib": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Root disk size in GiB.",
									},
									"memory_mib": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Memory size in MiB.",
									},
									"num_cpus": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Number of CPUs for the instance.",
									},
									"cpu_set": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "CPU set for the instance.",
									},
									"name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Name for the instance configuration.",
									},
									"category": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Category for the instance configuration.",
									},
								},
							},
						},
						"template": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Apache CloudStack template override for this machine pool. If not specified, inherits cluster default from profile.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Template ID. Either ID or name must be provided.",
									},
									"name": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Template name. Either ID or name must be provided.",
									},
								},
							},
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

func resourceClusterApacheCloudStackCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	resourceClusterApacheCloudStackRead(ctx, d, m)

	return diags
}

func resourceClusterApacheCloudStackRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)
	var diags diag.Diagnostics

	cluster, err := c.GetCluster(d.Id())
	if err != nil {
		return diag.FromErr(err)
	} else if cluster == nil {
		// Deleted - Terraform will recreate it
		d.SetId("")
		return diags
	}

	// Verify cluster type
	err = ValidateCloudType("spectrocloud_cluster_apache_cloudstack", cluster)
	if err != nil {
		return diag.FromErr(err)
	}

	diagnostics, done := readCommonFields(c, d, cluster)
	if done {
		return diagnostics
	}

	// Flatten cluster_template variables using variables API
	if err := flattenClusterTemplateVariables(c, d, d.Id()); err != nil {
		return diag.FromErr(err)
	}

	return flattenCloudConfigApacheCloudStack(cluster.Spec.CloudConfigRef.UID, d, c)
}

func resourceClusterApacheCloudStackUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

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

	resourceClusterApacheCloudStackRead(ctx, d, m)

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
		SSHKeyName:           cloudConfig["ssh_key_name"].(string),
		ControlPlaneEndpoint: cloudConfig["control_plane_endpoint"].(string),
		SyncWithCKS:          cloudConfig["sync_with_cks"].(bool),
	}

	// Process project if specified
	if projects, ok := cloudConfig["project"].([]interface{}); ok && len(projects) > 0 {
		project := projects[0].(map[string]interface{})
		config.Project = &models.V1CloudStackResource{
			ID:   project["id"].(string),
			Name: project["name"].(string),
		}
	}

	// Process zones
	if zones, ok := cloudConfig["zone"].([]interface{}); ok && len(zones) > 0 {
		config.Zones = make([]*models.V1CloudStackZoneSpec, 0, len(zones))
		for _, z := range zones {
			zone := z.(map[string]interface{})
			zoneSpec := &models.V1CloudStackZoneSpec{
				ID:   zone["id"].(string),
				Name: zone["name"].(string),
			}

			// Process network configuration for the zone
			if networks, ok := zone["network"].([]interface{}); ok && len(networks) > 0 {
				network := networks[0].(map[string]interface{})
				zoneSpec.Network = &models.V1CloudStackNetworkSpec{
					ID:          network["id"].(string),
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
						ID:       vpc["id"].(string),
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
		Offering: &models.V1CloudStackResource{
			Name: mp["offering"].(string),
		},
	}

	// Note: instance_config is computed (returned by API based on offering) - not sent in requests

	// Process template (RE-ADDED in new SDK)
	if templates, ok := mp["template"].([]interface{}); ok && len(templates) > 0 {
		tmpl := templates[0].(map[string]interface{})
		cloudConfig.Template = &models.V1CloudStackResource{}
		if id, ok := tmpl["id"].(string); ok && id != "" {
			cloudConfig.Template.ID = id
		}
		if name, ok := tmpl["name"].(string); ok && name != "" {
			cloudConfig.Template.Name = name
		}
	}

	// NOTE: RootDiskSizeGB, DiskOffering, AffinityGroupIds, and Details have been REMOVED from the new SDK model

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

func resourceMachinePoolApacheCloudStackHash(v interface{}) int {
	m := v.(map[string]interface{})
	buf := CommonHash(m)

	// Add CloudStack-specific fields
	if val, ok := m["offering"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", val.(string)))
	}

	// Note: instance_config is computed and excluded from hash to prevent false change detection

	// Hash template
	if templateList, ok := m["template"].([]interface{}); ok && len(templateList) > 0 {
		tmpl := templateList[0].(map[string]interface{})
		if val, ok := tmpl["id"]; ok {
			buf.WriteString(fmt.Sprintf("%s-", val.(string)))
		}
		if val, ok := tmpl["name"]; ok {
			buf.WriteString(fmt.Sprintf("%s-", val.(string)))
		}
	}

	// Hash networks
	if networksList, ok := m["network"].([]interface{}); ok && len(networksList) > 0 {
		var networkNames []string
		for _, n := range networksList {
			network := n.(map[string]interface{})
			if val, ok := network["network_name"]; ok {
				networkNames = append(networkNames, val.(string))
			}
		}
		sort.Strings(networkNames)
		buf.WriteString(strings.Join(networkNames, "-"))
	}

	return int(hash(buf.String()))
}

func updateMachinePoolCloudStack(ctx context.Context, c *client.V1Client, d *schema.ResourceData, cloudConfigId string) error {
	log.Printf("[DEBUG] === MACHINE POOL CHANGE DETECTED ===")

	old, new := d.GetChange("machine_pool")
	oldMachinePools := old.(*schema.Set)
	newMachinePools := new.(*schema.Set)

	log.Printf("[DEBUG] Old machine pools count: %d, New machine pools count: %d", oldMachinePools.Len(), newMachinePools.Len())

	// Create maps by machine pool name for proper comparison
	osMap := make(map[string]interface{})
	for _, mp := range oldMachinePools.List() {
		machinePoolResource := mp.(map[string]interface{})
		name := machinePoolResource["name"].(string)
		if name != "" {
			osMap[name] = machinePoolResource
		}
	}

	nsMap := make(map[string]interface{})
	for _, mp := range newMachinePools.List() {
		machinePoolResource := mp.(map[string]interface{})
		name := machinePoolResource["name"].(string)
		if name != "" {
			nsMap[name] = machinePoolResource

			// Check if this is a new, updated, or unchanged machine pool
			if oldMachinePool, exists := osMap[name]; !exists {
				// NEW machine pool - CREATE
				log.Printf("[DEBUG] Creating new machine pool %s", name)
				machinePool := toMachinePoolCloudStack(machinePoolResource)
				if err := c.CreateMachinePoolCloudStack(cloudConfigId, machinePool); err != nil {
					return err
				}
			} else {
				// EXISTING machine pool - check if hash changed
				oldHash := resourceMachinePoolApacheCloudStackHash(oldMachinePool)
				newHash := resourceMachinePoolApacheCloudStackHash(machinePoolResource)

				if oldHash != newHash {
					// MODIFIED machine pool - UPDATE
					log.Printf("[DEBUG] Updating machine pool %s (hash changed: %d -> %d)", name, oldHash, newHash)
					machinePool := toMachinePoolCloudStack(machinePoolResource)
					if err := c.UpdateMachinePoolCloudStack(cloudConfigId, machinePool); err != nil {
						return err
					}
					// Note: Node maintenance actions are not supported for CloudStack clusters
				} else {
					// UNCHANGED machine pool - no action needed
					log.Printf("[DEBUG] Machine pool %s unchanged (hash: %d)", name, oldHash)
				}
			}

			// Mark as processed
			delete(osMap, name)
		} else {
			log.Printf("[DEBUG] WARNING: Machine pool has empty name!")
		}
	}

	// REMOVED machine pools - DELETE
	for name := range osMap {
		log.Printf("[DEBUG] Deleting removed machine pool %s", name)
		if err := c.DeleteMachinePoolCloudStack(cloudConfigId, name); err != nil {
			return err
		}
	}

	return nil
}

func resourceClusterApacheCloudStackImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	_, err := GetCommonCluster(d, m)
	if err != nil {
		return nil, err
	}

	diags := resourceClusterApacheCloudStackRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read cluster for import: %v", diags)
	}

	return []*schema.ResourceData{d}, nil
}

func flattenCloudConfigApacheCloudStack(configUID string, d *schema.ResourceData, c *client.V1Client) diag.Diagnostics {
	var diags diag.Diagnostics
	if err := d.Set("cloud_config_id", configUID); err != nil {
		return diag.FromErr(err)
	}
	if err := ReadCommonAttributes(d); err != nil {
		return diag.FromErr(err)
	}

	config, err := c.GetCloudConfigCloudStack(configUID)
	if err != nil {
		return diag.FromErr(err)
	}

	if config.Spec != nil && config.Spec.CloudAccountRef != nil {
		if err := d.Set("cloud_account_id", config.Spec.CloudAccountRef.UID); err != nil {
			return diag.FromErr(err)
		}
	}

	if err := d.Set("cloud_config", flattenClusterConfigsApacheCloudStack(config)); err != nil {
		return diag.FromErr(err)
	}

	mp := flattenMachinePoolConfigsApacheCloudStack(config.Spec.MachinePoolConfig)
	mp, err = flattenNodeMaintenanceStatus(c, d, c.GetNodeStatusMapCloudStack, mp, configUID)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("machine_pool", mp); err != nil {
		return diag.FromErr(err)
	}

	generalWarningForRepave(&diags)
	return diags
}

func flattenClusterConfigsApacheCloudStack(config *models.V1CloudStackCloudConfig) []interface{} {
	if config == nil || config.Spec == nil || config.Spec.ClusterConfig == nil {
		return make([]interface{}, 0)
	}

	clusterConfig := config.Spec.ClusterConfig
	m := make(map[string]interface{})

	// Flatten project (V1CloudStackResource)
	if clusterConfig.Project != nil {
		projectMap := make(map[string]interface{})
		if clusterConfig.Project.ID != "" {
			projectMap["id"] = clusterConfig.Project.ID
		}
		if clusterConfig.Project.Name != "" {
			projectMap["name"] = clusterConfig.Project.Name
		}
		m["project"] = []interface{}{projectMap}
	}
	if clusterConfig.SSHKeyName != "" {
		m["ssh_key_name"] = clusterConfig.SSHKeyName
	}
	if clusterConfig.ControlPlaneEndpoint != "" {
		m["control_plane_endpoint"] = clusterConfig.ControlPlaneEndpoint
	}
	m["sync_with_cks"] = clusterConfig.SyncWithCKS

	// Flatten zones
	if len(clusterConfig.Zones) > 0 {
		zones := make([]interface{}, 0, len(clusterConfig.Zones))
		for _, zone := range clusterConfig.Zones {
			zoneMap := make(map[string]interface{})
			if zone.ID != "" {
				zoneMap["id"] = zone.ID
			}
			if zone.Name != "" {
				zoneMap["name"] = zone.Name
			}

			// Flatten network
			if zone.Network != nil {
				network := make(map[string]interface{})
				if zone.Network.ID != "" {
					network["id"] = zone.Network.ID
				}
				if zone.Network.Name != "" {
					network["name"] = zone.Network.Name
				}
				if zone.Network.Type != "" {
					network["type"] = zone.Network.Type
				}
				if zone.Network.Gateway != "" {
					network["gateway"] = zone.Network.Gateway
				}
				if zone.Network.Netmask != "" {
					network["netmask"] = zone.Network.Netmask
				}
				if zone.Network.Offering != "" {
					network["offering"] = zone.Network.Offering
				}
				if zone.Network.RoutingMode != "" {
					network["routing_mode"] = zone.Network.RoutingMode
				}

				// Flatten VPC
				if zone.Network.Vpc != nil {
					vpc := make(map[string]interface{})
					if zone.Network.Vpc.ID != "" {
						vpc["id"] = zone.Network.Vpc.ID
					}
					if zone.Network.Vpc.Name != "" {
						vpc["name"] = zone.Network.Vpc.Name
					}
					if zone.Network.Vpc.Cidr != "" {
						vpc["cidr"] = zone.Network.Vpc.Cidr
					}
					if zone.Network.Vpc.Offering != "" {
						vpc["offering"] = zone.Network.Vpc.Offering
					}
					network["vpc"] = []interface{}{vpc}
				}

				zoneMap["network"] = []interface{}{network}
			}

			zones = append(zones, zoneMap)
		}
		m["zone"] = zones
	}

	return []interface{}{m}
}

func flattenMachinePoolConfigsApacheCloudStack(machinePools []*models.V1CloudStackMachinePoolConfig) []interface{} {
	if machinePools == nil {
		return make([]interface{}, 0)
	}

	ois := make([]interface{}, len(machinePools))

	for i, machinePool := range machinePools {
		oi := make(map[string]interface{})

		// Flatten pool configuration (from V1MachinePoolBaseConfig embedded)
		// Note: AdditionalLabels and Taints are not available in the GET response for CloudStack
		// They're only used during creation. So we skip flattening them to avoid state inconsistencies.
		flattenUpdateStrategy(machinePool.UpdateStrategy, oi)
		if machinePool.IsControlPlane != nil {
			oi["control_plane"] = *machinePool.IsControlPlane
		}
		oi["control_plane_as_worker"] = machinePool.UseControlPlaneAsWorker
		oi["name"] = machinePool.Name
		oi["count"] = int(machinePool.Size)

		if machinePool.MinSize > 0 {
			oi["min"] = int(machinePool.MinSize)
		}
		if machinePool.MaxSize > 0 {
			oi["max"] = int(machinePool.MaxSize)
		}

		// Note: Labels field contains internal cluster-api labels (like "master"), not user-defined node labels
		// User-defined node labels are managed separately through the node schema

		// Flatten machine configuration (from V1CloudStackMachineConfig embedded)
		// Flatten offering
		if machinePool.Offering != nil {
			oi["offering"] = machinePool.Offering.Name
		}

		// Flatten instance_config
		if machinePool.InstanceConfig != nil {
			instanceConfig := make(map[string]interface{})
			instanceConfig["disk_gib"] = int(machinePool.InstanceConfig.DiskGiB)
			instanceConfig["memory_mib"] = int(machinePool.InstanceConfig.MemoryMiB)
			instanceConfig["num_cpus"] = int(machinePool.InstanceConfig.NumCPUs)
			instanceConfig["cpu_set"] = int(machinePool.InstanceConfig.CPUSet)
			if machinePool.InstanceConfig.Name != "" {
				instanceConfig["name"] = machinePool.InstanceConfig.Name
			}
			if machinePool.InstanceConfig.Category != "" {
				instanceConfig["category"] = machinePool.InstanceConfig.Category
			}
			oi["instance_config"] = []interface{}{instanceConfig}
		}

		// Flatten template
		if machinePool.Template != nil {
			template := make(map[string]interface{})
			if machinePool.Template.ID != "" {
				template["id"] = machinePool.Template.ID
			}
			if machinePool.Template.Name != "" {
				template["name"] = machinePool.Template.Name
			}
			oi["template"] = []interface{}{template}
		}

		// Flatten networks
		if len(machinePool.Networks) > 0 {
			networks := make([]interface{}, 0, len(machinePool.Networks))
			for _, network := range machinePool.Networks {
				netMap := make(map[string]interface{})
				if network.Name != "" {
					netMap["network_name"] = network.Name
				}
				networks = append(networks, netMap)
			}
			oi["network"] = networks
		}

		ois[i] = oi
	}

	return ois
}

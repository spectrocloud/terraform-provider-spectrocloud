package spectrocloud

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/gomi/pkg/ptr"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"
)

func resourceClusterNested() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterNestedCreate,
		ReadContext:   resourceClusterNestedRead,
		UpdateContext: resourceClusterNestedUpdate,
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
			"host_cluster_uid": {
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
			"cluster_profile": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"pack": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "spectro",
									},
									"name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"registry_uid": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"tag": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"values": {
										Type:     schema.TypeString,
										Required: true,
									},
									"manifest": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:     schema.TypeString,
													Required: true,
												},
												"content": {
													Type:     schema.TypeString,
													Required: true,
													DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
														// UI strips the trailing newline on save
														return strings.TrimSpace(old) == strings.TrimSpace(new)
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"apply_setting": {
				Type:     schema.TypeString,
				Optional: true,
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
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"chart_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"chart_repo": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"chart_version": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"chart_values": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"k8s_version": {
							Type:     schema.TypeString,
							Optional: true,
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
						"registry_uid": {
							Type:     schema.TypeString,
							Optional: true,
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
				Optional: true,
				Set:      resourceMachinePoolNestedHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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
						"resource_pool": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"backup_policy": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"prefix": {
							Type:     schema.TypeString,
							Required: true,
						},
						"backup_location_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"schedule": {
							Type:     schema.TypeString,
							Required: true,
						},
						"expiry_in_hour": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"include_disks": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"include_cluster_resources": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"namespaces": {
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
			"scan_policy": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"configuration_scan_schedule": {
							Type:     schema.TypeString,
							Required: true,
						},
						"penetration_scan_schedule": {
							Type:     schema.TypeString,
							Required: true,
						},
						"conformance_scan_schedule": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"cluster_rbac_binding": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"namespace": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"role": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"subjects": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:     schema.TypeString,
										Required: true,
									},
									"name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"namespace": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"namespaces": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"resource_allocation": {
							Type:     schema.TypeMap,
							Required: true,
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

func resourceClusterNestedCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	cluster := toNestedCluster(c, d)

	uid, err := c.CreateClusterNested(cluster)
	if err != nil {
		return diag.FromErr(err)
	}

	diagnostics, isError := waitForClusterCreation(ctx, d, uid, diags, c)
	if isError {
		return diagnostics
	}

	resourceClusterNestedRead(ctx, d, m)

	return diags
}

//goland:noinspection GoUnhandledErrorResult
func resourceClusterNestedRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	diagnostics, done := readCommonFields(c, d, cluster)
	if done {
		return diagnostics
	}

	return flattenCloudConfigNested(cluster.Spec.CloudConfigRef.UID, d, c)
}

func flattenCloudConfigNested(configUID string, d *schema.ResourceData, c *client.V1Client) diag.Diagnostics {
	err := d.Set("cloud_config_id", configUID)
	if err != nil {
		return diag.FromErr(err)
	}
	if config, err := c.GetCloudConfigNested(configUID); err != nil {
		return diag.FromErr(err)
	} else {
		mp := flattenMachinePoolConfigsNested(config.Spec.MachinePoolConfig)
		if err := d.Set("machine_pool", mp); err != nil {
			return diag.FromErr(err)
		}
	}

	return diag.Diagnostics{}
}

func flattenMachinePoolConfigsNested(machinePools []*models.V1NestedMachinePoolConfig) []interface{} {

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
		oi["count"] = int(machinePool.Size)
		flattenUpdateStrategy(machinePool.UpdateStrategy, oi)

		oi["resource_pool"] = machinePool.ResourcePool

		ois[i] = oi
	}

	return ois
}

func resourceClusterNestedUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
			hash := resourceMachinePoolNestedHash(machinePoolResource)

			machinePool := toMachinePoolNested(machinePoolResource)

			var err error
			if oldMachinePool, ok := osMap[name]; !ok {
				log.Printf("Create machine pool %s", name)
				err = c.CreateMachinePoolNested(cloudConfigId, machinePool)
			} else if hash != resourceMachinePoolNestedHash(oldMachinePool) {
				log.Printf("Change in machine pool %s", name)
				err = c.UpdateMachinePoolNested(cloudConfigId, machinePool)
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
			if err := c.DeleteMachinePoolNested(cloudConfigId, name); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	diagnostics, done := updateCommonFields(d, c)
	if done {
		return diagnostics
	}

	resourceClusterNestedRead(ctx, d, m)

	return diags
}

func toNestedCluster(c *client.V1Client, d *schema.ResourceData) *models.V1SpectroNestedClusterEntity {
	var chartName, chartRepo, chartVersion, chartValues, kubernetesVersion string

	val, ok := d.GetOk("cloud_config")
	if ok {
		cloudConfig := val.([]interface{})[0].(map[string]interface{})
		chartName = cloudConfig["chart_name"].(string)
		chartRepo = cloudConfig["chart_repo"].(string)
		chartVersion = cloudConfig["chart_version"].(string)
		chartValues = cloudConfig["chart_values"].(string)
		kubernetesVersion = cloudConfig["k8s_version"].(string)
	}

	cluster := &models.V1SpectroNestedClusterEntity{
		Metadata: &models.V1ObjectMeta{
			Name:   d.Get("name").(string),
			UID:    d.Id(),
			Labels: toTags(d),
		},
		Spec: &models.V1SpectroNestedClusterEntitySpec{
			CloudConfig: &models.V1NestedClusterConfig{
				// these values get overridden by the capvc-controller,
				// so it is okay to provide dummy values initially
				ControlPlaneEndpoint: &models.V1APIEndpoint{
					Host: ptr.StringPtr("nested-cluster"),
					Port: ptr.Int32Ptr(443),
				},
				HelmRelease: &models.V1VirtualClusterHelmRelease{
					Chart: &models.V1VirtualClusterHelmChart{
						Name:    chartName,
						Repo:    chartRepo,
						Version: chartVersion,
					},
					Values: chartValues,
				},
				KubernetesVersion: kubernetesVersion,
			},
			ClusterConfig:     toClusterConfig(d),
			HostClusterUID:    d.Get("host_cluster_uid").(string),
			Machinepoolconfig: nil,
			Profiles:          toProfiles(c, d),
			Policies:          toPolicies(d),
		},
	}
	// Hubble raises an error if a nested cluster provides hostClusterConfig
	cluster.Spec.ClusterConfig.HostClusterConfig = nil

	// Specification of machine_pool is optional
	machinePoolConfigs := make([]*models.V1NestedMachinePoolConfigEntity, 0)
	machinePool, ok := d.GetOk("machine_pool")
	if ok {
		for _, machinePool := range machinePool.(*schema.Set).List() {
			mp := toMachinePoolNested(machinePool)
			machinePoolConfigs = append(machinePoolConfigs, mp)
		}
	} else {
		mp := toMachinePoolNested(nil)
		machinePoolConfigs = append(machinePoolConfigs, mp)
	}
	cluster.Spec.Machinepoolconfig = machinePoolConfigs

	return cluster
}

func toMachinePoolNested(machinePool interface{}) *models.V1NestedMachinePoolConfigEntity {
	var controlPlane, controlPlaneAsWorker bool
	var count int
	var name, resourcePool string
	var m map[string]interface{}

	if machinePool == nil {
		controlPlane = true
		controlPlaneAsWorker = true
		count = 1
		name = "nested"
		resourcePool = "nested"
	} else {
		m = machinePool.(map[string]interface{})
		controlPlane = m["control_plane"].(bool)
		controlPlaneAsWorker = m["control_plane_as_worker"].(bool)
		count = m["count"].(int)
		name = m["name"].(string)
		placement := m["placement"].([]interface{})[0].(map[string]interface{})
		resourcePool = placement["resource_pool"].(string)
	}

	labels := make([]string, 0)
	if controlPlane {
		labels = append(labels, "master")
	}

	mp := &models.V1NestedMachinePoolConfigEntity{
		CloudConfig: &models.V1NestedMachinePoolCloudConfigEntity{
			// hardcode for now
			InstanceType: &models.V1NestedInstanceType{
				MinCPU:      2,
				MinMemInMiB: 4096,
			},
			ResourcePool: ptr.StringPtr(resourcePool),
		},
		PoolConfig: &models.V1MachinePoolConfigEntity{
			AdditionalLabels: toAdditionalNodePoolLabels(m),
			Taints:           toClusterTaints(m),
			IsControlPlane:   controlPlane,
			Labels:           labels,
			Name:             ptr.StringPtr(name),
			Size:             ptr.Int32Ptr(int32(count)),
			UpdateStrategy: &models.V1UpdateStrategy{
				Type: getUpdateStrategy(m),
			},
			UseControlPlaneAsWorker: controlPlaneAsWorker,
		},
	}

	return mp
}

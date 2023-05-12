package spectrocloud

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/palette-sdk-go/client"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/schemas"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

func resourceClusterCoxedge() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCoxEdgeClusterCreate,
		ReadContext:   resourceCoxEdgeClusterRead,
		UpdateContext: resourceCoxEdgeClusterUpdate,
		DeleteContext: resourceClusterDelete,

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
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The AWS cloud account id to use for this cluster.",
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
				Description:      "Cron schedule for OS patching. This must be in the form of `0 0 * * *`.",
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
				Type:        schema.TypeList,
				ForceNew:    true,
				Required:    true,
				MaxItems:    1,
				Description: "The Cox Edge environment configuration settings that apply to this cluster.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ssh_keys": {
							Type:     schema.TypeList,
							ForceNew: true,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"lb_config": {
							Type:     schema.TypeList,
							ForceNew: true,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"pops": {
										Type:     schema.TypeList,
										ForceNew: true,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
								},
							},
						},
						"worker_lb": {
							Type:     schema.TypeList,
							ForceNew: true,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"pops": {
										Type:     schema.TypeList,
										ForceNew: true,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
								},
							},
						},
						"organization_id": {
							Type:     schema.TypeString,
							ForceNew: true,
							Optional: true,
						},
						"environment": {
							Type:     schema.TypeString,
							ForceNew: true,
							Optional: true,
						},
					},
				},
			},
			"machine_pool": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "The machine pool configuration for the cluster.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"update_strategy": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "RollingUpdateScaleOut",
							Description: "Update strategy for the machine pool. Valid values are `RollingUpdateScaleOut` and `RollingUpdateScaleIn`.",
						},
						"count": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Number of nodes in the machine pool.",
						},
						"cloud_config": {
							Type:        schema.TypeList,
							ForceNew:    true,
							Required:    true,
							MaxItems:    1,
							Description: "The Cox Edge environment configuration settings that apply to this machine pool.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"spec": {
										Type:     schema.TypeString,
										ForceNew: true,
										Required: true,
									},
									"persistent_storage": {
										Type:     schema.TypeList,
										ForceNew: true,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"path": {
													Type:     schema.TypeString,
													ForceNew: true,
													Required: true,
												},
												"size": {
													Type:     schema.TypeInt,
													ForceNew: true,
													Required: true,
												},
											},
										},
									},
									// Add other fields as needed
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
			"location_config":      schemas.ClusterLocationSchemaComputed(),
			"skip_completion": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "If `true`, the cluster will be created asynchronously. Default value is `false`.",
			},
		},
	}
}

func resourceCoxEdgeClusterCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	cluster := toCoxEdgeCluster(c, d)

	ClusterContext := d.Get("context").(string)
	uid, err := c.CreateClusterCoxEdge(cluster, ClusterContext)
	if err != nil {
		return diag.FromErr(err)
	}

	diagnostics, isError := waitForClusterCreation(ctx, d, uid, diags, c, true)
	if isError {
		return diagnostics
	}

	resourceClusterEksRead(ctx, d, m)

	return diags
}

func resourceCoxEdgeClusterRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	var config *models.V1CoxEdgeCloudConfig
	if config, err = c.GetCloudConfigCoxEdge(configUID); err != nil {
		return diag.FromErr(err)
	}

	cloudConfigFlatten := flattenClusterConfigsCoxEdge(config)
	if err := d.Set("cloud_config", cloudConfigFlatten); err != nil {
		return diag.FromErr(err)
	}

	mp := flattenMachinePoolConfigsCoxEdge(config.Spec.MachinePoolConfig)
	if err := d.Set("machine_pool", mp); err != nil {
		return diag.FromErr(err)
	}

	diagnostics, done := readCommonFields(c, d, cluster)
	if done {
		return diagnostics
	}

	return diags
}

func flattenClusterConfigsCoxEdge(config *models.V1CoxEdgeCloudConfig) []interface{} {
	if config == nil || config.Spec == nil || config.Spec.ClusterConfig == nil {
		return make([]interface{}, 0)
	}

	m := make(map[string]interface{})

	m["ssh_keys"] = config.Spec.ClusterConfig.SSHAuthorizedKeys
	m["organization_id"] = config.Spec.ClusterConfig.OrganizationID
	m["environment"] = config.Spec.ClusterConfig.Environment
	m["lb_config"] = flattenCoxEdgeLoadBalancerConfig(config.Spec.ClusterConfig.CoxEdgeLoadBalancerConfig)
	m["worker_lb"] = flattenCoxEdgeWorkerLoadBalancerConfig(config.Spec.ClusterConfig.CoxEdgeWorkerLoadBalancerConfig)

	return []interface{}{m}
}

func flattenCoxEdgeLoadBalancerConfig(config *models.V1CoxEdgeLoadBalancerConfig) []interface{} {
	if config == nil {
		return make([]interface{}, 0)
	}

	m := make(map[string]interface{})
	m["pops"] = config.Pops

	return []interface{}{m}
}

func flattenCoxEdgeWorkerLoadBalancerConfig(config *models.V1CoxEdgeLoadBalancerConfig) []interface{} {
	if config == nil {
		return make([]interface{}, 0)
	}

	m := make(map[string]interface{})
	m["pops"] = config.Pops

	return []interface{}{m}
}

func flattenMachinePoolConfigsCoxEdge(machinePools []*models.V1CoxEdgeMachinePoolConfig) []interface{} {
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

		// Assuming you have flatten functions for these complex types
		oi["deployments"] = flattenCoxEdgeDeployments(machinePool.Deployments)
		oi["persistent_storage"] = flattenCoxEdgePersistentStorages(machinePool.PersistentStorages)
		oi["s_rules"] = flattenCoxEdgeSecurityGroupRules(machinePool.SecurityGroupRules)

		ois[i] = oi
	}

	return ois
}

func flattenCoxEdgeDeployments(deployments []*models.V1CoxEdgeDeployment) []interface{} {
	if deployments == nil {
		return make([]interface{}, 0)
	}

	ois := make([]interface{}, 0)

	for _, deployment := range deployments {
		oi := make(map[string]interface{})

		// Fill in the map with fields from the deployment.
		oi["name"] = deployment.Name

		ois = append(ois, oi)
	}

	return ois
}

func flattenCoxEdgePersistentStorages(persistentStorages []*models.V1CoxEdgeLoadPersistentStorage) []interface{} {
	if persistentStorages == nil {
		return make([]interface{}, 0)
	}

	ois := make([]interface{}, 0)

	for _, persistentStorage := range persistentStorages {
		oi := make(map[string]interface{})

		// Fill in the map with fields from the persistentStorage.
		oi["name"] = persistentStorage.Path

		ois = append(ois, oi)
	}

	return ois
}

func flattenCoxEdgeSecurityGroupRules(securityGroupRules []*models.V1CoxEdgeSecurityGroupRule) []interface{} {
	if securityGroupRules == nil {
		return make([]interface{}, 0)
	}

	ois := make([]interface{}, 0)

	for _, securityGroupRule := range securityGroupRules {
		oi := make(map[string]interface{})

		// Fill in the map with fields from the securityGroupRule.
		oi["name"] = securityGroupRule.Type

		ois = append(ois, oi)
	}

	return ois
}

func resourceCoxEdgeClusterUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	cloudConfigId := d.Get("cloud_config_id").(string)

	_ = d.Get("machine_pool")

	if d.HasChange("machine_pool") {
		oraw, nraw := d.GetChange("machine_pool")
		if oraw == nil {
			oraw = new(schema.Set)
		}
		if nraw == nil {
			nraw = new(schema.Set)
		}

		os := oraw.([]interface{})
		ns := nraw.([]interface{})

		osMap := make(map[string]interface{})
		for _, mp := range os {
			machinePool := mp.(map[string]interface{})
			osMap[machinePool["name"].(string)] = machinePool
		}

		for _, mp := range ns {
			machinePoolResource := mp.(map[string]interface{})
			name := machinePoolResource["name"].(string)
			hash := resourceMachinePoolCoxEdgeHash(machinePoolResource)

			machinePool := toMachinePoolCoxEdge(machinePoolResource)

			var err error
			if oldMachinePool, ok := osMap[name]; !ok {
				log.Printf("Create machine pool %s", name)
				err = c.CreateMachinePoolCoxEdge(cloudConfigId, machinePool)
			} else if hash != resourceMachinePoolCoxEdgeHash(oldMachinePool) {
				// TODO
				log.Printf("Change in machine pool %s", name)
				err = c.UpdateMachinePoolCoxEdge(cloudConfigId, machinePool)
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
			if err := c.DeleteMachinePoolCoxEdge(cloudConfigId, name); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	diagnostics, done := updateCommonFields(d, c)
	if done {
		return diagnostics
	}

	resourceCoxEdgeClusterRead(ctx, d, m)

	return diags
}

func toCoxEdgeCluster(c *client.V1Client, d *schema.ResourceData) *models.V1SpectroCoxEdgeClusterEntity {
	cloudConfig := d.Get("cloud_config").([]interface{})[0].(map[string]interface{})

	machinePoolConfigs := make([]*models.V1CoxEdgeMachinePoolConfigEntity, 0)
	for _, machinePool := range d.Get("machine_pool").([]interface{}) {
		mp := toMachinePoolCoxEdge(machinePool)
		machinePoolConfigs = append(machinePoolConfigs, mp)
	}

	var LoadBalancer *models.V1CoxEdgeLoadBalancerConfig
	var WorkerLoadBalancer *models.V1CoxEdgeLoadBalancerConfig
	if cloudConfig["lb_config"] != nil {
		lbConfig := cloudConfig["lb_config"].([]interface{})
		if len(lbConfig) > 0 {
			LoadBalancer = &models.V1CoxEdgeLoadBalancerConfig{
				Pops: []string{lbConfig[0].(map[string]interface{})["pops"].([]interface{})[0].(string)},
			}
		}
	}
	if cloudConfig["worker_lb"] != nil {
		lbConfig := cloudConfig["worker_lb"].([]interface{})
		WorkerLoadBalancer = &models.V1CoxEdgeLoadBalancerConfig{
			Pops: []string{lbConfig[0].(map[string]interface{})["pops"].([]interface{})[0].(string)},
		}
	}

	cluster := &models.V1SpectroCoxEdgeClusterEntity{
		Metadata: &models.V1ObjectMeta{
			Name:   d.Get("name").(string),
			UID:    d.Id(),
			Labels: toTags(d),
		},
		Spec: &models.V1SpectroCoxEdgeClusterEntitySpec{
			CloudAccountUID: types.Ptr(d.Get("cloud_account_id").(string)),
			CloudType:       types.Ptr("coxedge"),
			Profiles:        toProfiles(c, d),
			Policies:        toPolicies(d),
			CloudConfig: &models.V1CoxEdgeClusterConfig{
				CoxEdgeLoadBalancerConfig:       LoadBalancer,
				CoxEdgeWorkerLoadBalancerConfig: WorkerLoadBalancer,
				SSHAuthorizedKeys:               []string{cloudConfig["ssh_keys"].([]interface{})[0].(string)},
				OrganizationID:                  cloudConfig["organization_id"].(string),
				Environment:                     cloudConfig["environment"].(string),
			},
			Machinepoolconfig: machinePoolConfigs,
		},
	}

	return cluster
}

func toMachinePoolCoxEdge(machinePool interface{}) *models.V1CoxEdgeMachinePoolConfigEntity {
	m := machinePool.(map[string]interface{})

	labels := make([]string, 0)
	controlPlane, _ := m["control_plane"].(bool)
	controlPlaneAsWorker := m["control_plane_as_worker"].(bool)
	if controlPlane {
		labels = append(labels, "master")
	}

	deployments := make([]*models.V1CoxEdgeDeployment, 0)
	if m["deployments"] != nil {
		for _, deployment := range m["deployments"].([]interface{}) {
			deployments = append(deployments, toCoxEdgeDeployment(deployment))
		}
	}

	persistentStorages := make([]*models.V1CoxEdgeLoadPersistentStorage, 0)
	if m["persistent_storage"] != nil {
		for _, persistentStorage := range m["persistent_storage"].([]interface{}) {
			persistentStorages = append(persistentStorages, toCoxEdgePersistentStorage(persistentStorage))
		}
	}

	securityGroupRules := make([]*models.V1CoxEdgeSecurityGroupRule, 0)
	if m["s_rules"] != nil {
		for _, securityGroupRule := range m["s_rules"].([]interface{}) {
			securityGroupRules = append(securityGroupRules, toCoxEdgeSecurityGroupRule(securityGroupRule))
		}
	}

	mp := &models.V1CoxEdgeMachinePoolConfigEntity{
		CloudConfig: &models.V1CoxEdgeMachinePoolCloudConfigEntity{
			Deployments:        deployments,
			PersistentStorages: persistentStorages,
			SecurityGroupRules: securityGroupRules,
			Spec:               m["cloud_config"].([]interface{})[0].(map[string]interface{})["spec"].(string),
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

func toCoxEdgeDeployment(deployment interface{}) *models.V1CoxEdgeDeployment {
	d := deployment.(map[string]interface{})

	return &models.V1CoxEdgeDeployment{
		Name: d["name"].(string),
	}
}

func toCoxEdgePersistentStorage(persistentStorage interface{}) *models.V1CoxEdgeLoadPersistentStorage {

	return &models.V1CoxEdgeLoadPersistentStorage{}
}

func toCoxEdgeSecurityGroupRule(securityGroupRule interface{}) *models.V1CoxEdgeSecurityGroupRule {
	sgr := securityGroupRule.(map[string]interface{})

	return &models.V1CoxEdgeSecurityGroupRule{
		Protocol: sgr["protocol"].(string),
	}
}

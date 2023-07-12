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

func resourceClusterCoxEdge() *schema.Resource {
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
				Description: "The CoxEdge cloud account id to use for this cluster.",
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
							Description: "Name of the machine pool. This must be unique within the cluster. ",
						},
						"update_strategy": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "RollingUpdateScaleOut",
							Description:  "Update strategy for the machine pool. Valid values are `RollingUpdateScaleOut` and `RollingUpdateScaleIn`.",
							ValidateFunc: validation.StringInSlice([]string{"RollingUpdateScaleOut", "RollingUpdateScaleIn"}, false),
						},
						"count": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Number of nodes in the machine pool.",
						},
						"cox_config": {
							Type:        schema.TypeList,
							ForceNew:    true,
							Required:    true,
							MaxItems:    1,
							Description: "The Cox Edge environment configuration settings that apply to this machine pool.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"spec": {
										Type:        schema.TypeString,
										ForceNew:    true,
										Required:    true,
										Description: "The Cox Edge environment configuration settings that apply to the machine pool.",
									},
									"persistent_storage": {
										Type:     schema.TypeList,
										ForceNew: true,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"path": {
													Type:        schema.TypeString,
													ForceNew:    true,
													Required:    true,
													Description: "Mount path for the persistent storage. ",
												},
												"size": {
													Type:        schema.TypeInt,
													ForceNew:    true,
													Required:    true,
													Description: "Size of the persistent storage in GB. ",
												},
											},
										},
									},
									"deployments": {
										Type:        schema.TypeList,
										Optional:    true,
										Description: "The deployments associated with this machine pool.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:        schema.TypeString,
													Optional:    true,
													Description: "The name of the deployment. ",
												},
												"instances_per_pop": {
													Type:        schema.TypeInt,
													Optional:    true,
													Description: "The number of instances per pop. ",
												},
												"pops": {
													Type:        schema.TypeList,
													Elem:        &schema.Schema{Type: schema.TypeString},
													Optional:    true,
													Description: "The pops to deploy to. ",
												},
											},
										},
									},
									"security_group_rules": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"action": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringInSlice([]string{"allow", "deny"}, false),
													Description:  "Action for the rule, 'allow' or 'deny'.",
												},
												"description": {
													Type:        schema.TypeString,
													Optional:    true,
													Description: "Description of what this rule is used for.",
												},
												"port_range": {
													Type:        schema.TypeString,
													Optional:    true,
													Description: "Port range for the rule. Could be a single port or a range like '80-433'.",
												},
												"protocol": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringInSlice([]string{"TCP", "UDP"}, false),
													Description:  "Protocol for the rule, for example 'TCP' or 'UDP'.",
												},
												"source": {
													Type:        schema.TypeString,
													Optional:    true,
													Description: "Source IP range for the rule, in CIDR notation.",
												},
												"type": {
													Type:        schema.TypeString,
													Optional:    true,
													Description: "Type of the rule 'inbound' or 'outbound'.",
												},
											},
										},
										Description: "List of security group rules that apply to this cox_config.",
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

	cluster, err := toCoxEdgeCluster(c, d)
	if err != nil {
		return diag.FromErr(err)
	}

	ClusterContext := d.Get("context").(string)
	uid, err := c.CreateClusterCoxEdge(cluster, ClusterContext)
	if err != nil {
		return diag.FromErr(err)
	}

	diagnostics, isError := waitForClusterCreation(ctx, d, ClusterContext, uid, diags, c, true)
	if isError {
		return diagnostics
	}

	resourceClusterEksRead(ctx, d, m)

	return diags
}

func resourceCoxEdgeClusterRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	configUID := cluster.Spec.CloudConfigRef.UID
	if err := d.Set("cloud_config_id", configUID); err != nil {
		return diag.FromErr(err)
	}

	var config *models.V1CoxEdgeCloudConfig
	ClusterContext := d.Get("context").(string)
	if config, err = c.GetCloudConfigCoxEdge(configUID, ClusterContext); err != nil {
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
		if machinePool == nil {
			continue
		}

		oi := make(map[string]interface{})

		SetAdditionalLabelsAndTaints(machinePool.AdditionalLabels, machinePool.Taints, oi)

		oi["control_plane"] = machinePool.IsControlPlane
		oi["control_plane_as_worker"] = machinePool.UseControlPlaneAsWorker
		oi["name"] = machinePool.Name
		oi["count"] = int(machinePool.Size)
		flattenUpdateStrategy(machinePool.UpdateStrategy, oi)

		coxConfig := make(map[string]interface{})
		coxConfig["spec"] = machinePool.Spec
		coxConfig["persistent_storage"] = flattenCoxEdgePersistentStorages(machinePool.PersistentStorages)
		coxConfig["deployments"] = flattenCoxEdgeDeployments(machinePool.Deployments)
		coxConfig["security_group_rules"] = flattenCoxEdgeSecurityGroupRules(machinePool.SecurityGroupRules)

		oi["cox_config"] = []interface{}{coxConfig}

		ois[i] = oi
	}

	return ois
}

func flattenCoxEdgeDeployments(deployments []*models.V1CoxEdgeDeployment) []interface{} {
	if deployments == nil {
		return make([]interface{}, 0)
	}

	ois := make([]interface{}, len(deployments))

	for i, deployment := range deployments {
		oi := make(map[string]interface{})

		oi["name"] = deployment.Name

		oi["instances_per_pop"] = deployment.InstancesPerPop

		if deployment.Pops != nil {
			oi["pops"] = deployment.Pops
		}

		ois[i] = oi
	}

	return ois
}

func flattenCoxEdgePersistentStorages(persistentStorages []*models.V1CoxEdgeLoadPersistentStorage) []interface{} {
	if persistentStorages == nil {
		return make([]interface{}, 0)
	}

	ois := make([]interface{}, len(persistentStorages))

	for i, persistentStorage := range persistentStorages {
		oi := make(map[string]interface{})

		oi["path"] = persistentStorage.Path
		oi["size"] = persistentStorage.Size

		ois[i] = oi
	}

	return ois
}

func flattenCoxEdgeSecurityGroupRules(securityGroupRules []*models.V1CoxEdgeSecurityGroupRule) []interface{} {
	if securityGroupRules == nil {
		return make([]interface{}, 0)
	}

	ois := make([]interface{}, len(securityGroupRules))

	for i, rule := range securityGroupRules {
		if rule == nil {
			continue
		}

		oi := make(map[string]interface{})

		oi["action"] = rule.Action
		oi["description"] = rule.Description
		oi["port_range"] = rule.PortRange
		oi["protocol"] = rule.Protocol
		oi["source"] = rule.Source
		oi["type"] = rule.Type

		ois[i] = oi
	}

	return ois
}

func resourceCoxEdgeClusterUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

		os := oraw.([]interface{})
		ns := nraw.([]interface{})

		osMap := make(map[string]interface{})
		for _, mp := range os {
			machinePool := mp.(map[string]interface{})
			osMap[machinePool["name"].(string)] = machinePool
		}

		for _, mp := range ns {
			machinePoolResource := mp.(map[string]interface{})
			// since known issue in TF SDK: https://github.com/hashicorp/terraform-plugin-sdk/issues/588
			if machinePoolResource["name"].(string) != "" {
				name := machinePoolResource["name"].(string)
				hash := resourceMachinePoolCoxEdgeHash(machinePoolResource)

				machinePool, err := toMachinePoolCoxEdge(machinePoolResource)
				if err != nil {
					return diag.FromErr(err)
				}

				if oldMachinePool, ok := osMap[name]; !ok {
					log.Printf("Create machine pool %s", name)
					err = c.CreateMachinePoolCoxEdge(cloudConfigId, machinePool, ClusterContext)
				} else if hash != resourceMachinePoolCoxEdgeHash(oldMachinePool) {
					// TODO
					log.Printf("Change in machine pool %s", name)
					err = c.UpdateMachinePoolCoxEdge(cloudConfigId, machinePool, ClusterContext)
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
			if err := c.DeleteMachinePoolCoxEdge(cloudConfigId, name, ClusterContext); err != nil {
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

func toCoxEdgeCluster(c *client.V1Client, d *schema.ResourceData) (*models.V1SpectroCoxEdgeClusterEntity, error) {
	cloudConfig := d.Get("cloud_config").([]interface{})[0].(map[string]interface{})

	machinePoolConfigs := make([]*models.V1CoxEdgeMachinePoolConfigEntity, 0)
	for _, machinePool := range d.Get("machine_pool").([]interface{}) {
		mp, err := toMachinePoolCoxEdge(machinePool)
		if err != nil {
			return nil, err
		}
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

	return cluster, nil
}

func toMachinePoolCoxEdge(machinePool interface{}) (*models.V1CoxEdgeMachinePoolConfigEntity, error) {
	m := machinePool.(map[string]interface{})

	labels := make([]string, 0)
	controlPlane, _ := m["control_plane"].(bool)
	controlPlaneAsWorker := m["control_plane_as_worker"].(bool)
	if controlPlane {
		labels = append(labels, "master")
	}

	deployments := make([]*models.V1CoxEdgeDeployment, 0)
	persistentStorages := make([]*models.V1CoxEdgeLoadPersistentStorage, 0)
	securityGroupRules := make([]*models.V1CoxEdgeSecurityGroupRule, 0)

	if m["cox_config"] != nil && m["cox_config"].([]interface{})[0] != nil {
		CoxConfig := m["cox_config"].([]interface{})[0]

		if CoxConfig.(map[string]interface{})["deployments"] != nil {
			for _, deployment := range CoxConfig.(map[string]interface{})["deployments"].([]interface{}) {
				deployments = append(deployments, toCoxEdgeDeployment(deployment))
			}
		}

		if CoxConfig.(map[string]interface{})["persistent_storage"] != nil {
			for _, persistentStorage := range CoxConfig.(map[string]interface{})["persistent_storage"].([]interface{}) {
				storage, err := toCoxEdgePersistentStorage(persistentStorage)
				if err != nil {
					return nil, err
				}
				persistentStorages = append(persistentStorages, storage)
			}
		}
		SecGroups := CoxConfig.(map[string]interface{})["security_group_rules"]
		// get security groups
		if SecGroups != nil {
			for _, securityGroupRule := range SecGroups.([]interface{}) {
				securityGroupRules = append(securityGroupRules, toCoxEdgeSecurityGroupRule(securityGroupRule))
			}
		}
	}

	mp := &models.V1CoxEdgeMachinePoolConfigEntity{
		CloudConfig: &models.V1CoxEdgeMachinePoolCloudConfigEntity{
			Deployments:        deployments,
			PersistentStorages: persistentStorages,
			SecurityGroupRules: securityGroupRules,
			Spec:               m["cox_config"].([]interface{})[0].(map[string]interface{})["spec"].(string),
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

	return mp, nil
}

func toCoxEdgeDeployment(deployment interface{}) *models.V1CoxEdgeDeployment {
	d := deployment.(map[string]interface{})

	popsInterface, ok := d["pops"].([]interface{})
	if !ok {
		popsInterface = make([]interface{}, 0)
	}

	pops := make([]string, len(popsInterface))
	for i, v := range popsInterface {
		pops[i] = v.(string)
	}

	instancesPerPop := 0
	if val, ok := d["instances_per_pop"]; ok {
		instancesPerPop = val.(int)
	}

	return &models.V1CoxEdgeDeployment{
		Name:            d["name"].(string),
		Pops:            pops,
		InstancesPerPop: int32(instancesPerPop),
	}
}

func toCoxEdgePersistentStorage(persistentStorage interface{}) (*models.V1CoxEdgeLoadPersistentStorage, error) {
	path := persistentStorage.(map[string]interface{})["path"].(string)
	size := persistentStorage.(map[string]interface{})["size"].(int)
	return &models.V1CoxEdgeLoadPersistentStorage{
		Path: path,
		Size: int64(size),
	}, nil
}

func toCoxEdgeSecurityGroupRule(securityGroupRule interface{}) *models.V1CoxEdgeSecurityGroupRule {
	sgr := securityGroupRule.(map[string]interface{})

	coxEdgeSecurityGroupRule := &models.V1CoxEdgeSecurityGroupRule{}

	if protocol, ok := sgr["protocol"].(string); ok {
		coxEdgeSecurityGroupRule.Protocol = protocol
	}
	if portRange, ok := sgr["port_range"].(string); ok {
		coxEdgeSecurityGroupRule.PortRange = portRange
	}
	if action, ok := sgr["action"].(string); ok {
		coxEdgeSecurityGroupRule.Action = action
	}
	if source, ok := sgr["source"].(string); ok {
		coxEdgeSecurityGroupRule.Source = source
	}
	if description, ok := sgr["description"].(string); ok {
		coxEdgeSecurityGroupRule.Description = description
	}
	if typ, ok := sgr["type"].(string); ok {
		coxEdgeSecurityGroupRule.Type = typ
	}

	return coxEdgeSecurityGroupRule
}

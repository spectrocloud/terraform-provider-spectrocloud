package spectrocloud

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/gomi/pkg/ptr"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"
)

func resourceClusterTke() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterTkeCreate,
		ReadContext:   resourceClusterTkeRead,
		UpdateContext: resourceClusterTkeUpdate,
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
			"cluster_profile": {
				Type:          schema.TypeList,
				Optional:      true,
				ConflictsWith: []string{"pack"},
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
									"registry_uid": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"tag": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"values": {
										Type:     schema.TypeString,
										Optional: true,
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
														if strings.TrimSpace(old) == strings.TrimSpace(new) {
															return true
														}
														return false
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
							ForceNew: true,
							Optional: true,
						},
						"region": {
							Type:     schema.TypeString,
							ForceNew: true,
							Required: true,
						},
						"vpc_id": {
							Type:     schema.TypeString,
							ForceNew: true,
							Optional: true,
						},
						"azs": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"az_subnets": {
							Type:     schema.TypeMap,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"endpoint_access": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice([]string{"public", "private", "private_and_public"}, false),
							Default:      "public",
						},
						"public_access_cidrs": {
							Type:     schema.TypeSet,
							Optional: true,
							ForceNew: true,
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
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
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
						"disk_size_gb": {
							Type:     schema.TypeInt,
							Required: true,
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
						"min": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"max": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"instance_type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"capacity_type": {
							Type:     schema.TypeString,
							Default:  "on-demand",
							Optional: true,
						},
						"max_price": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"azs": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"az_subnets": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
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

func resourceClusterTkeCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	var diags diag.Diagnostics

	cluster := toTkeCluster(d)

	uid, err := c.CreateClusterTke(cluster)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(uid)

	if _, found := toTags(d)["skip_completion"]; found {
		return diags
	}

	stateConf := &resource.StateChangeConf{
		Pending:    resourceClusterCreatePendingStates,
		Target:     []string{"Running"},
		Refresh:    resourceClusterStateRefreshFunc(c, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutCreate) - 1*time.Minute,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	resourceClusterTkeRead(ctx, d, m)

	return diags
}

func resourceClusterTkeRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	var diags diag.Diagnostics

	uid := d.Id()

	cluster, err := c.GetCluster(uid)
	if err != nil {
		return diag.FromErr(err)
	} else if cluster == nil {
		d.SetId("")
		return diags
	}

	configUID := cluster.Spec.CloudConfigRef.UID
	d.Set("cloud_config_id", configUID)

	var config *models.V1TencentCloudConfig
	if config, err = c.GetCloudConfigTke(configUID); err != nil {
		return diag.FromErr(err)
	}
	mp := flattenMachinePoolConfigsTke(config.Spec.MachinePoolConfig)
	if err := d.Set("machine_pool", mp); err != nil {
		return diag.FromErr(err)
	}

	diagnostics, done := readCommonFields(c, d, cluster)
	if done {
		return diagnostics
	}

	return diags
}

func flattenMachinePoolConfigsTke(machinePools []*models.V1TencentMachinePoolConfig) []interface{} {

	if machinePools == nil {
		return make([]interface{}, 0)
	}

	ois := make([]interface{}, 0)

	for _, machinePool := range machinePools {
		oi := make(map[string]interface{})

		if machinePool.AdditionalLabels == nil || len(machinePool.AdditionalLabels) == 0 {
			oi["additional_labels"] = make(map[string]interface{})
		} else {
			oi["additional_labels"] = machinePool.AdditionalLabels
		}

		if machinePool.IsControlPlane {
			continue
		}

		oi["name"] = machinePool.Name
		oi["count"] = int(machinePool.Size)
		oi["min"] = int(machinePool.MinSize)
		oi["max"] = int(machinePool.MaxSize)
		oi["instance_type"] = machinePool.InstanceType
		oi["disk_size_gb"] = int(machinePool.RootDeviceSize)
		oi["azs"] = machinePool.Azs

		ois = append(ois, oi)
	}

	return ois
}

func resourceClusterTkeUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

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
			hash := resourceMachinePoolTkeHash(machinePoolResource)

			machinePool := toMachinePoolTke(machinePoolResource)

			var err error
			if oldMachinePool, ok := osMap[name]; !ok {
				log.Printf("Create machine pool %s", name)
				err = c.CreateMachinePoolTke(cloudConfigId, machinePool)
			} else if hash != resourceMachinePoolTkeHash(oldMachinePool) {
				log.Printf("Change in machine pool %s", name)
				err = c.UpdateMachinePoolTke(cloudConfigId, machinePool)
			}

			if err != nil {
				return diag.FromErr(err)
			}

			delete(osMap, name)
		}

		for _, mp := range osMap {
			machinePool := mp.(map[string]interface{})
			name := machinePool["name"].(string)
			log.Printf("Deleted machine pool %s", name)
			if err := c.DeleteMachinePoolTke(cloudConfigId, name); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	diagnostics, done := updateCommonFields(d, c)
	if done {
		return diagnostics
	}

	resourceClusterTkeRead(ctx, d, m)

	return diags
}

func toTkeCluster(d *schema.ResourceData) *models.V1SpectroTencentClusterEntity {
	cloudConfig := d.Get("cloud_config").([]interface{})[0].(map[string]interface{})
	sshKeyIds := make([]string, 0)

	if cloudConfig["ssh_key_name"] != nil {
		sshKeyIds = append(sshKeyIds, cloudConfig["ssh_key_name"].(string))
	}

	cluster := &models.V1SpectroTencentClusterEntity{
		Metadata: &models.V1ObjectMeta{
			Name:   d.Get("name").(string),
			UID:    d.Id(),
			Labels: toTags(d),
		},
		Spec: &models.V1SpectroTencentClusterEntitySpec{
			CloudAccountUID: ptr.StringPtr(d.Get("cloud_account_id").(string)),
			Profiles:        toProfiles(d),
			Policies:        toPolicies(d),
			CloudConfig: &models.V1TencentClusterConfig{
				VpcID:     cloudConfig["vpc_id"].(string),
				Region:    ptr.StringPtr(cloudConfig["region"].(string)),
				SSHKeyIDs: sshKeyIds,
			},
		},
	}

	access := &models.V1TkeEndpointAccess{}
	switch cloudConfig["endpoint_access"].(string) {
	case "public":
		access.Public = true
		access.Private = false
	case "private":
		access.Public = false
		access.Private = true
	case "private_and_public":
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

	cluster.Spec.CloudConfig.EndpointAccess = access

	machinePoolConfigs := make([]*models.V1TencentMachinePoolConfigEntity, 0)
	/*cpPool := map[string]interface{}{
		"control_plane": true,
		"name":          "master-pool",
		"az_subnets":    cloudConfig["az_subnets"],
		"instance_type": "S3.LARGE8",
		"disk_size_gb":  60,
		"count":         2,
	}
	machinePoolConfigs = append(machinePoolConfigs, toMachinePoolTke(cpPool))*/
	for _, machinePool := range d.Get("machine_pool").([]interface{}) {
		mp := toMachinePoolTke(machinePool)
		machinePoolConfigs = append(machinePoolConfigs, mp)
	}

	cluster.Spec.Machinepoolconfig = machinePoolConfigs
	cluster.Spec.ClusterConfig = &models.V1ClusterConfigEntity{
		Resources: toClusterResourceConfig(d),
	}

	return cluster
}

func toMachinePoolTke(machinePool interface{}) *models.V1TencentMachinePoolConfigEntity {
	m := machinePool.(map[string]interface{})

	labels := make([]string, 0)
	controlPlane, _ := m["control_plane"].(bool)
	if controlPlane {
		labels = append(labels, "master")
	}

	azs := make([]string, 0)
	for k, _ := range m["az_subnets"].(map[string]interface{}) {
		azs = append(azs, k)
	}

	min := int32(m["count"].(int))
	max := int32(m["count"].(int))

	if m["min"] != nil {
		min = int32(m["min"].(int))
	}

	if m["max"] != nil {
		max = int32(m["max"].(int))
	}

	mp := &models.V1TencentMachinePoolConfigEntity{
		CloudConfig: &models.V1TencentMachinePoolCloudConfigEntity{
			RootDeviceSize: int64(m["disk_size_gb"].(int)),
			InstanceType:   m["instance_type"].(string),
			Azs:            azs,
		},
		PoolConfig: &models.V1MachinePoolConfigEntity{
			AdditionalLabels: toAdditionalNodePoolLabels(m),
			Taints:           toClusterTaints(m),
			IsControlPlane:   controlPlane,
			Labels:           labels,
			Name:             ptr.StringPtr(m["name"].(string)),
			Size:             ptr.Int32Ptr(int32(m["count"].(int))),
			MinSize:          min,
			MaxSize:          max,
		},
	}

	return mp
}

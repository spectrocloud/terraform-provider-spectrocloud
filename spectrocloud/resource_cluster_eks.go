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

func resourceClusterEks() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterEksCreate,
		ReadContext:   resourceClusterEksRead,
		UpdateContext: resourceClusterEksUpdate,
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
			"cluster_profile_id": {
				Type:       schema.TypeString,
				Optional:   true,
				Deprecated: "Switch to cluster_profile",
			},
			"cluster_profile": {
				Type:          schema.TypeList,
				Optional:      true,
				ConflictsWith: []string{"cluster_profile_id", "pack"},
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
														// UI strips the trailing newline on save
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
						"disk_size_gb": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"count": {
							Type:     schema.TypeInt,
							Required: true,
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
			"fargate_profile": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"subnets": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"additional_tags": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"selector": {
							Type:     schema.TypeList,
							Required: true,
							//MinItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"namespace": {
										Type:     schema.TypeString,
										Required: true,
									},
									"labels": {
										Type:     schema.TypeMap,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
								},
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

func resourceClusterEksCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	cluster := toEksCluster(d)

	uid, err := c.CreateClusterEks(cluster)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(uid)

	stateConf := &resource.StateChangeConf{
		Pending:    resourceClusterCreatePendingStates,
		Target:     []string{"Running"},
		Refresh:    resourceClusterStateRefreshFunc(c, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutCreate) - 1*time.Minute,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	// Wait, catching any errors
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	resourceClusterEksRead(ctx, d, m)

	return diags
}

//goland:noinspection GoUnhandledErrorResult
func resourceClusterEksRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
	d.Set("cloud_config_id", configUID)

	if err := d.Set("tags", flattenTags(cluster.Metadata.Labels)); err != nil {
		return diag.FromErr(err)
	}

	var config *models.V1EksCloudConfig
	if config, err = c.GetCloudConfigEks(configUID); err != nil {
		return diag.FromErr(err)
	}

	// Update the kubeconfig
	kubecfg, err := c.GetClusterKubeConfig(uid)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("kubeconfig", kubecfg); err != nil {
		return diag.FromErr(err)
	}

	mp := flattenMachinePoolConfigsEks(config.Spec.MachinePoolConfig)
	if err := d.Set("machine_pool", mp); err != nil {
		return diag.FromErr(err)
	}

	fp := flattenFargateProfilesEks(config.Spec.FargateProfiles)
	if err := d.Set("fargate_profile", fp); err != nil {
		return diag.FromErr(err)
	}

	//read backup policy and scan policy
	if policy, err := c.GetClusterBackupConfig(d.Id()); err != nil {
		return diag.FromErr(err)
	} else if policy != nil && policy.Spec.Config != nil {
		if err := d.Set("backup_policy", flattenBackupPolicy(policy.Spec.Config)); err != nil {
			return diag.FromErr(err)
		}
	}

	if policy, err := c.GetClusterScanConfig(d.Id()); err != nil {
		return diag.FromErr(err)
	} else if policy != nil && policy.Spec.DriverSpec != nil {
		if err := d.Set("scan_policy", flattenScanPolicy(policy.Spec.DriverSpec)); err != nil {
			return diag.FromErr(err)
		}
	}

	if rbac, err := c.GetClusterRbacConfig(d.Id()); err != nil {
		return diag.FromErr(err)
	} else if rbac != nil && rbac.Items != nil {
		if err := d.Set("cluster_rbac_binding", flattenClusterRBAC(rbac.Items)); err != nil {
			return diag.FromErr(err)
		}
	}

	if namespace, err := c.GetClusterNamespaceConfig(d.Id()); err != nil {
		return diag.FromErr(err)
	} else if namespace != nil && namespace.Items != nil {
		if err := d.Set("namespaces", flattenClusterNamespaces(namespace.Items)); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func flattenMachinePoolConfigsEks(machinePools []*models.V1EksMachinePoolConfig) []interface{} {

	if machinePools == nil {
		return make([]interface{}, 0)
	}

	ois := make([]interface{}, 0)

	for _, machinePool := range machinePools {
		oi := make(map[string]interface{})

		oi["additional_labels"] = machinePool.AdditionalLabels
		oi["taints"] = flattenClusterTaints(machinePool.Taints)

		if *machinePool.IsControlPlane {
			continue
		}

		oi["name"] = machinePool.Name
		oi["count"] = int(machinePool.Size)
		oi["min"] = int(machinePool.MinSize)
		oi["max"] = int(machinePool.MaxSize)
		oi["instance_type"] = machinePool.InstanceType
		if machinePool.CapacityType != nil {
			oi["capacity_type"] = machinePool.CapacityType
		}
		if machinePool.SpotMarketOptions != nil {
			oi["max_price"] = machinePool.SpotMarketOptions.MaxPrice
		}
		oi["disk_size_gb"] = int(machinePool.RootDeviceSize)
		if len(machinePool.SubnetIds) > 0 {
			oi["az_subnets"] = machinePool.SubnetIds
		} else {
			oi["azs"] = machinePool.Azs
		}

		ois = append(ois, oi)
	}

	return ois
}

func flattenFargateProfilesEks(fargateProfiles []*models.V1FargateProfile) []interface{} {

	if fargateProfiles == nil {
		return make([]interface{}, 0)
	}

	ois := make([]interface{}, 0)

	for _, fargateProfile := range fargateProfiles {
		oi := make(map[string]interface{})

		oi["name"] = fargateProfile.Name
		oi["subnets"] = fargateProfile.SubnetIds
		oi["additional_tags"] = fargateProfile.AdditionalTags

		selectors := make([]interface{}, 0)
		for _, selector := range fargateProfile.Selectors {
			s := make(map[string]interface{})
			s["namespace"] = selector.Namespace
			s["labels"] = selector.Labels
			selectors = append(selectors, s)
		}
		oi["selector"] = selectors

		ois = append(ois, oi)
	}

	return ois
}

func resourceClusterEksUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	cloudConfigId := d.Get("cloud_config_id").(string)

	if d.HasChange("fargate_profile") {
		fargateProfiles := make([]*models.V1FargateProfile, 0)
		for _, fargateProfile := range d.Get("fargate_profile").([]interface{}) {
			f := toFargateProfileEks(fargateProfile)
			fargateProfiles = append(fargateProfiles, f)
		}

		log.Printf("Updating fargate profiles")
		fargateProfilesList := &models.V1EksFargateProfiles{
			FargateProfiles: fargateProfiles,
		}

		err := c.UpdateFargateProfiles(cloudConfigId, fargateProfilesList)
		if err != nil {
			return diag.FromErr(err)
		}
	}

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
			hash := resourceMachinePoolEksHash(machinePoolResource)

			machinePool := toMachinePoolEks(machinePoolResource)

			var err error
			if oldMachinePool, ok := osMap[name]; !ok {
				log.Printf("Create machine pool %s", name)
				err = c.CreateMachinePoolEks(cloudConfigId, machinePool)
			} else if hash != resourceMachinePoolEksHash(oldMachinePool) {
				// TODO
				log.Printf("Change in machine pool %s", name)
				err = c.UpdateMachinePoolEks(cloudConfigId, machinePool)
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
			if err := c.DeleteMachinePoolEks(cloudConfigId, name); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	//if d.HasChange("fargate_profile") {
	//	oraw, nraw := d.GetChange("fargate_profile")
	//	if oraw == nil {
	//		oraw = new(schema.Set)
	//	}
	//	if nraw == nil {
	//		nraw = new(schema.Set)
	//	}
	//
	//	os := oraw.([]interface{})
	//	ns := nraw.([]interface{})
	//
	//	osMap := make(map[string]interface{})
	//	for _, mp := range os {
	//		fargateProfile := mp.(map[string]interface{})
	//		osMap[fargateProfile["name"].(string)] = fargateProfile
	//	}
	//
	//	for _, mp := range ns {
	//		fargateProfileResource := mp.(map[string]interface{})
	//		name := fargateProfileResource["name"].(string)
	//		hash := resourceFargateProfileEksHash(fargateProfileResource)
	//
	//		fargateProfile := toFargateProfileEks(fargateProfileResource)
	//
	//		var err error
	//		if oldMachinePool, ok := osMap[name]; !ok {
	//			log.Printf("Create fargate profile %s", name)
	//			err = c.CreateFargateProfileEks(cloudConfigId, fargateProfile)
	//		} else if hash != resourceFargateProfileEksHash(oldMachinePool) {
	//			// TODO
	//			log.Printf("Change in fargate profile %s", name)
	//			err = c.UpdateFargateProfileEks(cloudConfigId, fargateProfile)
	//		}
	//
	//		if err != nil {
	//			return diag.FromErr(err)
	//		}
	//
	//		// Processed (if exists)
	//		delete(osMap, name)
	//	}
	//
	//	// Deleted old fargate profiles
	//	for _, mp := range osMap {
	//		fargateProfile := mp.(map[string]interface{})
	//		name := fargateProfile["name"].(string)
	//		log.Printf("Deleted fargate profile %s", name)
	//		if err := c.DeleteFargateProfileEks(cloudConfigId, name); err != nil {
	//			return diag.FromErr(err)
	//		}
	//	}
	//}
	//

	//TODO(saamalik) update for cluster as well
	//if err := waitForClusterU(ctx, c, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
	//	return diag.FromErr(err)
	//}

	if d.HasChange("namespaces") {
		if err := updateClusterNamespaces(c, d); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("cluster_rbac_binding") {
		if err := updateClusterRBAC(c, d); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChanges("cluster_profile") {
		if err := updateProfiles(c, d); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("backup_policy") {
		if err := updateBackupPolicy(c, d); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("scan_policy") {
		if err := updateScanPolicy(c, d); err != nil {
			return diag.FromErr(err)
		}
	}

	resourceClusterEksRead(ctx, d, m)

	return diags
}

func toEksCluster(d *schema.ResourceData) *models.V1SpectroEksClusterEntity {
	// gnarly, I know! =/
	cloudConfig := d.Get("cloud_config").([]interface{})[0].(map[string]interface{})
	//clientSecret := strfmt.Password(d.Get("Eks_client_secret").(string))

	cluster := &models.V1SpectroEksClusterEntity{
		Metadata: &models.V1ObjectMeta{
			Name:   d.Get("name").(string),
			UID:    d.Id(),
			Labels: toTags(d),
		},
		Spec: &models.V1SpectroEksClusterEntitySpec{
			CloudAccountUID: ptr.StringPtr(d.Get("cloud_account_id").(string)),
			Profiles:        toProfiles(d),
			Policies:        toPolicies(d),
			CloudConfig: &models.V1EksClusterConfig{
				BastionDisabled: true,
				VpcID:           cloudConfig["vpc_id"].(string),
				Region:          ptr.StringPtr(cloudConfig["region"].(string)),
				SSHKeyName:      cloudConfig["ssh_key_name"].(string),
			},
		},
	}

	access := &models.V1EksClusterConfigEndpointAccess{}
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

	machinePoolConfigs := make([]*models.V1EksMachinePoolConfigEntity, 0)
	cpPool := map[string]interface{}{
		"control_plane": true,
		"name":          "master-pool",
		"az_subnets":    cloudConfig["az_subnets"],
		"instance_type": "t3.large",
		"disk_size_gb":  60,
		"count":         2,
	}
	machinePoolConfigs = append(machinePoolConfigs, toMachinePoolEks(cpPool))
	for _, machinePool := range d.Get("machine_pool").([]interface{}) {
		mp := toMachinePoolEks(machinePool)
		machinePoolConfigs = append(machinePoolConfigs, mp)
	}

	cluster.Spec.Machinepoolconfig = machinePoolConfigs
	cluster.Spec.ClusterConfig = &models.V1ClusterConfigEntity{
		Resources: toClusterResourceConfig(d),
	}

	fargateProfiles := make([]*models.V1FargateProfile, 0)
	for _, fargateProfile := range d.Get("fargate_profile").([]interface{}) {
		f := toFargateProfileEks(fargateProfile)
		fargateProfiles = append(fargateProfiles, f)
	}

	cluster.Spec.FargateProfiles = fargateProfiles

	return cluster
}

func toMachinePoolEks(machinePool interface{}) *models.V1EksMachinePoolConfigEntity {
	m := machinePool.(map[string]interface{})

	labels := make([]string, 0)
	controlPlane, _ := m["control_plane"].(bool)
	if controlPlane {
		labels = append(labels, "master")
	}

	azs := make([]string, 0)
	subnets := make([]*models.V1EksSubnetEntity, 0)
	for k, val := range m["az_subnets"].(map[string]interface{}) {
		azs = append(azs, k)
		if val.(string) != "" && val.(string) != "-" {
			subnets = append(subnets, &models.V1EksSubnetEntity{
				Az: k,
				ID: val.(string),
			})
		}
	}

	capacityType := "on-demand" // on-demand by default.
	if m["capacity_type"] != nil && len(m["capacity_type"].(string)) > 0 {
		capacityType = m["capacity_type"].(string)
	}

	additionalLabels := make(map[string]string)
	if m["additional_labels"] != nil {
		additionalLabels = expandStringMap(m["additional_labels"].(map[string]interface{}))
	}

	min := int32(m["count"].(int))
	max := int32(m["count"].(int))

	if m["min"] != nil {
		min = int32(m["min"].(int))
	}

	if m["max"] != nil {
		max = int32(m["max"].(int))
	}

	mp := &models.V1EksMachinePoolConfigEntity{
		CloudConfig: &models.V1EksMachineCloudConfigEntity{
			RootDeviceSize: int64(m["disk_size_gb"].(int)),
			InstanceType:   m["instance_type"].(string),
			CapacityType:   &capacityType,
			Azs:            azs,
			Subnets:        subnets,
		},
		PoolConfig: &models.V1MachinePoolConfigEntity{
			AdditionalLabels: additionalLabels,
			Taints:           toClusterTaints(m),
			IsControlPlane:   controlPlane,
			Labels:           labels,
			Name:             ptr.StringPtr(m["name"].(string)),
			Size:             ptr.Int32Ptr(int32(m["count"].(int))),
			MinSize:          min,
			MaxSize:          max,
		},
	}

	if capacityType == "spot" {
		maxPrice := "0.0" // default value
		if m["max_price"] != nil && len(m["max_price"].(string)) > 0 {
			maxPrice = m["max_price"].(string)
		}

		mp.CloudConfig.SpotMarketOptions = &models.V1SpotMarketOptions{
			MaxPrice: maxPrice,
		}
	}

	return mp
}

func toFargateProfileEks(fargateProfile interface{}) *models.V1FargateProfile {
	m := fargateProfile.(map[string]interface{})

	labels := make([]string, 0)
	controlPlane, _ := m["control_plane"].(bool)
	if controlPlane {
		labels = append(labels, "master")
	}

	selectors := make([]*models.V1FargateSelector, 0)
	for _, val := range m["selector"].([]interface{}) {
		s := val.(map[string]interface{})

		selectors = append(selectors, &models.V1FargateSelector{
			Labels:    expandStringMap(s["labels"].(map[string]interface{})),
			Namespace: ptr.StringPtr(s["namespace"].(string)),
		})
	}

	f := &models.V1FargateProfile{
		Name:           ptr.StringPtr(m["name"].(string)),
		AdditionalTags: expandStringMap(m["additional_tags"].(map[string]interface{})),
		Selectors:      selectors,
		SubnetIds:      expandStringList(m["subnets"].([]interface{})),
	}

	return f
}

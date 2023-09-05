package spectrocloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/schemas"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/hapi/models"

	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"
)

func resourceClusterLibvirt() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterVirtCreate,
		ReadContext:   resourceClusterLibvirtRead,
		UpdateContext: resourceClusterVirtUpdate,
		DeleteContext: resourceClusterDelete,
		Description:   "Resource for managing Libvirt clusters in Spectro Cloud through Palette.",

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
			"cluster_meta_attribute": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "`cluster_meta_attribute` can be used to set additional cluster metadata information, eg `{'nic_name': 'test', 'env': 'stage'}`",
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
						"type": {
							Type:     schema.TypeString,
							Optional: true,
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
										Required: true,
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
			"cloud_account_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
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
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ssh_key": {
							Type:         schema.TypeString,
							Optional:     true,
							ExactlyOneOf: []string{"cloud_config.0.ssh_key", "cloud_config.0.ssh_keys"},
						},
						"ssh_keys": {
							Type:         schema.TypeSet,
							Optional:     true,
							Set:          schema.HashString,
							ExactlyOneOf: []string{"cloud_config.0.ssh_key", "cloud_config.0.ssh_keys"},
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"vip": {
							Type:     schema.TypeString,
							Required: true,
						},
						"ntp_servers": {
							Type:     schema.TypeSet,
							Optional: true,
							Set:      schema.HashString,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						// DHCP or VIP Properties
						"network_search_domain": {
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
							Description: "The search domain to use for the cluster in case of DHCP.",
						},
						"network_type": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "VIP",
							ForceNew:    true,
							Description: "The type of network to use for the cluster. This can be `VIP` or `DDNS`.",
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
				// disable hash to preserve machine pool order PE-255
				//Set:      resourceMachinePoolLibvirtHash,
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
						"count": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"update_strategy": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "RollingUpdateScaleOut",
						},
						"instance_type": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"attached_disks": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"managed": {
													Type:     schema.TypeBool,
													Optional: true,
													Default:  false,
												},
												"size_in_gb": {
													Type:     schema.TypeInt,
													Required: true,
												},
											},
										},
									},
									"cpus_sets": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"cache_passthrough": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"gpu_config": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"device_model": {
													Type:     schema.TypeString,
													Required: true,
												},
												"vendor": {
													Type:     schema.TypeString,
													Required: true,
												},
												"num_gpus": {
													Type:     schema.TypeInt,
													Required: true,
												},
												"addresses": {
													Type:     schema.TypeMap,
													Optional: true,
													Elem: &schema.Schema{
														Type: schema.TypeString,
													},
												},
											},
										},
									},
									"disk_size_gb": {
										Type:     schema.TypeInt,
										Required: true,
									},
									"memory_mb": {
										Type:     schema.TypeInt,
										Required: true,
									},
									"cpu": {
										Type:     schema.TypeInt,
										Required: true,
									},
								},
							},
						},
						"placements": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"appliance_id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"network_type": {
										Type:         schema.TypeString,
										ValidateFunc: validation.StringInSlice([]string{"default", "bridge"}, false),
										Required:     true,
									},
									"network_names": {
										Type:     schema.TypeString,
										Required: true,
									},
									"image_storage_pool": {
										Type:     schema.TypeString,
										Required: true,
									},
									"target_storage_pool": {
										Type:     schema.TypeString,
										Required: true,
									},
									"data_storage_pool": {
										Type:     schema.TypeString,
										Required: true,
									},
									"network": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"gpu_device": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"device_model": {
													Type:     schema.TypeString,
													Required: true,
												},
												"vendor": {
													Type:     schema.TypeString,
													Required: true,
												},
												"addresses": {
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
						"xsl_template": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "XSL template to use.",
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
			"host_config":     schemas.ClusterHostConfigSchema(),
			"location_config": schemas.ClusterLocationSchema(),
			"skip_completion": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func resourceClusterVirtCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	cluster, err := toLibvirtCluster(c, d)
	if err != nil {
		return diag.FromErr(err)
	}

	uid, err := c.CreateClusterLibvirt(cluster)
	if err != nil {
		return diag.FromErr(err)
	}

	diagnostics, isError := waitForClusterCreation(ctx, d, uid, diags, c, true)
	if isError {
		return diagnostics
	}

	diags = resourceClusterLibvirtRead(ctx, d, m)

	return diags
}

//goland:noinspection GoUnhandledErrorResult
func resourceClusterLibvirtRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	// Update the kubeconfig
	diagnostics, errorSet := readCommonFields(c, d, cluster)
	if errorSet {
		return diagnostics
	}

	diags = flattenCloudConfigLibvirt(cluster.Spec.CloudConfigRef.UID, d, c)
	return diags
}

func flattenCloudConfigLibvirt(configUID string, d *schema.ResourceData, c *client.V1Client) diag.Diagnostics {
	if err := d.Set("cloud_config_id", configUID); err != nil {
		return diag.FromErr(err)
	}
	if config, err := c.GetCloudConfigLibvirt(configUID); err != nil {
		return diag.FromErr(err)
	} else {
		mp := flattenMachinePoolConfigsLibvirt(config.Spec.MachinePoolConfig)
		if err := d.Set("machine_pool", mp); err != nil {
			return diag.FromErr(err)
		}
	}

	return diag.Diagnostics{}
}

func flattenMachinePoolConfigsLibvirt(machinePools []*models.V1LibvirtMachinePoolConfig) []interface{} {

	if machinePools == nil {
		return make([]interface{}, 0)
	}

	ois := make([]interface{}, 0, 1)

	for _, machinePool := range machinePools {
		oi := make(map[string]interface{})

		SetAdditionalLabelsAndTaints(machinePool.AdditionalLabels, machinePool.Taints, oi)

		oi["control_plane"] = machinePool.IsControlPlane
		oi["control_plane_as_worker"] = machinePool.UseControlPlaneAsWorker
		oi["name"] = machinePool.Name
		oi["count"] = machinePool.Size
		flattenUpdateStrategy(machinePool.UpdateStrategy, oi)

		if machinePool.InstanceType != nil {
			s := make(map[string]interface{})
			additionalDisks := make([]interface{}, 0)

			if machinePool.NonRootDisksInGB != nil && len(machinePool.NonRootDisksInGB) > 0 {
				for _, disk := range machinePool.NonRootDisksInGB {
					addDisk := make(map[string]interface{})
					addDisk["managed"] = disk.Managed
					addDisk["size_in_gb"] = *disk.SizeInGB
					additionalDisks = append(additionalDisks, addDisk)
				}
			}
			s["disk_size_gb"] = int(*machinePool.RootDiskInGB)
			if len(machinePool.InstanceType.Cpuset) > 0 {
				s["cpus_sets"] = machinePool.InstanceType.Cpuset
			}

			if machinePool.InstanceType.CPUPassthroughSpec != nil && (*machinePool.InstanceType.CPUPassthroughSpec).IsEnabled {
				s["cache_passthrough"] = (*machinePool.InstanceType.CPUPassthroughSpec).CachePassthrough
			}

			config := make([]interface{}, 0)
			if machinePool.InstanceType.GpuConfig != nil {
				gpuConfig := *machinePool.InstanceType.GpuConfig

				if !(gpuConfig.DeviceModel == "" || gpuConfig.VendorName == "") && gpuConfig.NumGPUs != 0 {
					aconfig := make(map[string]interface{})

					aconfig["device_model"] = gpuConfig.DeviceModel
					aconfig["vendor"] = gpuConfig.VendorName
					aconfig["num_gpus"] = gpuConfig.NumGPUs
					aconfig["addresses"] = gpuConfig.Addresses
				}
			}

			s["gpu_config"] = config
			s["memory_mb"] = int(*machinePool.InstanceType.MemoryInMB)
			s["cpu"] = int(*machinePool.InstanceType.NumCPUs)

			oi["instance_type"] = []interface{}{s}
			s["attached_disks"] = additionalDisks
		}

		placements := make([]interface{}, len(machinePool.Placements))
		for j, p := range machinePool.Placements {
			pj := make(map[string]interface{})
			pj["appliance_id"] = p.HostUID
			if p.Networks != nil {
				for _, network := range p.Networks {
					pj["network_type"] = network.NetworkType
					break
				}
			}
			networkNames := make([]string, 0)
			for _, network := range p.Networks {
				networkNames = append(networkNames, *network.NetworkName)
			}
			networkNamesStr := strings.Join(networkNames, ",")

			pj["network_names"] = networkNamesStr
			pj["image_storage_pool"] = p.SourceStoragePool
			pj["target_storage_pool"] = p.TargetStoragePool
			pj["data_storage_pool"] = p.DataStoragePool
			pj["gpu_device"] = flattenGpuDevice(p.GpuDevices)
			placements[j] = pj
		}
		oi["placements"] = placements
		oi["xsl_template"] = machinePool.XslTemplate

		ois = append(ois, oi)
	}

	return ois
}

func flattenGpuDevice(gpus []*models.V1GPUDeviceSpec) []interface{} {
	if gpus != nil {
		dConfig := make([]interface{}, 0)
		for _, d := range gpus {
			if !(d.Model == "" || d.Vendor == "") {
				dElem := make(map[string]interface{})
				dElem["device_model"] = d.Model
				dElem["vendor"] = d.Vendor
				dElem["addresses"] = d.Addresses
				dConfig = append(dConfig, dElem)
			}
		}
		return dConfig
	}
	return make([]interface{}, 0)
}

func resourceClusterVirtUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
			if name == "" {
				continue
			}
			hash := resourceMachinePoolLibvirtHash(machinePoolResource)

			machinePool, err := toMachinePoolLibvirt(machinePoolResource)
			if err != nil {
				return diag.FromErr(err)
			}

			if oldMachinePool, ok := osMap[name]; !ok {
				log.Printf("Create machine pool %s", name)
				err = c.CreateMachinePoolLibvirt(cloudConfigId, machinePool)
			} else if hash != resourceMachinePoolLibvirtHash(oldMachinePool) {
				log.Printf("Change in machine pool %s", name)
				err = c.UpdateMachinePoolLibvirt(cloudConfigId, machinePool)
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
			if err := c.DeleteMachinePoolLibvirt(cloudConfigId, name); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	diagnostics, errorSet := updateCommonFields(d, c)
	if errorSet {
		return diagnostics
	}

	diags = resourceClusterLibvirtRead(ctx, d, m)

	return diags
}

func toLibvirtCluster(c *client.V1Client, d *schema.ResourceData) (*models.V1SpectroLibvirtClusterEntity, error) {
	cloudConfig := d.Get("cloud_config").([]interface{})[0].(map[string]interface{})
	sshKeys, err := toSSHKeys(cloudConfig)
	if err != nil {
		return nil, err
	}
	cluster := &models.V1SpectroLibvirtClusterEntity{
		Metadata: &models.V1ObjectMeta{
			Name:   d.Get("name").(string),
			UID:    d.Id(),
			Labels: toTags(d),
		},
		Spec: &models.V1SpectroLibvirtClusterEntitySpec{
			Profiles: toProfiles(c, d),
			Policies: toPolicies(d),
			CloudConfig: &models.V1LibvirtClusterConfig{
				NtpServers: toNtpServers(cloudConfig),
				SSHKeys:    sshKeys,
				ControlPlaneEndpoint: &models.V1LibvirtControlPlaneEndPoint{
					Host:             cloudConfig["vip"].(string),
					Type:             cloudConfig["network_type"].(string),
					DdnsSearchDomain: cloudConfig["network_search_domain"].(string),
				},
			},
		},
	}

	machinePoolConfigs := make([]*models.V1LibvirtMachinePoolConfigEntity, 0)
	for _, machinePool := range d.Get("machine_pool").([]interface{}) {
		mp, err := toMachinePoolLibvirt(machinePool)
		if err != nil {
			return nil, err
		}
		machinePoolConfigs = append(machinePoolConfigs, mp)
	}

	// sort
	/*sort.SliceStable(machinePoolConfigs, func(i, j int) bool {
		return machinePoolConfigs[i].PoolConfig.IsControlPlane
	})*/

	cluster.Spec.Machinepoolconfig = machinePoolConfigs
	cluster.Spec.ClusterConfig = toClusterConfig(d)

	return cluster, nil
}

func toMachinePoolLibvirt(machinePool interface{}) (*models.V1LibvirtMachinePoolConfigEntity, error) {
	m := machinePool.(map[string]interface{})

	labels := make([]string, 0)
	controlPlane := m["control_plane"].(bool)
	controlPlaneAsWorker := m["control_plane_as_worker"].(bool)
	if controlPlane {
		labels = append(labels, "master")
	}

	placements := make([]*models.V1LibvirtPlacementEntity, 0)
	for _, pos := range m["placements"].([]interface{}) {
		p := pos.(map[string]interface{})
		networks := getNetworks(p)

		imageStoragePool := p["image_storage_pool"].(string)
		targetStoragePool := p["target_storage_pool"].(string)
		dataStoragePool := p["data_storage_pool"].(string)
		gpuDevices := p["gpu_device"]
		placements = append(placements, &models.V1LibvirtPlacementEntity{
			Networks:          networks,
			SourceStoragePool: imageStoragePool,
			TargetStoragePool: targetStoragePool,
			DataStoragePool:   dataStoragePool,
			HostUID:           types.Ptr(p["appliance_id"].(string)),
			GpuDevices:        getGPUDevices(gpuDevices),
		})

	}

	ins := m["instance_type"].([]interface{})[0].(map[string]interface{})

	var cpuPassthroughSpec *models.V1CPUPassthroughSpec
	if ins["cache_passthrough"] != nil {
		cpuPassthroughSpec = &models.V1CPUPassthroughSpec{
			CachePassthrough: ins["cache_passthrough"].(bool),
			IsEnabled:        true,
		}
	}

	instanceType := models.V1LibvirtInstanceType{
		MemoryInMB:         types.Ptr(int32(ins["memory_mb"].(int))),
		NumCPUs:            types.Ptr(int32(ins["cpu"].(int))),
		GpuConfig:          getGPUConfig(ins),
		CPUPassthroughSpec: cpuPassthroughSpec,
	}

	if ins["cpus_sets"] != nil && len(ins["cpus_sets"].(string)) > 0 {
		instanceType.Cpuset = ins["cpus_sets"].(string)
	}
	addDisks := getAdditionalDisks(ins)

	if m["name"].(string) == "master-pool" && len(addDisks) > 0 {
		// If master pool has additional disks, return an error
		return nil, fmt.Errorf("Attached disks are not allowed for the 'master-pool' machine pool")
	}

	updateStrategyType := getUpdateStrategy(m)
	if m["name"].(string) == "master-pool" && updateStrategyType == "RollingUpdateScaleIn" {
		// If master pool has RollingUpdateScaleIn as an update strategy, return an error
		return nil, fmt.Errorf("Update strategy RollingUpdateScaleIn is not allowed for the 'master-pool' machine pool")
	}

	var xlstemplate string
	if m["xsl_template"] != nil {
		xlstemplate = m["xsl_template"].(string)
	}

	mp := &models.V1LibvirtMachinePoolConfigEntity{
		CloudConfig: &models.V1LibvirtMachinePoolCloudConfigEntity{
			Placements:       placements,
			RootDiskInGB:     types.Ptr(int32(ins["disk_size_gb"].(int))),
			NonRootDisksInGB: addDisks,
			InstanceType:     &instanceType,
			XslTemplate:      xlstemplate,
		},
		PoolConfig: &models.V1MachinePoolConfigEntity{
			AdditionalLabels: toAdditionalNodePoolLabels(m),
			Taints:           toClusterTaints(m),
			IsControlPlane:   controlPlane,
			Labels:           labels,
			Name:             types.Ptr(m["name"].(string)),
			Size:             types.Ptr(int32(m["count"].(int))),
			UpdateStrategy: &models.V1UpdateStrategy{
				Type: updateStrategyType,
			},
			UseControlPlaneAsWorker: controlPlaneAsWorker,
		},
	}
	return mp, nil
}

func getGPUConfig(ins map[string]interface{}) *models.V1GPUConfig {
	if ins["gpu_config"] != nil {
		for _, t := range ins["gpu_config"].([]interface{}) {
			config := t.(map[string]interface{})
			mapAddresses := make(map[string]string)
			// "TU104GL [Quadro RTX 4000]": "11:00.0", ...
			if config["addresses"] != nil && len(config["addresses"].(map[string]interface{})) > 0 {
				mapAddresses = expandStringMap(config["addresses"].(map[string]interface{}))
			}
			if config != nil {
				return &models.V1GPUConfig{
					DeviceModel: config["device_model"].(string),
					NumGPUs:     int32(config["num_gpus"].(int)),
					VendorName:  config["vendor"].(string),
					Addresses:   mapAddresses,
				}
			}
		}
	}
	return nil
}

func getGPUDevices(gpuDevice interface{}) []*models.V1GPUDeviceSpec {
	if gpuDevice != nil {
		gpuDevices := make([]*models.V1GPUDeviceSpec, 0)
		for _, t := range gpuDevice.([]interface{}) {
			config := t.(map[string]interface{})
			mapAddresses := make(map[string]string)
			if config["addresses"] != nil && len(config["addresses"].(map[string]interface{})) > 0 {
				mapAddresses = expandStringMap(config["addresses"].(map[string]interface{}))
			}
			if config != nil {
				gpuDevices = append(gpuDevices, &models.V1GPUDeviceSpec{
					Model:     config["device_model"].(string),
					Vendor:    config["vendor"].(string),
					Addresses: mapAddresses,
				})
			}
		}
		return gpuDevices
	}
	return nil
}

func getAdditionalDisks(ins map[string]interface{}) []*models.V1LibvirtDiskSpec {
	addDisks := make([]*models.V1LibvirtDiskSpec, 0)

	if ins["attached_disks"] != nil {
		for _, disk := range ins["attached_disks"].([]interface{}) {
			size := int32(0)
			managed := false
			for j, prop := range disk.(map[string]interface{}) {
				switch {
				case j == "managed":
					managed = prop.(bool)
					break
				case j == "size_in_gb":
					size = int32(prop.(int))
					break
				default:
					return nil
				}
			}

			addDisks = append(addDisks, &models.V1LibvirtDiskSpec{
				SizeInGB: &size,
				Managed:  managed,
			})
		}
	}
	return addDisks
}

func getNetworks(p map[string]interface{}) []*models.V1LibvirtNetworkSpec {
	networkType := ""
	networks := make([]*models.V1LibvirtNetworkSpec, 0)

	if p["network_names"] != nil {
		for _, n := range strings.Split(p["network_names"].(string), ",") {
			networkName := strings.TrimSpace(n)
			networkType = p["network_type"].(string)
			network := &models.V1LibvirtNetworkSpec{
				NetworkName: &networkName,
				NetworkType: &networkType,
			}
			networks = append(networks, network)
		}
	}
	return networks
}

package spectrocloud

import (
	"context"
	"log"
	"time"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/schemas"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/palette-sdk-go/client"
)

func resourceClusterVirtual() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterVirtualCreate,
		ReadContext:   resourceClusterVirtualRead,
		UpdateContext: resourceClusterVirtualUpdate,
		DeleteContext: resourceClusterDelete,
		Description:   "A resource to manage a Palette Virtual Cluster.",

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
				Description: "A list of tags to be applied to the cluster. Tags must be in the form of `key:value`.",
			},
			"host_cluster_uid": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"host_cluster_uid", "cluster_group_uid"},
				ValidateFunc: validation.StringNotInSlice([]string{""}, false),
			},
			"cluster_group_uid": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringNotInSlice([]string{""}, false),
			},
			"pause_cluster": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "To pause and resume cluster state. Set to true to pause running cluster & false to resume it.",
			},
			"resources": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"max_cpu": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"max_mem_in_mb": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"max_storage_in_gb": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"min_cpu": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"min_mem_in_mb": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"min_storage_in_gb": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
			"cluster_profile": schemas.ClusterProfileSchema(),
			"apply_setting": {
				Type:     schema.TypeString,
				Optional: true,
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
				Description:      "The date and time after which to patch the cluster. Prefix the time value with the respective RFC. Ex: `RFC3339: 2006-01-02T15:04:05Z07:00`",
			},
			"kubeconfig": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Kubeconfig for the cluster. This can be used to connect to the cluster using `kubectl`.",
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
			"backup_policy":        schemas.BackupPolicySchema(),
			"scan_policy":          schemas.ScanPolicySchema(),
			"cluster_rbac_binding": schemas.ClusterRbacBindingSchema(),
			"namespaces":           schemas.ClusterNamespacesSchema(),
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

func resourceClusterVirtualCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	cluster := toVirtualCluster(c, d)

	uid, err := c.CreateClusterVirtual(cluster)
	if err != nil {
		return diag.FromErr(err)
	}

	diagnostics, isError := waitForClusterCreation(ctx, d, uid, diags, c, true)
	if isError {
		return diagnostics
	}

	resourceClusterVirtualRead(ctx, d, m)

	return diags
}

//goland:noinspection GoUnhandledErrorResult
func resourceClusterVirtualRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	return flattenCloudConfigVirtual(cluster.Spec.CloudConfigRef.UID, d, c)
}

func flattenCloudConfigVirtual(configUID string, d *schema.ResourceData, c *client.V1Client) diag.Diagnostics {
	err := d.Set("cloud_config_id", configUID)
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}

func resourceClusterVirtualUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
			if machinePoolResource["name"].(string) != "" {
				name := machinePoolResource["name"].(string)
				hash := resourceMachinePoolVirtualHash(machinePoolResource)

				machinePool := toMachinePoolVirtual(machinePoolResource)

				var err error
				if oldMachinePool, ok := osMap[name]; !ok {
					log.Printf("Create machine pool %s", name)
					err = c.CreateMachinePoolVirtual(cloudConfigId, machinePool)
				} else if hash != resourceMachinePoolVirtualHash(oldMachinePool) {
					log.Printf("Change in machine pool %s", name)
					err = c.UpdateMachinePoolVirtual(cloudConfigId, machinePool)
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
			if err := c.DeleteMachinePoolVirtual(cloudConfigId, name); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	diagnostics, done := updateCommonFields(d, c)
	if done {
		return diagnostics
	}
	if d.HasChange("pause_cluster") {
		var body *models.V1LifecycleConfigEntity
		pause := new(bool)
		if d.Get("pause_cluster").(bool) {
			*pause = true
		} else {
			*pause = false
		}
		body = &models.V1LifecycleConfigEntity{
			LifecycleConfig: &models.V1LifecycleConfig{
				Pause: pause,
			},
		}
		_, err := c.VirtualClusterLifecycleConfigChange(d.Id(), body)
		if err != nil {
			return diag.FromErr(err)
		}
		if *pause {
			diagnostics, isError := waitForVirtualClusterLifecyclePause(ctx, d, d.Id(), diags, c)
			if isError {
				return diagnostics
			}
		} else {
			diagnostics, isError := waitForVirtualClusterLifecycleResume(ctx, d, d.Id(), diags, c)
			if isError {
				return diagnostics
			}
		}

	}
	resourceClusterVirtualRead(ctx, d, m)
	return diags
}

func toVirtualCluster(c *client.V1Client, d *schema.ResourceData) *models.V1SpectroVirtualClusterEntity {
	// parse host cluster / cluster group uid
	hostClusterUid := d.Get("host_cluster_uid").(string)
	clusterGroupUid := d.Get("cluster_group_uid").(string)

	// parse CloudConfig
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

	// init cluster
	cluster := &models.V1SpectroVirtualClusterEntity{
		Metadata: &models.V1ObjectMeta{
			Name:   d.Get("name").(string),
			UID:    d.Id(),
			Labels: toTags(d),
		},
		Spec: &models.V1SpectroVirtualClusterEntitySpec{
			CloudConfig: &models.V1VirtualClusterConfig{
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
			ClusterConfig: &models.V1ClusterConfigEntity{
				HostClusterConfig: &models.V1HostClusterConfig{
					ClusterGroup: &models.V1ObjectReference{
						UID: clusterGroupUid,
					},
					HostCluster: &models.V1ObjectReference{
						UID: hostClusterUid,
					},
				},
			},
			Machinepoolconfig: nil,
			Profiles:          toProfiles(c, d),
			Policies:          toPolicies(d),
		},
	}

	// init cluster resources (machinepool)
	machinePoolConfigs := make([]*models.V1VirtualMachinePoolConfigEntity, 0)
	resourcesObj, ok := d.GetOk("resources")
	if ok {
		resources := resourcesObj.([]interface{})[0].(map[string]interface{})
		mp := toMachinePoolVirtual(resources)
		machinePoolConfigs = append(machinePoolConfigs, mp)
	}
	cluster.Spec.Machinepoolconfig = machinePoolConfigs

	return cluster
}

func toMachinePoolVirtual(resources map[string]interface{}) *models.V1VirtualMachinePoolConfigEntity {
	maxCpu := resources["max_cpu"].(int)
	maxMemInMb := resources["max_mem_in_mb"].(int)
	maxStorageInGb := resources["max_storage_in_gb"].(int)
	minCpu := resources["min_cpu"].(int)
	minMemInMb := resources["min_mem_in_mb"].(int)
	minStorageInGb := resources["min_storage_in_gb"].(int)

	mp := &models.V1VirtualMachinePoolConfigEntity{
		CloudConfig: &models.V1VirtualMachinePoolCloudConfigEntity{
			InstanceType: &models.V1VirtualInstanceType{
				MaxCPU:        int32(maxCpu),
				MaxMemInMiB:   int32(maxMemInMb),
				MaxStorageGiB: int32(maxStorageInGb),
				MinCPU:        int32(minCpu),
				MinMemInMiB:   int32(minMemInMb),
				MinStorageGiB: int32(minStorageInGb),
			},
		},
	}

	return mp
}

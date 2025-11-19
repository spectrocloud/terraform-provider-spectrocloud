package spectrocloud

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/schemas"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
)

func resourceClusterEks() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterEksCreate,
		ReadContext:   resourceClusterEksRead,
		UpdateContext: resourceClusterEksUpdate,
		DeleteContext: resourceClusterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceClusterEksImport,
		},
		Description: "Resource for managing EKS clusters in Spectro Cloud through Palette.",

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

		SchemaVersion: 3,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceClusterEksResourceV2().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceClusterEksStateUpgradeV2,
				Version: 2,
			},
		},
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
				Description: "The context of the EKS cluster. Allowed values are `project` or `tenant`. " +
					"Default is `project`. " + PROJECT_NAME_NUANCE,
			},
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      schema.HashString,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "A list of tags to be applied to the cluster. Tags must be in the form of `key:value`. The `tags` attribute will soon be deprecated. It is recommended to use `tags_map` instead.",
			},
			"tags_map": {
				Type:          schema.TypeMap,
				Optional:      true,
				ConflictsWith: []string{"tags"},
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "A map of tags to be applied to the cluster. tags and tags_map are mutually exclusive — only one should be used at a time",
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
				Description: "The AWS cloud account id to use for this cluster.",
			},
			"cloud_config_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the cloud config used for the cluster. This cloud config must be of type `azure`.",
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
				Description:      "Date and time after which to patch cluster `RFC3339: 2006-01-02T15:04:05Z07:00`",
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
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Description: "The AWS environment configuration settings such as network parameters and encryption parameters that apply to this cluster.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ssh_key_name": {
							Type:        schema.TypeString,
							ForceNew:    true,
							Optional:    true,
							Description: "Public SSH key to be used for the cluster nodes.",
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
							Type:        schema.TypeList,
							Description: "Mutually exclusive with `az_subnets`. Use for Dynamic provisioning.",
							Optional:    true,
							ForceNew:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"az_subnets": {
							Type:        schema.TypeMap,
							Description: "Mutually exclusive with `azs`. Use for Static provisioning.",
							Optional:    true,
							ForceNew:    true,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								// UI strips the trailing newline on save
								return strings.TrimSpace(old) == strings.TrimSpace(new)
							},
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"endpoint_access": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice([]string{"public", "private", "private_and_public"}, false),
							Description:  "Choose between `private`, `public`, or `private_and_public` to define how communication is established with the endpoint for the managed Kubernetes API server and your cluster. The default value is `public`.",
							Default:      "public",
						},
						"public_access_cidrs": {
							Type:        schema.TypeSet,
							Optional:    true,
							Set:         schema.HashString,
							Description: "List of CIDR blocks that define the allowed public access to the resource. Requests originating from addresses within these CIDR blocks will be permitted to access the resource. All other addresses will be denied access.",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"private_access_cidrs": {
							Type:        schema.TypeSet,
							Optional:    true,
							Set:         schema.HashString,
							Description: "List of CIDR blocks that define the allowed private access to the resource. Only requests originating from addresses within these CIDR blocks will be permitted to access the resource.",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"encryption_config_arn": {
							Type:        schema.TypeString,
							Description: "The ARN of the KMS encryption key to use for the cluster. Refer to the [Enable Secrets Encryption for EKS Cluster](https://docs.spectrocloud.com/clusters/public-cloud/aws/enable-secrets-encryption-kms-key/) for additional guidance.",
							ForceNew:    true,
							Optional:    true,
						},
					},
				},
			},
			"machine_pool": {
				Type:        schema.TypeSet,
				Required:    true,
				Set:         resourceMachinePoolEksHash,
				Description: "The machine pool configuration for the cluster.",
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
						"disk_size_gb": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"count": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Number of nodes in the machine pool.",
						},
						"update_strategy": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "RollingUpdateScaleOut",
							Description: "Update strategy for the machine pool. Valid values are `RollingUpdateScaleOut` and `RollingUpdateScaleIn`.",
						},
						"min": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Minimum number of nodes in the machine pool. This is used for autoscaling the machine pool.",
						},
						"max": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Maximum number of nodes in the machine pool. This is used for autoscaling the machine pool.",
						},
						"instance_type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"ami_type": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "AL2_x86_64",
							Description: "Specifies the type of Amazon Machine Image (AMI) to use for the machine pool. Valid values are [`AL2_x86_64`, `AL2_x86_64_GPU`, `AL2023_x86_64_STANDARD`, `AL2023_x86_64_NEURON` and `AL2023_x86_64_NVIDIA`]. Defaults to `AL2_x86_64`.",
						},
						"capacity_type": {
							Type:         schema.TypeString,
							Default:      "on-demand",
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"on-demand", "spot"}, false),
							Description:  "Capacity type is an instance type,  can be 'on-demand' or 'spot'. Defaults to 'on-demand'.",
						},
						"max_price": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},
						"azs": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Mutually exclusive with `az_subnets`.",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"az_subnets": {
							Type:        schema.TypeMap,
							Optional:    true,
							Description: "Mutually exclusive with `azs`. Use for Static provisioning.",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"eks_launch_template": schemas.AwsLaunchTemplate(),
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

func resourceClusterEksCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	cluster, err := toEksCluster(c, d)
	if err != nil {
		return diag.FromErr(err)
	}

	uid, err := c.CreateClusterEks(cluster)
	if err != nil {
		return diag.FromErr(err)
	}

	diagnostics, isError := waitForClusterCreation(ctx, d, uid, diags, c, true)
	if len(diagnostics) > 0 {
		diags = append(diags, diagnostics...)
	}
	if isError {
		return diagnostics
	}

	resourceClusterEksRead(ctx, d, m)

	return diags
}

func resourceClusterEksRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	var diags diag.Diagnostics

	cluster, err := resourceClusterRead(d, c, diags)
	if err != nil {
		return handleReadError(d, err, diags)
	} else if cluster == nil {
		// Deleted - Terraform will recreate it
		d.SetId("")
		return diags
	}

	configUID := cluster.Spec.CloudConfigRef.UID
	if err := d.Set("cloud_config_id", configUID); err != nil {
		return diag.FromErr(err)
	}

	var config *models.V1EksCloudConfig
	if config, err = c.GetCloudConfigEks(configUID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("cloud_account_id", config.Spec.CloudAccountRef.UID); err != nil {
		return diag.FromErr(err)
	}
	cloudConfigFlatten := flattenClusterConfigsEKS(config)
	if err := d.Set("cloud_config", cloudConfigFlatten); err != nil {
		return diag.FromErr(err)
	}

	mp := flattenMachinePoolConfigsEks(config.Spec.MachinePoolConfig)

	mp, err = flattenNodeMaintenanceStatus(c, d, c.GetNodeStatusMapEks, mp, configUID)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("machine_pool", mp); err != nil {
		return diag.FromErr(err)
	}

	fp := flattenFargateProfilesEks(config.Spec.FargateProfiles)
	if err := d.Set("fargate_profile", fp); err != nil {
		return diag.FromErr(err)
	}

	// verify cluster type
	err = ValidateCloudType("spectrocloud_cluster_eks", cluster)
	if err != nil {
		return diag.FromErr(err)
	}

	diagnostics, done := readCommonFields(c, d, cluster)

	// handling flatten tags_map for aws  cluster
	if _, ok := d.GetOk("tags_map"); ok {
		// setting to empty since tags_map is present
		_ = d.Set("tags", []string{})
		tagMaps := flattenTagsMap(cluster.Metadata.Labels)
		if err := d.Set("tags_map", tagMaps); err != nil {
			return diag.FromErr(err)
		}
	}

	if done {
		return diagnostics
	}

	// Flatten cluster_template variables using variables API
	if err := flattenClusterTemplateVariables(c, d, d.Id()); err != nil {
		return diag.FromErr(err)
	}

	generalWarningForRepave(&diags)
	return diags
}

func flattenClusterConfigsEKS(cloudConfig *models.V1EksCloudConfig) interface{} {

	cloudConfigFlatten := make([]interface{}, 0)
	if cloudConfig == nil {
		return cloudConfigFlatten
	}

	ret := make(map[string]interface{})

	ret["region"] = cloudConfig.Spec.ClusterConfig.Region

	ret["public_access_cidrs"] = make([]string, 0)
	if cloudConfig.Spec.ClusterConfig.EndpointAccess.PublicCIDRs != nil {
		ret["public_access_cidrs"] = cloudConfig.Spec.ClusterConfig.EndpointAccess.PublicCIDRs
	}

	ret["private_access_cidrs"] = make([]string, 0)
	if cloudConfig.Spec.ClusterConfig.EndpointAccess.PrivateCIDRs != nil {
		ret["private_access_cidrs"] = cloudConfig.Spec.ClusterConfig.EndpointAccess.PrivateCIDRs
	}

	for _, pool := range cloudConfig.Spec.MachinePoolConfig {
		if pool.Name == "cp-pool" {
			ret["az_subnets"] = pool.SubnetIds
		}

	}

	if cloudConfig.Spec.ClusterConfig.EncryptionConfig != nil && cloudConfig.Spec.ClusterConfig.EncryptionConfig.IsEnabled {
		ret["encryption_config_arn"] = cloudConfig.Spec.ClusterConfig.EncryptionConfig.Provider
	}

	if cloudConfig.Spec.ClusterConfig.EndpointAccess.Private && cloudConfig.Spec.ClusterConfig.EndpointAccess.Public {
		ret["endpoint_access"] = "private_and_public"
	}
	if cloudConfig.Spec.ClusterConfig.EndpointAccess.Private && !cloudConfig.Spec.ClusterConfig.EndpointAccess.Public {
		ret["endpoint_access"] = "private"
	}
	if !cloudConfig.Spec.ClusterConfig.EndpointAccess.Private && cloudConfig.Spec.ClusterConfig.EndpointAccess.Public {
		ret["endpoint_access"] = "public"
	}
	ret["region"] = *cloudConfig.Spec.ClusterConfig.Region
	ret["vpc_id"] = cloudConfig.Spec.ClusterConfig.VpcID
	ret["ssh_key_name"] = cloudConfig.Spec.ClusterConfig.SSHKeyName

	cloudConfigFlatten = append(cloudConfigFlatten, ret)

	return cloudConfigFlatten
}

func flattenMachinePoolConfigsEks(machinePools []*models.V1EksMachinePoolConfig) []interface{} {

	if machinePools == nil {
		return make([]interface{}, 0)
	}

	ois := make([]interface{}, 0)

	for _, machinePool := range machinePools {
		oi := make(map[string]interface{})

		FlattenAdditionalLabelsAndTaints(machinePool.AdditionalLabels, machinePool.Taints, oi)

		if machinePool.IsControlPlane != nil && *machinePool.IsControlPlane {
			continue
		}

		oi["name"] = machinePool.Name
		oi["count"] = int(machinePool.Size)
		flattenUpdateStrategy(machinePool.UpdateStrategy, oi)

		oi["min"] = int(machinePool.MinSize)
		oi["max"] = int(machinePool.MaxSize)
		oi["instance_type"] = machinePool.InstanceType
		oi["ami_type"] = machinePool.AmiType
		if machinePool.CapacityType != nil {
			oi["capacity_type"] = machinePool.CapacityType
		}
		if machinePool.SpotMarketOptions != nil {
			if machinePool.SpotMarketOptions.MaxPrice != "" {
				oi["max_price"] = machinePool.SpotMarketOptions.MaxPrice
			} else {
				oi["max_price"] = ""
			}
		}
		oi["disk_size_gb"] = int(machinePool.RootDeviceSize)
		if len(machinePool.SubnetIds) > 0 {
			oi["az_subnets"] = machinePool.SubnetIds
		} else {
			oi["azs"] = machinePool.Azs
		}
		eksLaunchTemplates := flattenEksLaunchTemplate(machinePool.AwsLaunchTemplate)

		if eksLaunchTemplates != nil {
			oi["eks_launch_template"] = flattenEksLaunchTemplate(machinePool.AwsLaunchTemplate)
		}

		ois = append(ois, oi)
	}

	return ois
}

func flattenEksLaunchTemplate(launchTemplate *models.V1AwsLaunchTemplate) []interface{} {
	if launchTemplate == nil {
		return make([]interface{}, 0)
	}

	lt := make(map[string]interface{})

	if launchTemplate.Ami != nil {
		lt["ami_id"] = launchTemplate.Ami.ID
	}
	if launchTemplate.RootVolume != nil {
		lt["root_volume_type"] = launchTemplate.RootVolume.Type
		lt["root_volume_iops"] = launchTemplate.RootVolume.Iops
		lt["root_volume_throughput"] = launchTemplate.RootVolume.Throughput
	}
	if len(launchTemplate.AdditionalSecurityGroups) > 0 {
		var additionalSecurityGroups []string
		for _, sg := range launchTemplate.AdditionalSecurityGroups {
			additionalSecurityGroups = append(additionalSecurityGroups, sg.ID)
		}
		lt["additional_security_groups"] = additionalSecurityGroups
	}
	// handling eks template flatten with this code eks template will not set back to schema
	if lt["ami_id"].(string) != "" ||
		lt["root_volume_type"].(string) != "" ||
		lt["root_volume_iops"].(int64) != 0 ||
		lt["root_volume_throughput"].(int64) != 0 ||
		len(launchTemplate.AdditionalSecurityGroups) > 0 {
		return []interface{}{lt}
	}
	return nil
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
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	err := validateSystemRepaveApproval(d, c)
	if err != nil {
		return diag.FromErr(err)
	}
	cloudConfigId := d.Get("cloud_config_id").(string)

	if d.HasChange("cloud_config") {
		cloudConfig := d.Get("cloud_config").([]interface{})[0].(map[string]interface{})
		cloudConfigEntity := toCloudConfigEks(cloudConfig)
		err := c.UpdateCloudConfigEks(cloudConfigId, cloudConfigEntity)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	CloudConfig, err := c.GetCloudConfigEks(cloudConfigId)
	if err != nil {
		return diag.FromErr(err)
	}
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

		err := c.UpdateFargateProfilesEks(cloudConfigId, fargateProfilesList)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	_ = d.Get("machine_pool")

	if d.HasChange("machine_pool") {
		log.Printf("[DEBUG] === MACHINE POOL CHANGE DETECTED ===")
		oraw, nraw := d.GetChange("machine_pool")
		if oraw == nil {
			oraw = new(schema.Set)
		}
		if nraw == nil {
			nraw = new(schema.Set)
		}

		os := oraw.(*schema.Set)
		ns := nraw.(*schema.Set)

		log.Printf("[DEBUG] Old machine pools count: %d, New machine pools count: %d", os.Len(), ns.Len())

		// Create maps by machine pool name for proper comparison
		osMap := make(map[string]interface{})
		for _, mp := range os.List() {
			machinePoolResource := mp.(map[string]interface{})
			name := machinePoolResource["name"].(string)
			if name != "" {
				osMap[name] = machinePoolResource
			}
		}

		nsMap := make(map[string]interface{})
		for _, mp := range ns.List() {
			machinePoolResource := mp.(map[string]interface{})
			name := machinePoolResource["name"].(string)
			if name != "" {
				nsMap[name] = machinePoolResource

				// Check if this is a new, updated, or unchanged machine pool
				if oldMachinePool, exists := osMap[name]; !exists {
					// NEW machine pool - CREATE
					log.Printf("[DEBUG] Creating new machine pool %s", name)
					machinePool := toMachinePoolEks(machinePoolResource)
					if err := c.CreateMachinePoolEks(cloudConfigId, machinePool); err != nil {
						return diag.FromErr(err)
					}
				} else {
					// EXISTING machine pool - check if hash changed
					oldHash := resourceMachinePoolEksHash(oldMachinePool)
					newHash := resourceMachinePoolEksHash(machinePoolResource)

					if oldHash != newHash {
						// MODIFIED machine pool - UPDATE
						log.Printf("[DEBUG] Updating machine pool %s (hash changed: %d -> %d)", name, oldHash, newHash)
						machinePool := toMachinePoolEks(machinePoolResource)
						if err := c.UpdateMachinePoolEks(cloudConfigId, machinePool); err != nil {
							return diag.FromErr(err)
						}
						// Node Maintenance Actions
						err := resourceNodeAction(c, ctx, machinePoolResource, c.GetNodeMaintenanceStatusEks, CloudConfig.Kind, cloudConfigId, name)
						if err != nil {
							return diag.FromErr(err)
						}
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
			if err := c.DeleteMachinePoolEks(cloudConfigId, name); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	diagnostics, done := updateCommonFields(d, c)
	if done {
		return diagnostics
	}

	resourceClusterEksRead(ctx, d, m)

	return diags
}

// to create
func toEksCluster(c *client.V1Client, d *schema.ResourceData) (*models.V1SpectroEksClusterEntity, error) {
	// gnarly, I know! =/
	cloudConfig := d.Get("cloud_config").([]interface{})[0].(map[string]interface{})
	//clientSecret := strfmt.Password(d.Get("Eks_client_secret").(string))
	var encryptionConfig *models.V1EncryptionConfig

	if cloudConfig["encryption_config_arn"] != nil {
		encryptionConfig = &models.V1EncryptionConfig{
			IsEnabled: true,
			Provider:  cloudConfig["encryption_config_arn"].(string),
		}
	}

	clusterContext := d.Get("context").(string)
	profiles, err := toProfiles(c, d, clusterContext)
	if err != nil {
		return nil, err
	}
	cluster := &models.V1SpectroEksClusterEntity{
		Metadata: getClusterMetadata(d),
		Spec: &models.V1SpectroEksClusterEntitySpec{
			CloudAccountUID: types.Ptr(d.Get("cloud_account_id").(string)),
			Profiles:        profiles,
			ClusterTemplate: toClusterTemplateReference(d),
			Policies:        toPolicies(d),
			CloudConfig: &models.V1EksClusterConfig{
				BastionDisabled:  true,
				VpcID:            cloudConfig["vpc_id"].(string),
				Region:           types.Ptr(cloudConfig["region"].(string)),
				SSHKeyName:       cloudConfig["ssh_key_name"].(string),
				EncryptionConfig: encryptionConfig,
			},
		},
	}

	// handling to tags_map for eks cluster
	if _, ok := d.GetOk("tags_map"); ok {
		tagMaps := toTagsMap(d)
		cluster.Metadata.Labels = tagMaps
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

	if cloudConfig["private_access_cidrs"] != nil {
		cidrs := make([]string, 0, 1)
		for _, cidr := range cloudConfig["private_access_cidrs"].(*schema.Set).List() {
			cidrs = append(cidrs, cidr.(string))
		}
		access.PrivateCIDRs = cidrs
	}

	cluster.Spec.CloudConfig.EndpointAccess = access

	machinePoolConfigs := make([]*models.V1EksMachinePoolConfigEntity, 0)

	// Following same logic as UI for setting up control plane for static cluster
	// Only add cp-pool for dynamic cluster provisioning when az_subnets is not empty and has more than one element
	if cloudConfig["az_subnets"] != nil && len(cloudConfig["az_subnets"].(map[string]interface{})) > 0 {
		cpPool := map[string]interface{}{
			"control_plane": true,
			"name":          "cp-pool",
			"az_subnets":    cloudConfig["az_subnets"],
			"capacity_type": "spot",
			"count":         0,
		}
		machinePoolConfigs = append(machinePoolConfigs, toMachinePoolEks(cpPool))
	}

	for _, machinePool := range d.Get("machine_pool").(*schema.Set).List() {
		mp := toMachinePoolEks(machinePool)
		machinePoolConfigs = append(machinePoolConfigs, mp)
	}

	cluster.Spec.Machinepoolconfig = machinePoolConfigs
	cluster.Spec.ClusterConfig = toClusterConfig(d)

	fargateProfiles := make([]*models.V1FargateProfile, 0)
	for _, fargateProfile := range d.Get("fargate_profile").([]interface{}) {
		f := toFargateProfileEks(fargateProfile)
		fargateProfiles = append(fargateProfiles, f)
	}

	cluster.Spec.FargateProfiles = fargateProfiles

	return cluster, nil
}

func toMachinePoolEks(machinePool interface{}) *models.V1EksMachinePoolConfigEntity {
	m := machinePool.(map[string]interface{})

	labels := make([]string, 0)
	controlPlane, _ := m["control_plane"].(bool)
	if controlPlane {
		labels = append(labels, "control-plane")
	} else {
		labels = append(labels, "worker")
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
	if len(azs) == 0 {
		if _, ok := m["azs"]; ok {
			for _, az := range m["azs"].([]interface{}) {
				azs = append(azs, az.(string))
			}
		}
	}

	capacityType := "on-demand" // on-demand by default.
	if m["capacity_type"] != nil && len(m["capacity_type"].(string)) > 0 {
		capacityType = m["capacity_type"].(string)
	}

	min := SafeInt32(m["count"].(int))
	max := SafeInt32(m["count"].(int))

	if m["min"] != nil {
		min = SafeInt32(m["min"].(int))
	}

	if m["max"] != nil {
		max = SafeInt32(m["max"].(int))
	}
	instanceType := ""
	if val, ok := m["instance_type"]; ok {
		instanceType = val.(string)
	}
	amiType := ""
	if val, ok := m["ami_type"]; ok {
		amiType = val.(string)
	}
	diskSizeGb := SafeInt64(0)
	if dVal, ok := m["disk_size_gb"]; ok {
		diskSizeGb = SafeInt64(dVal.(int))
	}
	mp := &models.V1EksMachinePoolConfigEntity{
		CloudConfig: &models.V1EksMachineCloudConfigEntity{
			RootDeviceSize: diskSizeGb,
			InstanceType:   instanceType,
			CapacityType:   &capacityType,
			Azs:            azs,
			Subnets:        subnets,
			AmiType:        amiType,
		},
		PoolConfig: &models.V1MachinePoolConfigEntity{
			AdditionalLabels: toAdditionalNodePoolLabels(m),
			Taints:           toClusterTaints(m),
			IsControlPlane:   controlPlane,
			Labels:           labels,
			Name:             types.Ptr(m["name"].(string)),
			Size:             types.Ptr(SafeInt32(m["count"].(int))),
			UpdateStrategy: &models.V1UpdateStrategy{
				Type: getUpdateStrategy(m),
			},
			MinSize: min,
			MaxSize: max,
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

	mp.CloudConfig.AwsLaunchTemplate = setEksLaunchTemplate(m)

	return mp
}

func setEksLaunchTemplate(machinePool map[string]interface{}) *models.V1AwsLaunchTemplate {
	var launchTemplate *models.V1AwsLaunchTemplate

	if machinePool["eks_launch_template"] != nil {
		eksLaunchTemplateList := machinePool["eks_launch_template"].([]interface{})
		if len(eksLaunchTemplateList) == 0 {
			return nil
		}

		eksLaunchTemplate := eksLaunchTemplateList[0].(map[string]interface{})

		keys := []string{"ami_id", "root_volume_type", "root_volume_iops", "root_volume_throughput", "additional_security_groups"}

		// if at least one key is present continue function body, otherwise return launchTemplate
		if hasNoneOfKeys(eksLaunchTemplate, keys) {
			return launchTemplate
		}

		launchTemplate = &models.V1AwsLaunchTemplate{
			RootVolume: &models.V1AwsRootVolume{},
		}

		if eksLaunchTemplate["ami_id"] != nil {
			launchTemplate.Ami = &models.V1AwsAmiReference{
				ID: eksLaunchTemplate["ami_id"].(string),
			}
		}

		if eksLaunchTemplate["root_volume_type"] != nil {
			launchTemplate.RootVolume.Type = eksLaunchTemplate["root_volume_type"].(string)
		}

		if eksLaunchTemplate["root_volume_iops"] != nil {
			launchTemplate.RootVolume.Iops = int64(eksLaunchTemplate["root_volume_iops"].(int))
		}

		if eksLaunchTemplate["root_volume_throughput"] != nil {
			launchTemplate.RootVolume.Throughput = int64(eksLaunchTemplate["root_volume_throughput"].(int))
		}

		launchTemplate.AdditionalSecurityGroups = setAdditionalSecurityGroups(eksLaunchTemplate)
	}

	return launchTemplate
}

func setAdditionalSecurityGroups(eksLaunchTemplate map[string]interface{}) []*models.V1AwsResourceReference {
	if eksLaunchTemplate["additional_security_groups"] != nil {
		securityGroups := expandStringList(eksLaunchTemplate["additional_security_groups"].(*schema.Set).List())
		additionalSecurityGroups := make([]*models.V1AwsResourceReference, 0)
		for _, securityGroup := range securityGroups {
			additionalSecurityGroups = append(additionalSecurityGroups, &models.V1AwsResourceReference{
				ID: securityGroup,
			})
		}
		return additionalSecurityGroups
	}

	return nil
}

func hasNoneOfKeys(m map[string]interface{}, keys []string) bool {
	for _, key := range keys {
		if m[key] != nil {
			return false
		}
	}
	return true
}

func toFargateProfileEks(fargateProfile interface{}) *models.V1FargateProfile {
	m := fargateProfile.(map[string]interface{})

	selectors := make([]*models.V1FargateSelector, 0)
	for _, val := range m["selector"].([]interface{}) {
		s := val.(map[string]interface{})

		selectors = append(selectors, &models.V1FargateSelector{
			Labels:    expandStringMap(s["labels"].(map[string]interface{})),
			Namespace: types.Ptr(s["namespace"].(string)),
		})
	}

	f := &models.V1FargateProfile{
		Name:           types.Ptr(m["name"].(string)),
		AdditionalTags: expandStringMap(m["additional_tags"].(map[string]interface{})),
		Selectors:      selectors,
		SubnetIds:      expandStringList(m["subnets"].([]interface{})),
	}

	return f
}

func toCloudConfigEks(cloudConfig map[string]interface{}) *models.V1EksCloudClusterConfigEntity {
	var encryptionConfig *models.V1EncryptionConfig
	if cloudConfig["encryption_config_arn"] != nil && cloudConfig["encryption_config_arn"].(string) != "" {
		encryptionConfig = &models.V1EncryptionConfig{
			IsEnabled: true,
			Provider:  cloudConfig["encryption_config_arn"].(string),
		}
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
		cidrs := make([]string, 0)
		for _, cidr := range cloudConfig["public_access_cidrs"].(*schema.Set).List() {
			cidrs = append(cidrs, cidr.(string))
		}
		access.PublicCIDRs = cidrs
	}

	if cloudConfig["private_access_cidrs"] != nil {
		cidrs := make([]string, 0)
		for _, cidr := range cloudConfig["private_access_cidrs"].(*schema.Set).List() {
			cidrs = append(cidrs, cidr.(string))
		}
		access.PrivateCIDRs = cidrs
	}

	clusterConfigEntity := &models.V1EksCloudClusterConfigEntity{
		ClusterConfig: &models.V1EksClusterConfig{
			BastionDisabled:  true,
			VpcID:            cloudConfig["vpc_id"].(string),
			Region:           types.Ptr(cloudConfig["region"].(string)),
			SSHKeyName:       cloudConfig["ssh_key_name"].(string),
			EncryptionConfig: encryptionConfig,
			EndpointAccess:   access,
		},
	}

	return clusterConfigEntity
}

func resourceClusterEksResourceV2() *schema.Resource {
	return &schema.Resource{
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
				Description: "The context of the EKS cluster. Allowed values are `project` or `tenant`. " +
					"Default is `project`. " + PROJECT_NAME_NUANCE,
			},
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      schema.HashString,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "A list of tags to be applied to the cluster. Tags must be in the form of `key:value`. The `tags` attribute will soon be deprecated. It is recommended to use `tags_map` instead.",
			},
			"tags_map": {
				Type:          schema.TypeMap,
				Optional:      true,
				ConflictsWith: []string{"tags"},
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "A map of tags to be applied to the cluster. tags and tags_map are mutually exclusive — only one should be used at a time",
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
				Description: "The AWS cloud account id to use for this cluster.",
			},
			"cloud_config_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the cloud config used for the cluster. This cloud config must be of type `azure`.",
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
				Description:      "Date and time after which to patch cluster `RFC3339: 2006-01-02T15:04:05Z07:00`",
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
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Description: "The AWS environment configuration settings such as network parameters and encryption parameters that apply to this cluster.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ssh_key_name": {
							Type:        schema.TypeString,
							ForceNew:    true,
							Optional:    true,
							Description: "Public SSH key to be used for the cluster nodes.",
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
							Type:        schema.TypeList,
							Description: "Mutually exclusive with `az_subnets`. Use for Dynamic provisioning.",
							Optional:    true,
							ForceNew:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"az_subnets": {
							Type:        schema.TypeMap,
							Description: "Mutually exclusive with `azs`. Use for Static provisioning.",
							Optional:    true,
							ForceNew:    true,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								// UI strips the trailing newline on save
								return strings.TrimSpace(old) == strings.TrimSpace(new)
							},
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"endpoint_access": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice([]string{"public", "private", "private_and_public"}, false),
							Description:  "Choose between `private`, `public`, or `private_and_public` to define how communication is established with the endpoint for the managed Kubernetes API server and your cluster. The default value is `public`.",
							Default:      "public",
						},
						"public_access_cidrs": {
							Type:        schema.TypeSet,
							Optional:    true,
							Set:         schema.HashString,
							Description: "List of CIDR blocks that define the allowed public access to the resource. Requests originating from addresses within these CIDR blocks will be permitted to access the resource. All other addresses will be denied access.",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"private_access_cidrs": {
							Type:        schema.TypeSet,
							Optional:    true,
							Set:         schema.HashString,
							Description: "List of CIDR blocks that define the allowed private access to the resource. Only requests originating from addresses within these CIDR blocks will be permitted to access the resource.",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"encryption_config_arn": {
							Type:        schema.TypeString,
							Description: "The ARN of the KMS encryption key to use for the cluster. Refer to the [Enable Secrets Encryption for EKS Cluster](https://docs.spectrocloud.com/clusters/public-cloud/aws/enable-secrets-encryption-kms-key/) for additional guidance.",
							ForceNew:    true,
							Optional:    true,
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
						"disk_size_gb": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"count": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Number of nodes in the machine pool.",
						},
						"update_strategy": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "RollingUpdateScaleOut",
							Description: "Update strategy for the machine pool. Valid values are `RollingUpdateScaleOut` and `RollingUpdateScaleIn`.",
						},
						"min": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Minimum number of nodes in the machine pool. This is used for autoscaling the machine pool.",
						},
						"max": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Maximum number of nodes in the machine pool. This is used for autoscaling the machine pool.",
						},
						"instance_type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"ami_type": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "AL2_x86_64",
							Description: "Specifies the type of Amazon Machine Image (AMI) to use for the machine pool. Valid values are [`AL2_x86_64`, `AL2_x86_64_GPU`, `AL2023_x86_64_STANDARD`, `AL2023_x86_64_NEURON` and `AL2023_x86_64_NVIDIA`]. Defaults to `AL2_x86_64`.",
						},
						"capacity_type": {
							Type:         schema.TypeString,
							Default:      "on-demand",
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"on-demand", "spot"}, false),
							Description:  "Capacity type is an instance type,  can be 'on-demand' or 'spot'. Defaults to 'on-demand'.",
						},
						"max_price": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"azs": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Mutually exclusive with `az_subnets`.",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"az_subnets": {
							Type:        schema.TypeMap,
							Optional:    true,
							Description: "Mutually exclusive with `azs`. Use for Static provisioning.",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"eks_launch_template": schemas.AwsLaunchTemplate(),
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

func resourceClusterEksStateUpgradeV2(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	log.Printf("[DEBUG] Upgrading cluster custom cloud state from version 2 to 3")

	// Convert machine_pool from TypeList to TypeSet
	// Note: We keep the data as a list in rawState and let Terraform's schema processing
	// convert it to TypeSet during normal resource loading. This avoids JSON serialization
	// issues with schema.Set objects that contain hash functions.
	if machinePoolRaw, exists := rawState["machine_pool"]; exists {
		if machinePoolList, ok := machinePoolRaw.([]interface{}); ok {
			log.Printf("[DEBUG] Keeping machine_pool as list during state upgrade with %d items", len(machinePoolList))

			// Keep the machine pool data as-is (as a list)
			// Terraform will convert it to TypeSet when loading the resource using the schema
			rawState["machine_pool"] = machinePoolList

			log.Printf("[DEBUG] Successfully prepared machine_pool for TypeSet conversion")
		} else {
			log.Printf("[DEBUG] machine_pool is not a list, skipping conversion")
		}
	} else {
		log.Printf("[DEBUG] No machine_pool found in state, skipping conversion")
	}

	return rawState, nil
}

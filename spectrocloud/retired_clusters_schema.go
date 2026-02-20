package spectrocloud

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/schemas"
)

// resourceClusterCustomCloudResourceV2 returns the schema for version 2 of the resource
func resourceClusterCustomCloudResourceV2() *schema.Resource {
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
			"cloud": {
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
				Description: "The cloud provider name.",
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
				Description: "The cloud account id to use for this cluster.",
			},
			"cloud_config_id": {
				Type:       schema.TypeString,
				Computed:   true,
				Deprecated: "This field is deprecated and will be removed in the future. Use `cloud_config` instead.",
			},
			"cluster_timezone": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "",
				ValidateFunc: validateTimezone,
				Description:  "Defines the time zone used by this cluster to interpret scheduled operations. Maintenance tasks like upgrades will follow this time zone to ensure they run at the appropriate local time for the cluster. Must be in IANA timezone format (e.g., 'America/New_York', 'Asia/Kolkata', 'Europe/London').",
			},
			"cloud_config": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Description: "The Cloud environment configuration settings such as network parameters and encryption parameters that apply to this cluster.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"values": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The values of the cloud config. The values are specified in YAML format. ",
						},
						"overrides": {
							Type:        schema.TypeMap,
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Key-value pairs to override specific values in the YAML.",
						},
					},
				},
			},
			// Version 2 used TypeList for machine_pool
			"machine_pool": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "The machine pool configuration for the cluster.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the machine pool. This will be derived from the name value in the `node_pool_config`.",
						},
						"count": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of nodes in the machine pool. This will be derived from the replica value in the 'node_pool_config'.",
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
						"node_pool_config": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The values of the node pool config. The values are specified in YAML format. ",
						},
						"overrides": {
							Type:        schema.TypeMap,
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Key-value pairs to override specific values in the node pool config YAML.",
						},
					},
				},
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
				Description:      "The cron schedule for OS patching. This must be in the form of cron syntax. Ex: `0 0 * * *`.",
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
			"backup_policy":        schemas.BackupPolicySchema(),
			"scan_policy":          schemas.ScanPolicySchema(),
			"cluster_rbac_binding": schemas.ClusterRbacBindingSchema(),
			"namespaces":           schemas.ClusterNamespacesSchema(),
			"location_config":      schemas.ClusterLocationSchema(),
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
			"cluster_timezone": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "",
				ValidateFunc: validateTimezone,
				Description:  "Defines the time zone used by this cluster to interpret scheduled operations. Maintenance tasks like upgrades will follow this time zone to ensure they run at the appropriate local time for the cluster. Must be in IANA timezone format (e.g., 'America/New_York', 'Asia/Kolkata', 'Europe/London').",
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
							Description: "Additional labels to be applied to the machine pool. Labels must be in the form of `key:value`.",
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
							Default:     "AL2023_x86_64_STANDARD",
							Description: "Specifies the type of Amazon Machine Image (AMI) to use for the machine pool. Valid values are [`AL2_x86_64`, `AL2_x86_64_GPU`, `AL2023_x86_64_STANDARD`, `AL2023_x86_64_NEURON` and `AL2023_x86_64_NVIDIA`]. Defaults to `AL2023_x86_64_STANDARD`.",
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

// resourceClusterEksResourceV3 returns the schema for version 3 of the EKS cluster resource.
// Version 3 matches the previous live schema before SchemaVersion 4:
// - machine_pool is TypeSet
// - cluster_profile is TypeList
func resourceClusterEksResourceV3() *schema.Resource {
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
			"cluster_timezone": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "",
				ValidateFunc: validateTimezone,
				Description:  "Defines the time zone used by this cluster to interpret scheduled operations. Maintenance tasks like upgrades will follow this time zone to ensure they run at the appropriate local time for the cluster. Must be in IANA timezone format (e.g., 'America/New_York', 'Asia/Kolkata', 'Europe/London').",
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
							Description: "Additional labels to be applied to the machine pool. Labels must be in the form of `key:value`.",
						},
						"additional_annotations": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "Additional annotations to be applied to the machine pool. Annotations must be in the form of `key:value`.",
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
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "RollingUpdateScaleOut",
							Description:  "Update strategy for the machine pool. Valid values are `RollingUpdateScaleOut`, `RollingUpdateScaleIn` and `OverrideScaling`. If `OverrideScaling` is used, `override_scaling` must be specified with both `max_surge` and `max_unavailable`.",
							ValidateFunc: validation.StringInSlice([]string{"RollingUpdateScaleOut", "RollingUpdateScaleIn", "OverrideScaling"}, false),
						},
						"override_scaling": schemas.OverrideScalingSchema(),
						"override_kubeadm_configuration": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "YAML config for kubeletExtraArgs, preKubeadmCommands, postKubeadmCommands. Overrides pack-level settings. Worker pools only.",
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
							Default:     "AL2023_x86_64_STANDARD",
							Description: "Specifies the type of Amazon Machine Image (AMI) to use for the machine pool. Valid values are [`AL2_x86_64`, `AL2_x86_64_GPU`, `AL2023_x86_64_STANDARD`, `AL2023_x86_64_NEURON` and `AL2023_x86_64_NVIDIA`]. Defaults to `AL2023_x86_64_STANDARD`.",
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

func resourceClusterAksResourceV2() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "",
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"", "project", "tenant"}, false),
				Description: "The context of the AKS cluster. Allowed values are `project` or `tenant`. " +
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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
				Description:      "The cron schedule for OS patching. This must be in the form of cron syntax. Ex: `0 0 * * *`.",
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
				ForceNew: true,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subscription_id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"resource_group": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"region": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"ssh_key": {
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
							Description: "Public SSH key to be used for the cluster nodes.",
						},
						"private_cluster": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							ForceNew:    true,
							Description: "Whether to create a private cluster(API endpoint). Default is `false`.",
						},

						// fields for static placement are having flat structure as backend currently doesn't support multiple subnets.
						"vnet_name": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},

						"vnet_resource_group": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},

						"vnet_cidr_block": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},

						"worker_subnet_name": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"worker_cidr": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"worker_subnet_security_group_name": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"control_plane_subnet_name": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"control_plane_cidr": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"control_plane_subnet_security_group_name": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
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
						"node":   schemas.NodeSchema(),
						"taints": schemas.ClusterTaintsSchema(),
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
						"instance_type": {
							Type:     schema.TypeString,
							Required: true,
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
						"disk_size_gb": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"is_system_node_pool": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"storage_account_type": {
							Type:     schema.TypeString,
							Required: true,
							//ExactlyOneOf: []string{"Standard_LRS", "Standard_GRS", "Standard_RAGRS", "Standard_ZRS", "Premium_LRS", "Premium_ZRS", "Standard_GZRS", "Standard_RAGZRS"},
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

// resourceClusterAksResourceV3 returns the schema for version 3 of the AKS cluster resource.
// Version 3 matches the previous live schema before SchemaVersion 4:
// - machine_pool is TypeSet
// - cluster_profile is TypeList
func resourceClusterAksResourceV3() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "",
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"", "project", "tenant"}, false),
				Description: "The context of the AKS cluster. Allowed values are `project` or `tenant`. " +
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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
				Description:      "The cron schedule for OS patching. This must be in the form of cron syntax. Ex: `0 0 * * *`.",
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
				ForceNew: true,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subscription_id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"resource_group": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"region": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"ssh_key": {
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
							Description: "Public SSH key to be used for the cluster nodes.",
						},
						"private_cluster": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							ForceNew:    true,
							Description: "Whether to create a private cluster(API endpoint). Default is `false`.",
						},
						"vnet_name": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"vnet_resource_group": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"vnet_cidr_block": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"worker_subnet_name": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"worker_cidr": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"worker_subnet_security_group_name": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"control_plane_subnet_name": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"control_plane_cidr": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"control_plane_subnet_security_group_name": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},
			"machine_pool": {
				Type:     schema.TypeSet,
				Required: true,
				Set:      resourceMachinePoolAksHash,
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
							Description: "Additional labels to be applied to the machine pool. Labels must be in the form of `key:value`.",
						},
						"additional_annotations": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "Additional annotations to be applied to the machine pool. Annotations must be in the form of `key:value`.",
						},
						"node":   schemas.NodeSchema(),
						"taints": schemas.ClusterTaintsSchema(),
						"count": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Number of nodes in the machine pool.",
						},
						"update_strategy": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "RollingUpdateScaleOut",
							Description:  "Update strategy for the machine pool. Valid values are `RollingUpdateScaleOut`, `RollingUpdateScaleIn` and `OverrideScaling`. If `OverrideScaling` is used, `override_scaling` must be specified with both `max_surge` and `max_unavailable`.",
							ValidateFunc: validation.StringInSlice([]string{"RollingUpdateScaleOut", "RollingUpdateScaleIn", "OverrideScaling"}, false),
						},
						"override_scaling": schemas.OverrideScalingSchema(),
						"override_kubeadm_configuration": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "YAML config for kubeletExtraArgs, preKubeadmCommands, postKubeadmCommands. Overrides pack-level settings. Worker pools only.",
						},
						"instance_type": {
							Type:     schema.TypeString,
							Required: true,
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
						"disk_size_gb": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"is_system_node_pool": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"storage_account_type": {
							Type:     schema.TypeString,
							Required: true,
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

func resourceClusterGkeResourceV1() *schema.Resource {
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
				Description: "The context of the GKE cluster. Allowed values are `project` or `tenant`. " +
					"Default is `project`. " + PROJECT_NAME_NUANCE,
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The description of the cluster. Default value is empty string.",
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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cloud_config_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the cloud config used for the cluster. This cloud config must be of type `azure`.",
				Deprecated:  "This field is deprecated and will be removed in the future. Use `cloud_config` instead.",
			},

			"cloud_config": {
				Type:        schema.TypeList,
				ForceNew:    true,
				Required:    true,
				MaxItems:    1,
				Description: "The GKE environment configuration settings such as project parameters and region parameters that apply to this cluster.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"project": {
							Type:        schema.TypeString,
							ForceNew:    true,
							Required:    true,
							Description: "GCP project name.",
						},
						"region": {
							Type:     schema.TypeString,
							ForceNew: true,
							Required: true,
						},
					},
				},
			},
			"update_worker_pool_in_parallel": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
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
						},
						"count": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Number of nodes in the machine pool.",
						},
						"disk_size_gb": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  60,
						},
						"additional_labels": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "Additional labels to be applied to the machine pool. Labels must be in the form of `key:value`.",
						},
						"instance_type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"update_strategy": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "RollingUpdateScaleOut",
							Description: "Update strategy for the machine pool. Valid values are `RollingUpdateScaleOut` and `RollingUpdateScaleIn`.",
						},
						"node":   schemas.NodeSchema(),
						"taints": schemas.ClusterTaintsSchema(),
					},
				},
			},
			"pause_agent_upgrades": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "unlock",
				ValidateFunc: validation.StringInSlice([]string{"lock", "unlock"}, false),
				Description:  "The pause agent upgrades setting allows to control the automatic upgrade of the Palette component and agent for an individual cluster. The default value is `unlock`, meaning upgrades occur automatically. Setting it to `lock` pauses automatic agent upgrades for the cluster.",
			},
			"review_repave_state": {
				Type:         schema.TypeString,
				Default:      "",
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"", "Approved", "Pending"}, false),
				Description:  "To authorize the cluster repave, set the value to `Approved` for approval and `\"\"` to decline. Default value is `\"\"`.",
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

func resourceClusterOpenStackResourceV1() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"", "project", "tenant"}, false),
				Description: "The context of the OpenStack cluster. Allowed values are `project` or `tenant`. " +
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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
				Description:      "The date and time after which to patch the cluster. Prefix the time value with the respective RFC. Ex: `RFC3339: 2006-01-02T15:04:05Z07:00`",
			},
			"cluster_timezone": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "",
				ValidateFunc: validateTimezone,
				Description:  "Defines the time zone used by this cluster to interpret scheduled operations. Maintenance tasks like upgrades will follow this time zone to ensure they run at the appropriate local time for the cluster. Must be in IANA timezone format (e.g., 'America/New_York', 'Asia/Kolkata', 'Europe/London').",
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
				ForceNew: true,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"domain": {
							Type:     schema.TypeString,
							Required: true,
						},
						"region": {
							Type:     schema.TypeString,
							Required: true,
						},
						"project": {
							Type:     schema.TypeString,
							Required: true,
						},
						"ssh_key": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Public SSH key to be used for the cluster nodes.",
						},
						"network_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"subnet_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"dns_servers": {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							Set:      schema.HashString,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"subnet_cidr": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"machine_pool": {
				Type:        schema.TypeList, // V2: TypeList, V3: TypeSet
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
							Description: "Additional labels to be applied to the machine pool. Labels must be in the form of `key:value`.",
						},
						"taints": schemas.ClusterTaintsSchema(),
						"node":   schemas.NodeSchema(),
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
							Type:     schema.TypeString,
							Required: true,
						},
						"count": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Number of nodes in the machine pool.",
						},
						"node_repave_interval": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     0,
							Description: "Minimum number of seconds node should be Ready, before the next node is selected for repave. Default value is `0`, Applicable only for worker pools.",
						},
						"update_strategy": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "RollingUpdateScaleOut",
							Description:  "Update strategy for the machine pool. Valid values are `RollingUpdateScaleOut` and `RollingUpdateScaleIn`.",
							ValidateFunc: validation.StringInSlice([]string{"RollingUpdateScaleOut", "RollingUpdateScaleIn"}, false),
						},
						"instance_type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"azs": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"subnet_id": {
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
			"host_config":          schemas.ClusterHostConfigSchema(),
			"location_config":      schemas.ClusterLocationSchema(),
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

// resourceClusterEdgeNativeResourceV2 returns the V2 schema with edge_host as TypeList
func resourceClusterEdgeNativeResourceV2() *schema.Resource {
	return &schema.Resource{
		SchemaVersion: 2,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"", "project", "tenant"}, false),
				Description: "The context of the Edge cluster. Allowed values are `project` or `tenant`. " +
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
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
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
				Description:      "The cron schedule for OS patching. This must be in the form of cron syntax. Ex: `0 0 * * *`.",
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
				Type:     schema.TypeList,
				ForceNew: true,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ssh_keys": {
							Type:     schema.TypeSet,
							Optional: true,
							Set:      schema.HashString,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "List of public SSH (Secure Shell) to establish, administer, and communicate with remote clusters.",
						},
						"vip": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The `vip` can be specified as either an IP address or a fully qualified domain name (FQDN). If `overlay_cidr_range` is set, the `vip` should be within the specified `overlay_cidr_range`. By default, the `vip` is set to the first IP address within the given `overlay_cidr_range`.",
						},
						"overlay_cidr_range": {
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
							Description: "The Overlay (VPN) creates a virtual network, using techniques like VxLAN. It overlays the existing network infrastructure, enhancing connectivity either at Layer 2 or Layer 3, making it flexible and adaptable for various needs. For example, `100.64.192.0/24`",
						},
						"is_two_node_cluster": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Set to `true` to enable a two-node cluster.",
						},
						"ntp_servers": {
							Type:     schema.TypeSet,
							Optional: true,
							Set:      schema.HashString,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "A list of NTP servers to be used by the cluster.",
						},
					},
				},
			},
			"machine_pool": {
				Type:     schema.TypeSet,
				Required: true,
				Set:      resourceMachinePoolEdgeNativeHash,
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
						"node_repave_interval": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     0,
							Description: "Minimum number of seconds node should be Ready, before the next node is selected for repave. Default value is `0`, Applicable only for worker pools.",
						},
						"update_strategy": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "RollingUpdateScaleOut",
							Description:  "Update strategy for the machine pool. Valid values are `RollingUpdateScaleOut` and `RollingUpdateScaleIn`.",
							ValidateFunc: validation.StringInSlice([]string{"RollingUpdateScaleOut", "RollingUpdateScaleIn"}, false),
						},
						"edge_host": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"host_name": {
										Type:        schema.TypeString,
										Optional:    true,
										Default:     "",
										Description: "Edge host name",
									},
									"host_uid": {
										Type:        schema.TypeString,
										Description: "Edge host id",
										Required:    true,
									},
									"static_ip": {
										Type:        schema.TypeString,
										Description: "Edge host static IP address",
										Optional:    true,
									},
									"nic_name": {
										Type:        schema.TypeString,
										Description: "NIC Name for edge host.",
										Optional:    true,
									},
									"default_gateway": {
										Type:        schema.TypeString,
										Description: "Edge host default gateway",
										Optional:    true,
									},
									"subnet_mask": {
										Type:        schema.TypeString,
										Description: "Edge host subnet mask",
										Optional:    true,
									},
									"dns_servers": {
										Type:        schema.TypeSet,
										Optional:    true,
										Set:         schema.HashString,
										Description: "Edge host DNS servers",
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"two_node_role": {
										Type:         schema.TypeString,
										Description:  "Two node role for edge host. Valid values are `primary` and `secondary`.",
										Optional:     true,
										ValidateFunc: validation.StringInSlice([]string{"primary", "secondary"}, false),
									},
								},
							},
						},
					},
				},
			},
			"pause_agent_upgrades": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "unlock",
				ValidateFunc: validation.StringInSlice([]string{"lock", "unlock"}, false),
				Description:  "The pause agent upgrades setting allows to control the automatic upgrade of the Palette component and agent for an individual cluster. The default value is `unlock`, meaning upgrades occur automatically. Setting it to `lock` pauses automatic agent upgrades for the cluster.",
			},
			"backup_policy":        schemas.BackupPolicySchema(),
			"scan_policy":          schemas.ScanPolicySchema(),
			"cluster_rbac_binding": schemas.ClusterRbacBindingSchema(),
			"namespaces":           schemas.ClusterNamespacesSchema(),
			"host_config":          schemas.ClusterHostConfigSchema(),
			"location_config":      schemas.ClusterLocationSchema(),
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

// resourceClusterAwsResourceV2 returns the schema for version 2 of the AWS cluster resource.
// Version 2 used TypeList for cluster_profile.
func resourceClusterAwsResourceV2() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"", "project", "tenant"}, false),
				Description: "The context of the AWS cluster. Allowed values are `project` or `tenant`. " +
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
			"cluster_type":     schemas.ClusterTypeSchema(),
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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
				Description:      "The cron schedule for OS patching. This must be in the form of cron syntax. Ex: `0 0 * * *`.",
			},
			"os_patch_after": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateOsPatchOnDemandAfter,
				Description:      "Date and time after which to patch cluster `RFC3339: 2006-01-02T15:04:05Z07:00`",
			},
			"cluster_timezone": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "",
				ValidateFunc: validateTimezone,
				Description:  "Defines the time zone used by this cluster to interpret scheduled operations. Maintenance tasks like upgrades will follow this time zone to ensure they run at the appropriate local time for the cluster. Must be in IANA timezone format (e.g., 'America/New_York', 'Asia/Kolkata', 'Europe/London').",
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
				ForceNew: true,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ssh_key_name": {
							Type:        schema.TypeString,
							ForceNew:    true,
							Required:    true,
							Description: "Public SSH key to be used for the cluster nodes.",
						},
						"region": {
							Type:        schema.TypeString,
							ForceNew:    true,
							Required:    true,
							Description: "The AWS region to deploy the cluster in.",
						},
						"vpc_id": {
							Type:        schema.TypeString,
							ForceNew:    true,
							Optional:    true,
							Description: "The VPC ID to deploy the cluster in. If not provided, VPC will be provisioned dynamically.",
						},
						"control_plane_lb": {
							Type:         schema.TypeString,
							ForceNew:     true,
							Default:      "",
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"", "Internet-facing", "internal"}, false),
							Description:  "Control plane load balancer type. Valid values are `Internet-facing` and `internal`. Defaults to `` (empty string).",
						},
					},
				},
			},
			"machine_pool": {
				Type:     schema.TypeSet,
				Required: true,
				Set:      resourceMachinePoolAwsHash,
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
						"additional_annotations": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "Additional annotations to be applied to the machine pool. Annotations must be in the form of `key:value`.",
						},
						"taints": schemas.ClusterTaintsSchema(),
						"control_plane": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
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
							Type:        schema.TypeString,
							Required:    true,
							Description: "The name of the machine pool.",
						},
						"count": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Number of nodes in the machine pool.",
						},
						"instance_type": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The instance type to use for the machine pool nodes.",
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
						"node_repave_interval": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     0,
							Description: "Minimum number of seconds node should be Ready, before the next node is selected for repave. Default value is `0`, Applicable only for worker pools.",
						},
						"skip_k8s_upgrade": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "disabled",
							ValidateFunc: validation.StringInSlice([]string{"enabled", "disabled"}, false),
							Description:  "Skip Kubernetes version upgrade for this worker pool. Use 'enabled' to skip OS/K8s update on profile upgrade (N-3 skew allowed); 'disabled' to upgrade with profile (default). Applicable only for worker pools. The skip_k8s_upgrade field is available from Palette 4.8.b.",
						},
						"capacity_type": {
							Type:         schema.TypeString,
							Default:      "on-demand",
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"on-demand", "spot"}, false),
							Description:  "Capacity type is an instance type,  can be 'on-demand' or 'spot'. Defaults to 'on-demand'.",
						},
						"max_price": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Maximum price to bid for spot instances. Only applied when instance type is 'spot'.",
						},
						"update_strategy": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "RollingUpdateScaleOut",
							Description:  "Update strategy for the machine pool. Valid values are `RollingUpdateScaleOut`, `RollingUpdateScaleIn` and `OverrideScaling`. If `OverrideScaling` is used, `override_scaling` must be specified with both `max_surge` and `max_unavailable`.",
							ValidateFunc: validation.StringInSlice([]string{"RollingUpdateScaleOut", "RollingUpdateScaleIn", "OverrideScaling"}, false),
						},
						"override_scaling": schemas.OverrideScalingSchema(),
						"override_kubeadm_configuration": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "YAML config for kubeletExtraArgs, preKubeadmCommands, postKubeadmCommands. Overrides pack-level settings. Worker pools only.",
						},
						"disk_size_gb": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     65,
							Description: "The disk size in GB for the machine pool nodes.",
						},
						"azs": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "Mutually exclusive with `az_subnets`. Use `azs` for Dynamic provisioning.",
							MinItems:    1,
							Set:         schema.HashString,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"az_subnets": {
							Type:        schema.TypeMap,
							Optional:    true,
							Description: "Mutually exclusive with `azs`. Use `az_subnets` for Static provisioning.",
							Elem: &schema.Schema{
								Type:     schema.TypeString,
								Required: true,
							},
						},
						"additional_security_groups": {
							Type: schema.TypeSet,
							Set:  schema.HashString,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Optional:    true,
							Description: "Additional security groups to attach to the instance.",
						},
						"node": schemas.NodeSchema(),
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

// resourceClusterGcpResourceV2 returns the schema for version 2 of the GCP cluster resource.
// Version 2 used TypeList for cluster_profile.
func resourceClusterGcpResourceV2() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"", "project", "tenant"}, false),
				Description: "The context of the GCP cluster. Allowed values are `project` or `tenant`. " +
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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
			"cluster_timezone": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "",
				ValidateFunc: validateTimezone,
				Description:  "Defines the time zone used by this cluster to interpret scheduled operations. Maintenance tasks like upgrades will follow this time zone to ensure they run at the appropriate local time for the cluster. Must be in IANA timezone format (e.g., 'America/New_York', 'Asia/Kolkata', 'Europe/London').",
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
				ForceNew: true,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"network": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"project": {
							Type:     schema.TypeString,
							Required: true,
						},
						"region": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"machine_pool": {
				Type:     schema.TypeSet,
				Required: true,
				Set:      resourceMachinePoolGcpHash,
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
						"additional_annotations": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "Additional annotations to be applied to the machine pool. Annotations must be in the form of `key:value`.",
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
							Type:     schema.TypeString,
							Required: true,
						},
						"count": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Number of nodes in the machine pool.",
						},
						"node_repave_interval": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     0,
							Description: "Minimum number of seconds node should be Ready, before the next node is selected for repave. Default value is `0`, Applicable only for worker pools.",
						},
						"instance_type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"update_strategy": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "RollingUpdateScaleOut",
							Description:  "Update strategy for the machine pool. Valid values are `RollingUpdateScaleOut`, `RollingUpdateScaleIn` and `OverrideScaling`. If `OverrideScaling` is used, `override_scaling` must be specified with both `max_surge` and `max_unavailable`.",
							ValidateFunc: validation.StringInSlice([]string{"RollingUpdateScaleOut", "RollingUpdateScaleIn", "OverrideScaling"}, false),
						},
						"override_scaling": schemas.OverrideScalingSchema(),
						"override_kubeadm_configuration": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "YAML config for kubeletExtraArgs, preKubeadmCommands, postKubeadmCommands. Overrides pack-level settings. Worker pools only.",
						},
						"disk_size_gb": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  65,
						},
						"azs": {
							Type:     schema.TypeSet,
							Required: true,
							MinItems: 1,
							Set:      schema.HashString,
							Elem: &schema.Schema{
								Type: schema.TypeString,
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

// resourceClusterAzureResourceV0 returns the schema for version 0 of the Azure cluster resource.
// Version 0 used TypeList for cluster_profile.
func resourceClusterAzureResourceV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the cluster. This name will be used to create the cluster in Azure.",
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"", "project", "tenant"}, false),
				Description: "The context of the Azure cluster. Allowed values are `project` or `tenant`. " +
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
					"Default is `DownloadAndInstall`.",
			},
			"cloud_account_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the cloud account to be used for the cluster. This cloud account must be of type `azure`.",
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
			"cluster_timezone": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "",
				ValidateFunc: validateTimezone,
				Description:  "Defines the time zone used by this cluster to interpret scheduled operations. Maintenance tasks like upgrades will follow this time zone to ensure they run at the appropriate local time for the cluster. Must be in IANA timezone format (e.g., 'America/New_York', 'Asia/Kolkata', 'Europe/London').",
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
				ForceNew: true,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subscription_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Azure subscription ID. This can be found in the Azure portal under `Subscriptions`.",
						},
						"resource_group": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Azure resource group. This can be found in the Azure portal under `Resource groups`.",
						},
						"region": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Azure region. This can be found in the Azure portal under `Resource groups`.",
						},
						"ssh_key": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Public SSH key to be used for the cluster nodes.",
						},
						"storage_account_name": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "Azure storage account name.",
						},
						"container_name": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "Container name within your azure storage account.",
						},
						"network_resource_group": {
							Type:         schema.TypeString,
							Optional:     true,
							RequiredWith: []string{"cloud_config.0.virtual_network_name", "cloud_config.0.virtual_network_cidr_block", "cloud_config.0.control_plane_subnet", "cloud_config.0.worker_node_subnet"},
							Description:  "Azure network resource group in which the cluster is to be provisioned.",
						},
						"virtual_network_name": {
							Type:         schema.TypeString,
							Optional:     true,
							RequiredWith: []string{"cloud_config.0.network_resource_group", "cloud_config.0.virtual_network_cidr_block", "cloud_config.0.control_plane_subnet", "cloud_config.0.worker_node_subnet"},
							Description:  "Azure virtual network in which the cluster is to be provisioned.",
						},
						"virtual_network_cidr_block": {
							Type:         schema.TypeString,
							Optional:     true,
							RequiredWith: []string{"cloud_config.0.network_resource_group", "cloud_config.0.virtual_network_name", "cloud_config.0.control_plane_subnet", "cloud_config.0.worker_node_subnet"},
							Description:  "Azure virtual network cidr block in which the cluster is to be provisioned.",
						},
						"control_plane_subnet": schemas.SubnetSchema(),
						"worker_node_subnet":   schemas.SubnetSchema(),
						"private_api_server": {
							Type:         schema.TypeList,
							Optional:     true,
							MaxItems:     1,
							RequiredWith: []string{"cloud_config.0.network_resource_group", "cloud_config.0.virtual_network_name", "cloud_config.0.virtual_network_cidr_block"},
							Description:  "Custom private DNS zone for your cluster's API server. For more details, refer to the https://docs.spectrocloud.com/clusters/public-cloud/azure/create-azure-cluster/#private-api-server-lb-settings",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"resource_group": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The resource group of the private DNS zone.",
									},
									"private_dns_zone": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "The private DNS zone for the cluster. This is optional. If not provided, a new private DNS zone will be created.",
									},
									"static_ip": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Static IP address for the private API server load balancer. This is optional. If not provided, Dynamic IP allocation will be used.",
									},
								},
							},
						},
					},
				},
			},
			"machine_pool": {
				Type:     schema.TypeSet,
				Required: true,
				Set:      resourceMachinePoolAzureHash,
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
						"additional_annotations": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "Additional annotations to be applied to the machine pool. Annotations must be in the form of `key:value`.",
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
							Description: "Name of the machine pool. This must be unique within the cluster.",
						},
						"count": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Number of nodes in the machine pool.",
						},
						"node_repave_interval": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     0,
							Description: "Minimum number of seconds node should be Ready, before the next node is selected for repave. Default value is `0`, Applicable only for worker pools.",
						},
						"instance_type": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Azure instance type from the Azure portal.",
						},
						"update_strategy": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "RollingUpdateScaleOut",
							Description:  "Update strategy for the machine pool. Valid values are `RollingUpdateScaleOut`, `RollingUpdateScaleIn` and `OverrideScaling`. If `OverrideScaling` is used, `override_scaling` must be specified with both `max_surge` and `max_unavailable`.",
							ValidateFunc: validation.StringInSlice([]string{"RollingUpdateScaleOut", "RollingUpdateScaleIn", "OverrideScaling"}, false),
						},
						"override_scaling": schemas.OverrideScalingSchema(),
						"override_kubeadm_configuration": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "YAML config for kubeletExtraArgs, preKubeadmCommands, postKubeadmCommands. Overrides pack-level settings. Worker pools only.",
						},
						"disk": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"size_gb": {
										Type:        schema.TypeInt,
										Required:    true,
										Description: "Size of the disk in GB.",
									},
									"type": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Type of the disk. Valid values are `Standard_LRS`, `StandardSSD_LRS`, `Premium_LRS`.",
									},
								},
							},
							Description: "Disk configuration for the machine pool.",
						},
						"azs": {
							Type:     schema.TypeSet,
							Optional: true,
							Set:      schema.HashString,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "Availability zones for the machine pool. Check if your region provides availability zones on [the Azure documentation](https://learn.microsoft.com/en-us/azure/reliability/availability-zones-service-support#azure-regions-with-availability-zone-support). Default value is `[\"\"]`.",
							DefaultFunc: func() (any, error) {
								return []string{""}, nil
							},
						},
						"is_system_node_pool": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Whether this machine pool is a system node pool. Default value is `false'.",
						},
						"os_type": {
							Type:     schema.TypeString,
							Optional: true,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								return false
							},
							Default:      "Linux",
							ValidateFunc: validation.StringInSlice([]string{"Linux", "Windows"}, false),
							Description:  "Operating system type for the machine pool. Valid values are `Linux` and `Windows`. Defaults to `Linux`.",
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

// resourceClusterVsphereResourceV0 returns the schema for version 0 of the vSphere cluster resource.
// Version 0 used TypeList for cluster_profile.
func resourceClusterVsphereResourceV0() *schema.Resource {
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
				Description: "The context of the VMware cluster. Allowed values are `project` or `tenant`. " +
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
				Description: "ID of the cloud account to be used for the cluster. This cloud account must be of type `vsphere`.",
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
				Description:      "The cron schedule for OS patching. This must be in the form of cron syntax. Ex: `0 0 * * *`.",
			},
			"os_patch_after": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateOsPatchOnDemandAfter,
				Description:      "The date and time after which to patch the cluster. Prefix the time value with the respective RFC. Ex: `RFC3339: 2006-01-02T15:04:05Z07:00`",
			},
			"cluster_timezone": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "",
				ValidateFunc: validateTimezone,
				Description:  "Defines the time zone used by this cluster to interpret scheduled operations. Maintenance tasks like upgrades will follow this time zone to ensure they run at the appropriate local time for the cluster. Must be in IANA timezone format (e.g., 'America/New_York', 'Asia/Kolkata', 'Europe/London').",
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
				ForceNew: true,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"datacenter": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The name of the datacenter in vSphere. This is the name of the datacenter as it appears in vSphere.",
						},
						"folder": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The name of the folder in vSphere. This is the name of the folder as it appears in vSphere.",
						},
						"image_template_folder": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The name of the image template folder in vSphere. This is the name of the folder as it appears in vSphere.",
						},
						"ssh_key": {
							Type:         schema.TypeString,
							Optional:     true,
							ExactlyOneOf: []string{"cloud_config.0.ssh_key", "cloud_config.0.ssh_keys"},
							Description:  "The SSH key to be used for the cluster. This is the public key that will be used to access the cluster nodes. `ssh_key & ssh_keys` are mutually exclusive.",
							Deprecated:   "This field is deprecated and will be removed in the future. Use `ssh_keys` instead.",
						},
						"ssh_keys": {
							Type:         schema.TypeSet,
							Optional:     true,
							Set:          schema.HashString,
							ExactlyOneOf: []string{"cloud_config.0.ssh_key", "cloud_config.0.ssh_keys"},
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "List of public SSH (Secure Shell) keys to establish, administer, and communicate with remote clusters, `ssh_key & ssh_keys` are mutually exclusive.",
						},
						"static_ip": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
							Description: "Whether to use static IP addresses for the cluster. If `true`, the cluster will use static IP addresses. " +
								"If `false`, the cluster will use DDNS. Default is `false`.",
						},
						"network_type": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The type of network to use for the cluster. This can be `VIP` or `DDNS`.",
						},
						"host_endpoint": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The host endpoint to use for the cluster. This can be `IP` or `FQDN(External/DDNS)`.",
						},
						"network_search_domain": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The search domain to use for the cluster in case of DHCP.",
						},
						"ntp_servers": {
							Type:     schema.TypeSet,
							Optional: true,
							Set:      schema.HashString,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "A list of NTP servers to be used by the cluster.",
						},
					},
				},
			},
			"machine_pool": {
				Type:     schema.TypeSet,
				Required: true,
				Set:      resourceMachinePoolVsphereHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The name of the machine pool. This is used to identify the machine pool in the cluster.",
						},
						"additional_labels": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "Additional labels to be applied to the machine pool. Labels must be in the form of `key:value`.",
						},
						"additional_annotations": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "Additional annotations to be applied to the machine pool. Annotations must be in the form of `key:value`.",
						},
						"taints": schemas.ClusterTaintsSchema(),
						"node":   schemas.NodeSchema(),
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
						"node_repave_interval": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     0,
							Description: "Minimum number of seconds node should be Ready, before the next node is selected for repave. Default value is `0`, Applicable only for worker pools.",
						},
						"count": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Number of nodes in the machine pool.",
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
						"update_strategy": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "RollingUpdateScaleOut",
							Description:  "Update strategy for the machine pool. Valid values are `RollingUpdateScaleOut`, `RollingUpdateScaleIn` and `OverrideScaling`. If `OverrideScaling` is used, `override_scaling` must be specified with both `max_surge` and `max_unavailable`.",
							ValidateFunc: validation.StringInSlice([]string{"RollingUpdateScaleOut", "RollingUpdateScaleIn", "OverrideScaling"}, false),
						},
						"override_scaling": schemas.OverrideScalingSchema(),
						"override_kubeadm_configuration": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "YAML config for kubeletExtraArgs, preKubeadmCommands, postKubeadmCommands. Overrides pack-level settings. Worker pools only.",
						},
						"instance_type": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"disk_size_gb": {
										Type:        schema.TypeInt,
										Required:    true,
										Description: "The size of the disk in GB.",
									},
									"memory_mb": {
										Type:        schema.TypeInt,
										Required:    true,
										Description: "The amount of memory in MB.",
									},
									"cpu": {
										Type:        schema.TypeInt,
										Required:    true,
										Description: "The number of CPUs.",
									},
								},
							},
						},
						"placement": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"cluster": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The name of the cluster to use for the machine pool. As it appears in the vSphere.",
									},
									"resource_pool": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The name of the resource pool to use for the machine pool. As it appears in the vSphere.",
									},
									"datastore": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The name of the datastore to use for the machine pool. As it appears in the vSphere.",
									},
									"network": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The name of the network to use for the machine pool. As it appears in the vSphere.",
									},
									"static_ip_pool_id": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "The ID of the static IP pool to use for the machine pool in case of static cluster placement.",
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
			"location_config":      schemas.ClusterLocationSchema(),
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

// resourceClusterMaasResourceV2 returns the schema for version 2 of the MAAS cluster resource.
// Version 2 used TypeList for cluster_profile.
func resourceClusterMaasResourceV2() *schema.Resource {
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
				Description: "The context of the MAAS configuration. Allowed values are `project` or `tenant`. " +
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
			"cluster_type":     schemas.ClusterTypeSchema(),
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
				Optional:    true,
				ForceNew:    true,
				Description: "ID of the Maas cloud account used for the cluster. This cloud account must be of type `maas`.",
			},
			"cloud_config_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the cloud config used for the cluster. This cloud config must be of type `maas`.",
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
			"cluster_timezone": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "",
				ValidateFunc: validateTimezone,
				Description:  "Defines the time zone used by this cluster to interpret scheduled operations. Maintenance tasks like upgrades will follow this time zone to ensure they run at the appropriate local time for the cluster. Must be in IANA timezone format (e.g., 'America/New_York', 'Asia/Kolkata', 'Europe/London').",
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
							Description: "Domain name in which the cluster to be provisioned.",
						},
						"enable_lxd_vm": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Whether to enable LXD VM. Default is `false`.",
						},
						"ntp_servers": {
							Type:     schema.TypeSet,
							Optional: true,
							Set:      schema.HashString,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "A list of NTP servers to use instead of the machine image's default NTP server list.",
						},
					},
				},
			},
			"machine_pool": {
				Type:     schema.TypeSet,
				Required: true,
				Set:      resourceMachinePoolMaasHash,
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
						"additional_annotations": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "Additional annotations to be applied to the machine pool. Annotations must be in the form of `key:value`.",
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
						"node_repave_interval": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     0,
							Description: "Minimum number of seconds node should be Ready, before the next node is selected for repave. Default value is `0`, Applicable only for worker pools.",
						},
						"min": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     0,
							Description: "Minimum number of nodes in the machine pool. This is used for autoscaling the machine pool.",
						},
						"max": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     0,
							Description: "Maximum number of nodes in the machine pool. This is used for autoscaling the machine pool.",
						},
						"instance_type": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"min_memory_mb": {
										Type:        schema.TypeInt,
										Required:    true,
										Description: "Minimum memory in MB required for the machine pool node.",
									},
									"min_cpu": {
										Type:        schema.TypeInt,
										Required:    true,
										Description: "Minimum number of CPU required for the machine pool node.",
									},
								},
							},
						},
						"update_strategy": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "RollingUpdateScaleOut",
							Description:  "Update strategy for the machine pool. Valid values are `RollingUpdateScaleOut`, `RollingUpdateScaleIn` and `OverrideScaling`. If `OverrideScaling` is used, `override_scaling` must be specified with both `max_surge` and `max_unavailable`.",
							ValidateFunc: validation.StringInSlice([]string{"RollingUpdateScaleOut", "RollingUpdateScaleIn", "OverrideScaling"}, false),
						},
						"override_scaling": schemas.OverrideScalingSchema(),
						"override_kubeadm_configuration": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "YAML config for kubeletExtraArgs, preKubeadmCommands, postKubeadmCommands. Overrides pack-level settings. Worker pools only.",
						},
						"azs": {
							Type:     schema.TypeSet,
							Required: true,
							Set:      schema.HashString,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "Availability zones in which the machine pool nodes to be provisioned.",
						},
						"node_tags": {
							Type:     schema.TypeSet,
							Optional: true,
							Set:      schema.HashString,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "Node tags to dynamically place nodes in a pool by using MAAS automatic tags. Specify the tag values that you want to apply to all nodes in the node pool.",
						},
						"placement": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "This is a computed(read-only) ID of the placement that is used to connect to the Maas cloud.",
									},
									"resource_pool": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The name of the resource pool in the Maas cloud.",
									},
								},
							},
						},
						"use_lxd_vm": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Whether to use LXD VM. Default is `false`. Available once **Palette with LXD support** is released.",
						},
						"network": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Network configuration for the machine pool. Available once **Palette with LXD support** is released.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"network_name": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The name of the network in which VMs are created/located.",
									},
									"parent_pool_uid": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "The UID of the parent pool which allocates IPs for this IPPool.",
									},
									"static_ip": {
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     false,
										Description: "Whether to use static IP. Default is `false`.",
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
			"location_config":      schemas.ClusterLocationSchema(),
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

// resourceClusterApacheCloudStackResourceV2 returns the schema for version 2 of the Apache CloudStack cluster resource.
// Version 2 used TypeList for cluster_profile.
func resourceClusterApacheCloudStackResourceV2() *schema.Resource {
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
			"cluster_timezone": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "",
				ValidateFunc: validateTimezone,
				Description:  "Defines the time zone used by this cluster to interpret scheduled operations. Maintenance tasks like upgrades will follow this time zone to ensure they run at the appropriate local time for the cluster. Must be in IANA timezone format (e.g., 'America/New_York', 'Asia/Kolkata', 'Europe/London').",
			},
			"update_worker_pools_in_parallel": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Controls whether worker pool updates occur in parallel or sequentially. When set to `true` (default), all worker pools are updated simultaneously. When `false`, worker pools are updated one at a time, reducing cluster disruption but taking longer to complete updates.",
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
						"additional_annotations": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "Additional annotation to be applied to the machine pool. annotation must be in the form of `key:value`.",
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
						"node_repave_interval": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     0,
							Description: "Minimum number of seconds node should be Ready, before the next node is selected for repave. Default value is `0`, Applicable only for worker pools.",
						},
						"update_strategy": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "RollingUpdateScaleOut",
							Description:  "Update strategy for the machine pool. Valid values are `RollingUpdateScaleOut`, `RollingUpdateScaleIn` and `OverrideScaling`. If `OverrideScaling` is used, `override_scaling` must be specified with both `max_surge` and `max_unavailable`.",
							ValidateFunc: validation.StringInSlice([]string{"RollingUpdateScaleOut", "RollingUpdateScaleIn", "OverrideScaling"}, false),
						},
						"override_scaling": schemas.OverrideScalingSchema(),
						"override_kubeadm_configuration": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "YAML config for kubeletExtraArgs, preKubeadmCommands, postKubeadmCommands. Overrides pack-level settings. Worker pools only.",
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
							Description: "Apache CloudStack template override for this machine pool. If not specified, inherits cluster default.",
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

// resourceClusterBrownfieldResourceV1 returns the schema for version 1 of the brownfield cluster resource.
// Version 1 used TypeList for cluster_profile.
func resourceClusterBrownfieldResourceV1() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the cluster to be registered. This field cannot be updated after creation.",
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
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The description of the cluster. Default value is empty string.",
			},
			"cloud_type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"aws",
					"eks-anywhere",
					"azure",
					"gcp",
					"vsphere",
					"openshift",
					"generic",
					"apache-cloudstack",
					"edge-native",
					"maas",
					"openstack",
				}, false),
				Description: "The cloud type of the cluster. Supported values: `aws`, `eks-anywhere`, `azure`, `gcp`, `vsphere`, `openshift`, `generic`,`apache-cloudstack`,`edge-native`,`maas`,`openstack`. This field cannot be updated after creation.",
			},
			"import_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "",
				ValidateFunc: validation.StringInSlice([]string{"read_only", "full", ""}, false),
				Description:  "The import mode for the cluster. Allowed values are `read_only` (imports cluster with read-only permissions) or `full` (imports cluster with full permissions). Defaults to `full`. This field cannot be updated after creation.",
			},
			"host_path": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Location for Proxy CA cert on host nodes. This is the file path on the host where the Proxy CA certificate is stored. This field cannot be updated after creation.",
			},
			"container_mount_path": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Location to mount Proxy CA cert inside container. This is the file path inside the container where the Proxy CA certificate will be mounted. This field cannot be updated after creation.",
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"project", "tenant"}, false),
				Description:  "The context for the cluster registration. Allowed values are `project` or `tenant`. Defaults to `project`. This field cannot be updated after creation." + PROJECT_NAME_NUANCE,
			},
			"proxy": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Location to mount Proxy CA cert inside container. This field supports vsphere and openshift clusters. This field cannot be updated after creation.",
			},
			"no_proxy": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Location to mount Proxy CA cert inside container. This field supports vsphere and openshift clusters. This field cannot be updated after creation.",
			},
			"manifest_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The URL of the import manifest that must be applied to your Kubernetes cluster to complete the import into Palette.",
			},
			"kubectl_command": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The kubectl command that must be executed on your Kubernetes cluster to complete the import process into Palette.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The current operational state of the cluster. Possible values include: `Pending`, `Provisioning`, `Running`, `Deleting`, `Deleted`, `Error`, `Importing`.",
			},
			"health_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The current health status of the cluster. Possible values include: `Healthy`, `UnHealthy`, `Unknown`.",
			},
			"cloud_config_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the cloud config used for the cluster. This is automatically set from the cluster's cloud config reference.",
			},
			"cluster_timezone": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "",
				ValidateFunc: validateTimezone,
				Description:  "Defines the time zone used by this cluster to interpret scheduled operations. Maintenance tasks like upgrades will follow this time zone to ensure they run at the appropriate local time for the cluster. Must be in IANA timezone format (e.g., 'America/New_York', 'Asia/Kolkata', 'Europe/London').",
			},
			"apply_setting": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "DownloadAndInstall",
				ValidateFunc: validation.StringInSlice([]string{"DownloadAndInstall", "DownloadAndInstallLater"}, false),
				Description: "The setting to apply the cluster profile. `DownloadAndInstall` will download and install packs in one action. " +
					"`DownloadAndInstallLater` will only download artifact and postpone install for later. " +
					"Default value is `DownloadAndInstall`.",
			},
			"cluster_profile": schemas.ClusterProfileSchema(),
			"machine_pool": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The name of the machine pool.",
						},
						"node": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"node_name": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "The name of the machine pool.",
									},
									"node_id": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "The node_id of the node, For example `i-07f899a33dee624f7`",
									},
									"action": {
										Type:         schema.TypeString,
										Required:     true,
										Description:  "The action to perform on the node. Valid values are: `cordon`, `uncordon`.",
										ValidateFunc: validation.StringInSlice([]string{"cordon", "uncordon"}, false),
									},
								},
							},
						},
					},
				},
				Description: "Machine pool configuration for Day-2 node maintenance operations. Used to perform node actions like cordon/uncordon on specific nodes.",
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

// resourceClusterGkeResourceV2 returns the schema for version 2 of the GKE cluster resource.
// Version 2 used TypeList for cluster_profile.
func resourceClusterGkeResourceV2() *schema.Resource {
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
				Description: "The context of the GKE cluster. Allowed values are `project` or `tenant`. " +
					"Default is `project`. " + PROJECT_NAME_NUANCE,
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The description of the cluster. Default value is empty string.",
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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cloud_config_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the cloud config used for the cluster. This cloud config must be of type `azure`.",
				Deprecated:  "This field is deprecated and will be removed in the future. Use `cloud_config` instead.",
			},
			"cloud_config": {
				Type:        schema.TypeList,
				ForceNew:    true,
				Required:    true,
				MaxItems:    1,
				Description: "The GKE environment configuration settings such as project parameters and region parameters that apply to this cluster.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"project": {
							Type:        schema.TypeString,
							ForceNew:    true,
							Required:    true,
							Description: "GCP project name.",
						},
						"region": {
							Type:     schema.TypeString,
							ForceNew: true,
							Required: true,
						},
					},
				},
			},
			"update_worker_pool_in_parallel": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"cluster_timezone": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "",
				ValidateFunc: validateTimezone,
				Description:  "Defines the time zone used by this cluster to interpret scheduled operations. Maintenance tasks like upgrades will follow this time zone to ensure they run at the appropriate local time for the cluster. Must be in IANA timezone format (e.g., 'America/New_York', 'Asia/Kolkata', 'Europe/London').",
			},
			"machine_pool": {
				Type:        schema.TypeSet,
				Required:    true,
				Set:         resourceMachinePoolGkeHash,
				Description: "The machine pool configuration for the cluster.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"count": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Number of nodes in the machine pool.",
						},
						"disk_size_gb": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  60,
						},
						"additional_labels": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "Additional labels to be applied to the machine pool. Labels must be in the form of `key:value`.",
						},
						"additional_annotations": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "Additional annotations to be applied to the machine pool. Annotations must be in the form of `key:value`.",
						},
						"instance_type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"update_strategy": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "RollingUpdateScaleOut",
							Description:  "Update strategy for the machine pool. Valid values are `RollingUpdateScaleOut`, `RollingUpdateScaleIn` and `OverrideScaling`. If `OverrideScaling` is used, `override_scaling` must be specified with both `max_surge` and `max_unavailable`.",
							ValidateFunc: validation.StringInSlice([]string{"RollingUpdateScaleOut", "RollingUpdateScaleIn", "OverrideScaling"}, false),
						},
						"override_scaling": schemas.OverrideScalingSchema(),
						"override_kubeadm_configuration": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "YAML config for kubeletExtraArgs, preKubeadmCommands, postKubeadmCommands. Overrides pack-level settings. Worker pools only.",
						},
						"node":   schemas.NodeSchema(),
						"taints": schemas.ClusterTaintsSchema(),
					},
				},
			},
			"pause_agent_upgrades": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "unlock",
				ValidateFunc: validation.StringInSlice([]string{"lock", "unlock"}, false),
				Description:  "The pause agent upgrades setting allows to control the automatic upgrade of the Palette component and agent for an individual cluster. The default value is `unlock`, meaning upgrades occur automatically. Setting it to `lock` pauses automatic agent upgrades for the cluster.",
			},
			"review_repave_state": {
				Type:         schema.TypeString,
				Default:      "",
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"", "Approved", "Pending"}, false),
				Description:  "To authorize the cluster repave, set the value to `Approved` for approval and `\"\"` to decline. Default value is `\"\"`.",
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

// resourceClusterEdgeNativeResourceV3 returns the schema for version 3 of the Edge Native cluster resource.
// Version 3 used TypeList for cluster_profile.
func resourceClusterEdgeNativeResourceV3() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"", "project", "tenant"}, false),
				Description: "The context of the Edge cluster. Allowed values are `project` or `tenant`. " +
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
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
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
				Description:      "The cron schedule for OS patching. This must be in the form of cron syntax. Ex: `0 0 * * *`.",
			},
			"os_patch_after": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateOsPatchOnDemandAfter,
				Description:      "Date and time after which to patch cluster `RFC3339: 2006-01-02T15:04:05Z07:00`",
			},
			"cluster_timezone": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "",
				ValidateFunc: validateTimezone,
				Description:  "Defines the time zone used by this cluster to interpret scheduled operations. Maintenance tasks like upgrades will follow this time zone to ensure they run at the appropriate local time for the cluster. Must be in IANA timezone format (e.g., 'America/New_York', 'Asia/Kolkata', 'Europe/London').",
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
				ForceNew: true,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ssh_keys": {
							Type:     schema.TypeSet,
							Optional: true,
							Set:      schema.HashString,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "List of public SSH (Secure Shell) to establish, administer, and communicate with remote clusters.",
						},
						"vip": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The `vip` can be specified as either an IP address or a fully qualified domain name (FQDN). If `overlay_cidr_range` is set, the `vip` should be within the specified `overlay_cidr_range`. By default, the `vip` is set to the first IP address within the given `overlay_cidr_range`.",
						},
						"overlay_cidr_range": {
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
							Description: "The Overlay (VPN) creates a virtual network, using techniques like VxLAN. It overlays the existing network infrastructure, enhancing connectivity either at Layer 2 or Layer 3, making it flexible and adaptable for various needs. For example, `100.64.192.0/24`",
						},
						"is_two_node_cluster": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Set to `true` to enable a two-node cluster.",
						},
						"ntp_servers": {
							Type:     schema.TypeSet,
							Optional: true,
							Set:      schema.HashString,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "A list of NTP servers to be used by the cluster.",
						},
					},
				},
			},
			"machine_pool": {
				Type:     schema.TypeSet,
				Required: true,
				Set:      resourceMachinePoolEdgeNativeHash,
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
							Description: "Additional labels to be applied to the machine pool. Labels must be in the form of `key:value`.",
						},
						"additional_annotations": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "Additional annotations to be applied to the machine pool. Annotations must be in the form of `key:value`.",
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
						"node_repave_interval": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     0,
							Description: "Minimum number of seconds node should be Ready, before the next node is selected for repave. Default value is `0`, Applicable only for worker pools.",
						},
						"update_strategy": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "RollingUpdateScaleOut",
							Description:  "Update strategy for the machine pool. Valid values are `RollingUpdateScaleOut`, `RollingUpdateScaleIn` and `OverrideScaling`. If `OverrideScaling` is used, `override_scaling` must be specified with both `max_surge` and `max_unavailable`.",
							ValidateFunc: validation.StringInSlice([]string{"RollingUpdateScaleOut", "RollingUpdateScaleIn", "OverrideScaling"}, false),
						},
						"override_scaling": schemas.OverrideScalingSchema(),
						"override_kubeadm_configuration": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "YAML config for kubeletExtraArgs, preKubeadmCommands, postKubeadmCommands. Overrides pack-level settings. Worker pools only.",
						},
						"edge_host": {
							Type:     schema.TypeSet,
							Required: true,
							Set:      resourceEdgeHostHash,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"host_name": {
										Type:        schema.TypeString,
										Optional:    true,
										Default:     "",
										Description: "Edge host name",
									},
									"host_uid": {
										Type:        schema.TypeString,
										Description: "Edge host id",
										Required:    true,
									},
									"static_ip": {
										Type:        schema.TypeString,
										Description: "Edge host static IP address",
										Optional:    true,
									},
									"nic_name": {
										Type:        schema.TypeString,
										Description: "NIC Name for edge host.",
										Optional:    true,
									},
									"default_gateway": {
										Type:        schema.TypeString,
										Description: "Edge host default gateway",
										Optional:    true,
									},
									"subnet_mask": {
										Type:        schema.TypeString,
										Description: "Edge host subnet mask",
										Optional:    true,
									},
									"dns_servers": {
										Type:        schema.TypeSet,
										Optional:    true,
										Set:         schema.HashString,
										Description: "Edge host DNS servers",
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"two_node_role": {
										Type:         schema.TypeString,
										Description:  "Two node role for edge host. Valid values are `primary` and `secondary`.",
										Optional:     true,
										ValidateFunc: validation.StringInSlice([]string{"primary", "secondary"}, false),
									},
								},
							},
						},
					},
				},
			},
			"pause_agent_upgrades": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "unlock",
				ValidateFunc: validation.StringInSlice([]string{"lock", "unlock"}, false),
				Description:  "The pause agent upgrades setting allows to control the automatic upgrade of the Palette component and agent for an individual cluster. The default value is `unlock`, meaning upgrades occur automatically. Setting it to `lock` pauses automatic agent upgrades for the cluster.",
			},
			"backup_policy":        schemas.BackupPolicySchema(),
			"scan_policy":          schemas.ScanPolicySchema(),
			"cluster_rbac_binding": schemas.ClusterRbacBindingSchema(),
			"namespaces":           schemas.ClusterNamespacesSchema(),
			"host_config":          schemas.ClusterHostConfigSchema(),
			"location_config":      schemas.ClusterLocationSchema(),
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

// resourceClusterEdgeVsphereResourceV0 returns the schema for version 0 of the Edge vSphere cluster resource.
// Version 0 used TypeList for cluster_profile.
func resourceClusterEdgeVsphereResourceV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"", "project", "tenant"}, false),
				Description:  "The context of the Edge cluster. Allowed values are `project` or `tenant`. " + "Default is `project`. " + PROJECT_NAME_NUANCE,
			},
			"edge_host_uid": {
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
			"cluster_timezone": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "",
				ValidateFunc: validateTimezone,
				Description:  "Defines the time zone used by this cluster to interpret scheduled operations. Maintenance tasks like upgrades will follow this time zone to ensure they run at the appropriate local time for the cluster. Must be in IANA timezone format (e.g., 'America/New_York', 'Asia/Kolkata', 'Europe/London').",
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
				ForceNew: true,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"datacenter": {
							Type:     schema.TypeString,
							Required: true,
						},
						"folder": {
							Type:     schema.TypeString,
							Required: true,
						},
						"image_template_folder": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"ssh_key": {
							Type:         schema.TypeString,
							Optional:     true,
							ExactlyOneOf: []string{"cloud_config.0.ssh_key", "cloud_config.0.ssh_keys"},
							Description:  "Public SSH Key (Secure Shell) to establish, administer, and communicate with remote clusters, `ssh_key & ssh_keys` are mutually exclusive.",
						},
						"ssh_keys": {
							Type:         schema.TypeSet,
							Optional:     true,
							Set:          schema.HashString,
							ExactlyOneOf: []string{"cloud_config.0.ssh_key", "cloud_config.0.ssh_keys"},
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "List of public SSH (Secure Shell) keys to establish, administer, and communicate with remote clusters, `ssh_key & ssh_keys` are mutually exclusive.",
						},
						"vip": {
							Type:     schema.TypeString,
							Required: true,
						},
						"static_ip": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"network_type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"network_search_domain": {
							Type:     schema.TypeString,
							Optional: true,
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
							Description: "Additional labels to be applied to the machine pool. Labels must be in the form of `key:value`.",
						},
						"additional_annotations": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "Additional annotations to be applied to the machine pool. Annotations must be in the form of `key:value`.",
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
						"count": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Number of nodes in the machine pool.",
						},
						"node_repave_interval": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     0,
							Description: "Minimum number of seconds node should be Ready, before the next node is selected for repave. Default value is `0`, Applicable only for worker pools.",
						},
						"update_strategy": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "RollingUpdateScaleOut",
							Description:  "Update strategy for the machine pool. Valid values are `RollingUpdateScaleOut`, `RollingUpdateScaleIn` and `OverrideScaling`. If `OverrideScaling` is used, `override_scaling` must be specified with both `max_surge` and `max_unavailable`.",
							ValidateFunc: validation.StringInSlice([]string{"RollingUpdateScaleOut", "RollingUpdateScaleIn", "OverrideScaling"}, false),
						},
						"override_scaling": schemas.OverrideScalingSchema(),
						"override_kubeadm_configuration": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "YAML config for kubeletExtraArgs, preKubeadmCommands, postKubeadmCommands. Overrides pack-level settings. Worker pools only.",
						},
						"instance_type": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
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
						"placement": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"cluster": {
										Type:     schema.TypeString,
										Required: true,
									},
									"resource_pool": {
										Type:     schema.TypeString,
										Required: true,
									},
									"datastore": {
										Type:     schema.TypeString,
										Required: true,
									},
									"network": {
										Type:     schema.TypeString,
										Required: true,
									},
									"static_ip_pool_id": {
										Type:     schema.TypeString,
										Optional: true,
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
			"location_config":      schemas.ClusterLocationSchema(),
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

// resourceClusterCustomCloudResourceV3 returns the schema for version 3 of the Custom Cloud cluster resource.
// Version 3 used TypeList for cluster_profile.
func resourceClusterCustomCloudResourceV3() *schema.Resource {
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
			"cloud": {
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
				Description: "The cloud provider name.",
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
			"cluster_profile":  schemas.ClusterProfileSchema(),
			"cluster_template": schemas.ClusterTemplateSchema(),
			"cluster_type":     schemas.ClusterTypeSchema(),
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
				Description: "The cloud account id to use for this cluster.",
			},
			"cloud_config_id": {
				Type:       schema.TypeString,
				Computed:   true,
				Deprecated: "This field is deprecated and will be removed in the future. Use `cloud_config` instead.",
			},
			"cluster_timezone": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "",
				ValidateFunc: validateTimezone,
				Description:  "Defines the time zone used by this cluster to interpret scheduled operations. Maintenance tasks like upgrades will follow this time zone to ensure they run at the appropriate local time for the cluster. Must be in IANA timezone format (e.g., 'America/New_York', 'Asia/Kolkata', 'Europe/London').",
			},
			"cloud_config": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Description: "The Cloud environment configuration settings such as network parameters and encryption parameters that apply to this cluster.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"values": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The values of the cloud config. The values are specified in YAML format. ",
							StateFunc: func(val interface{}) string {
								// Normalize YAML content to handle formatting differences
								if yamlStr, ok := val.(string); ok {
									return NormalizeYamlContent(yamlStr)
								}
								return val.(string)
							},
						},
						"overrides": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Description: "Key-value pairs to override specific values in the YAML. Supports template variables, wildcard patterns, field pattern search, document-specific and global overrides.\n\n" +
								"Template variables: Simple identifiers that replace ${var}, {{var}}, or $var patterns in YAML (e.g., 'cluster_name' replaces ${cluster_name})\n" +
								"Wildcard patterns: Patterns starting with '*' that match field names containing the specified substring (e.g., '*cluster-api-autoscaler-node-group-max-size' matches any field containing 'cluster-api-autoscaler-node-group-max-size')\n" +
								"Field pattern search: Patterns that find and update ALL matching nested fields anywhere in YAML (e.g., 'replicas' updates any 'replicas' field, 'rootVolume.size' updates any 'rootVolume.size' pattern)\n" +
								"Document-specific syntax: 'Kind.path' (e.g., 'Cluster.metadata.labels', 'AWSCluster.spec.region')\n" +
								"Global path syntax: 'path' (e.g., 'metadata.name', 'spec.region')\n\n" +
								"Processing order: 1) Template substitution, 2) Wildcard patterns, 3) Field pattern search, 4) Path-based overrides. " +
								"Supports dot notation for nested paths and array indexing with [index]. " +
								"Values are strings but support JSON syntax for arrays/objects.",
						},
					},
				},
			},
			"machine_pool": {
				Type:        schema.TypeSet,
				Required:    true,
				Set:         resourceMachinePoolCustomCloudHash,
				Description: "The machine pool configuration for the cluster.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the machine pool. This will be derived from the name value in the `node_pool_config`.",
						},
						"count": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of nodes in the machine pool. This will be derived from the replica value in the 'node_pool_config'.",
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
						"node_pool_config": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The values of the node pool config. The values are specified in YAML format. ",
							StateFunc: func(val interface{}) string {
								// Normalize YAML content to handle formatting differences
								if yamlStr, ok := val.(string); ok {
									return NormalizeYamlContent(yamlStr)
								}
								return val.(string)
							},
						},
						"overrides": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Description: "Key-value pairs to override specific values in the node pool config YAML. Supports template variables, wildcard patterns, field pattern search, document-specific and global overrides.\n\n" +
								"Template variables: Simple identifiers that replace ${var}, {{var}}, or $var patterns in YAML (e.g., 'node_count' replaces ${node_count})\n" +
								"Wildcard patterns: Patterns starting with '*' that match field names containing the specified substring (e.g., '*cluster-api-autoscaler-node-group-max-size' matches any field containing 'cluster-api-autoscaler-node-group-max-size')\n" +
								"Field pattern search: Patterns that find and update ALL matching nested fields anywhere in YAML (e.g., 'replicas' updates any 'replicas' field, 'rootVolume.size' updates any 'rootVolume.size' pattern)\n" +
								"Document-specific syntax: 'Kind.path' (e.g., 'AWSMachineTemplate.spec.template.spec.instanceType')\n" +
								"Global path syntax: 'path' (e.g., 'metadata.name', 'spec.instanceType')\n\n" +
								"Processing order: 1) Template substitution, 2) Wildcard patterns, 3) Field pattern search, 4) Path-based overrides. " +
								"Supports dot notation for nested paths and array indexing with [index]. " +
								"Values are strings but support JSON syntax for arrays/objects.",
						},
					},
				},
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
				Description:      "The cron schedule for OS patching. This must be in the form of cron syntax. Ex: `0 0 * * *`.",
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
			"backup_policy":        schemas.BackupPolicySchema(),
			"scan_policy":          schemas.ScanPolicySchema(),
			"cluster_rbac_binding": schemas.ClusterRbacBindingSchema(),
			"namespaces":           schemas.ClusterNamespacesSchema(),
			"location_config":      schemas.ClusterLocationSchema(),
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

// resourceClusterGroupResourceV2 returns the schema for version 2 of the cluster group resource.
// Version 2 used TypeList for cluster_profile.
func resourceClusterGroupResourceV2() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the cluster group",
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "tenant",
				ValidateFunc: validation.StringInSlice([]string{"", "project", "tenant"}, false),
				Description:  "The context of the Cluster group. Allowed values are `project` or `tenant`. " + "Defaults to `tenant`. " + PROJECT_NAME_NUANCE,
			},
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      schema.HashString,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "A list of tags to be applied to the cluster group. Tags must be in the form of `key:value`.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The description of the cluster. Default value is empty string.",
			},
			"config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"host_endpoint_type": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "Ingress",
							ValidateFunc: validation.StringInSlice([]string{"Ingress", "LoadBalancer"}, false),
							Description:  "The host endpoint type. Allowed values are 'Ingress' or 'LoadBalancer'. Defaults to 'Ingress'.",
						},
						"cpu_millicore": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "The CPU limit in millicores.",
						},
						"memory_in_mb": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "The memory limit in megabytes (MB).",
						},
						"storage_in_gb": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "The storage limit in gigabytes (GB).",
						},
						"oversubscription_percent": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "The allowed oversubscription percentage.",
						},
						"values": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},
						"k8s_distribution": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "vcluster-generic",
							ForceNew:    true,
							Description: "The Kubernetes distribution, allowed values are `vcluster-generic`,`k3s` and `cncf_k8s`.",
						},
					},
				},
			},
			"cluster_profile": schemas.ClusterProfileSchema(),
			"clusters": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "A list of clusters to include in the cluster group.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cluster_uid": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The UID of the host cluster.",
						},
						"host_dns": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The host DNS wildcard for the cluster. i.e. `*.dev` or `*test.com`",
						},
					},
				},
			},
		},
	}
}

// resourceClusterVirtualResourceV2 returns the schema for version 2 of the virtual cluster resource.
// Version 2 used TypeList for cluster_profile.
func resourceClusterVirtualResourceV2() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"project", "cluster"}, false),
				Description:  "The context of the virtual cluster. Allowed values are `project` or `tenant`. " + "Default is `project`. " + PROJECT_NAME_NUANCE,
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
			"host_cluster_uid": {
				Type:         schema.TypeString,
				Optional:     true,
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
				MaxItems: 1,
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
			"pause_agent_upgrades": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "unlock",
				ValidateFunc: validation.StringInSlice([]string{"lock", "unlock"}, false),
				Description:  "The pause agent upgrades setting allows to control the automatic upgrade of the Palette component and agent for an individual cluster. The default value is `unlock`, meaning upgrades occur automatically. Setting it to `lock` pauses automatic agent upgrades for the cluster.",
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
			"admin_kube_config": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Admin Kube-config for the cluster. This can be used to connect to the cluster using `kubectl`, With admin privilege.",
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

// resourceAddonDeploymentResourceV2 returns the schema for version 2 of the addon deployment resource.
// Version 2 used TypeList for cluster_profile.
func resourceAddonDeploymentResourceV2() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"cluster_uid": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The UID of the cluster to attach the addon profile to.",
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"project", "tenant"}, false),
				Default:      "project",
				Description: "Specifies cluster context where addon profile is attached. " +
					"Allowed values are `project` or `tenant`. Defaults to `project`. " + PROJECT_NAME_NUANCE,
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
		},
	}
}

package spectrocloud

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/schemas"
)

func resourceClusterBrownfield() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterBrownfieldImportCreate,
		ReadContext:   resourceClusterBrownfieldRead,
		UpdateContext: resourceClusterBrownfieldUpdate,
		DeleteContext: resourceClusterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceClusterBrownfieldImport,
		},
		SchemaVersion: 1,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Description: "Register an existing Kubernetes cluster (brownfield) with Palette. This resource allows you to import and manage existing Kubernetes clusters. Supported cloud platforms: (AWS, Azure, GCP, vSphere, OpenShift, Generic, Apache CloudStack, Edge Native, MAAS, and OpenStack). This feature is currently in preview.",
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
				Default:      "full",
				ValidateFunc: validation.StringInSlice([]string{"read_only", "full"}, false),
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

func resourceClusterBrownfieldImportCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)
	var diags diag.Diagnostics

	cloudType := d.Get("cloud_type").(string)
	// name := d.Get("name").(string)

	// Build metadata
	metadata := toBrownfieldClusterMetadata(d)

	// Register the cluster based on cloud type
	var clusterUID string
	var err error

	switch cloudType {
	case "aws":
		entity := &models.V1SpectroAwsClusterImportEntity{
			Metadata: metadata,
			Spec:     toBrownfieldClusterSpecAws(d),
		}
		clusterUID, err = c.ImportSpectroClusterAws(entity)
	case "azure":
		entity := &models.V1SpectroAzureClusterImportEntity{
			Metadata: metadata,
			Spec:     toBrownfieldClusterSpecAzure(d),
		}
		clusterUID, err = c.ImportSpectroClusterAzure(entity)
	case "gcp":
		entity := &models.V1SpectroGcpClusterImportEntity{
			Metadata: metadata,
			Spec:     toBrownfieldClusterSpecGcp(d),
		}
		clusterUID, err = c.ImportSpectroClusterGcp(entity)
	case "vsphere", "openshift":
		entity := &models.V1SpectroVsphereClusterImportEntity{
			Metadata: metadata,
			Spec:     toBrownfieldClusterSpecVsphere(d),
		}
		clusterUID, err = c.ImportSpectroVsphereCluster(entity)
	case "generic", "eks-anywhere":
		entity := &models.V1SpectroGenericClusterImportEntity{
			Metadata: metadata,
			Spec:     toBrownfieldClusterSpecGeneric(d),
		}
		clusterUID, err = c.ImportSpectroClusterGeneric(entity)
	case "apache-cloudstack":
		entity := &models.V1SpectroCloudStackClusterImportEntity{
			Metadata: metadata,
			Spec:     toBrownfieldClusterSpecCloudStack(d),
		}
		clusterUID, err = c.ImportSpectroClusterApacheCloudStack(entity)
	case "maas":
		entity := &models.V1SpectroMaasClusterImportEntity{
			Metadata: metadata,
			Spec:     toBrownfieldClusterSpecMaas(d),
		}
		clusterUID, err = c.ImportSpectroClusterMaas(entity)
	case "edge-native":
		entity := &models.V1SpectroEdgeNativeClusterImportEntity{
			Metadata: metadata,
			Spec:     toBrownfieldClusterSpecEdgeNative(d),
		}
		clusterUID, err = c.ImportSpectroClusterEdgeNative(entity)

	default:
		return diag.FromErr(fmt.Errorf("unsupported cloud type: %s", cloudType))
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to register brownfield cluster: %w", err))
	}

	// Set the cluster UID as the resource ID
	d.SetId(clusterUID)

	// Wait 3 seconds for cluster to be initialized before fetching details
	time.Sleep(5 * time.Second)

	cluster, err := c.GetCluster(clusterUID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to get cluster: %w", err))
	}
	if cluster == nil {
		return diag.FromErr(fmt.Errorf("cluster not found"))
	}

	// Get the import link and manifest URL from cluster object
	kubectlCommand, manifestURL, err := getClusterImportInfo(cluster)

	if err != nil {
		// Log warning but don't fail - import link may not be available immediately
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Import link not immediately available",
			Detail:   fmt.Sprintf("Cluster registered successfully, but import link is not yet available: %v. You may need to run 'terraform refresh' to get the import link.", err),
		})
	} else {
		if err := d.Set("kubectl_command", kubectlCommand); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("manifest_url", manifestURL); err != nil {
			return diag.FromErr(err)
		}
		// Show warning message about applying manifest
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Cluster import submitted",
			Detail:   "Cluster import is submitted. Please apply the manifest using `manifest_url` and run `kubectl_command` on your cluster to start the import process. Once it becomes Running and Healthy, Day-2 operations will be allowed.",
		})
	}
	updateCommonFieldsForBrownfieldCluster(d, c)

	resourceClusterBrownfieldRead(ctx, d, m)
	return diags
}

// Read function - reads the current state of the cluster
func resourceClusterBrownfieldRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)
	var diags diag.Diagnostics
	clusterUID := d.Id()

	// Get the cluster to verify it exists
	cluster, err := c.GetCluster(clusterUID)
	if err != nil {
		return handleReadError(d, err, diags)
	}
	if cluster == nil {
		// Cluster has been deleted
		d.SetId("")
		return diags
	}

	// Always update computed fields
	// Status - always update from API
	if cluster.Status != nil && cluster.Status.State != "" {
		if err := d.Set("status", cluster.Status.State); err != nil {
			return diag.FromErr(err)
		}
	} else {
		// Set empty string if status is not available
		if err := d.Set("status", ""); err != nil {
			return diag.FromErr(err)
		}
	}

	// Read common fields (wrapped to skip fields not in schema)
	readDiags, hasError := readCommonFieldsBrownfield(c, d, cluster)
	if hasError {
		diags = append(diags, readDiags...)
		return diags
	}
	diags = append(diags, readDiags...)

	// Get the import link and manifest URL from cluster object
	kubectlCommand, manifestURL, err := getClusterImportInfo(cluster)
	if err != nil {
		// Import link may not be available - show warning
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "kubectl_command not available",
			Detail:   fmt.Sprintf("kubectl_command is not yet available for cluster %s: %v", clusterUID, err),
		})
		// Set empty strings for computed fields when not available
		if err := d.Set("kubectl_command", ""); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("manifest_url", ""); err != nil {
			return diag.FromErr(err)
		}
	} else {
		// Set computed fields from API
		if err := d.Set("kubectl_command", kubectlCommand); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("manifest_url", manifestURL); err != nil {
			return diag.FromErr(err)
		}

		// Show warning if cluster is not Running and kubectl_command is available
		if cluster.Status != nil && cluster.Status.State != "Running" {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "Cluster import pending",
				Detail:   "Cluster import is submitted. Please apply the manifest using `manifest_url` and run `kubectl_command` on your cluster to start the import process. Once it becomes Running and Healthy, Day-2 operations will be allowed.",
			})
		}
	}

	// Set cloud_config_id from cluster spec
	if cluster.Spec != nil && cluster.Spec.CloudConfigRef != nil && cluster.Spec.CloudConfigRef.UID != "" {
		if err := d.Set("cloud_config_id", cluster.Spec.CloudConfigRef.UID); err != nil {
			return diag.FromErr(err)
		}
	}

	// Set health_status from cluster overview
	clusterSummary, err := c.GetClusterOverview(clusterUID)
	if err != nil {
		// If we can't get overview, set to "Unknown"
		if err := d.Set("health_status", "Unknown"); err != nil {
			return diag.FromErr(err)
		}
	} else if clusterSummary != nil && clusterSummary.Status != nil && clusterSummary.Status.Health != nil && clusterSummary.Status.Health.State != "" {
		// Set health status from cluster overview
		if err := d.Set("health_status", clusterSummary.Status.Health.State); err != nil {
			return diag.FromErr(err)
		}
	} else {
		// If health is not available, set to "Unknown"
		if err := d.Set("health_status", "Unknown"); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

// Update function - handles Day-2 operations
func resourceClusterBrownfieldUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)
	var diags diag.Diagnostics

	// Validate Day-1 fields are immutable
	if day1Diags := validateDay1FieldsImmutable(d); len(day1Diags) > 0 {
		return day1Diags
	}

	// Get cluster to check status and get cloud config info
	cluster, err := c.GetCluster(d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to get cluster: %w", err))
	}
	if cluster == nil {
		return diag.FromErr(fmt.Errorf("cluster not found"))
	}

	// Check if any Day-2 fields have changed
	day2Fields := []string{
		"cluster_profile", "backup_policy", "scan_policy", "cluster_rbac_binding",
		"namespaces", "host_config", "location_config", "cluster_timezone",
		"apply_setting", "review_repave_state", "description", "tags", "machine_pool",
		"pause_agent_upgrades",
	}
	hasDay2Changes := false
	for _, field := range day2Fields {
		if d.HasChange(field) {
			hasDay2Changes = true
			break
		}
	}

	// If Day-2 fields changed, validate cluster is Running-Healthy
	if hasDay2Changes {
		isHealthy, currentState := isClusterRunningHealthy(cluster, c)
		if !isHealthy {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "Cluster is not in Running-Healthy state",
				Detail:   fmt.Sprintf("Day-2 operations may not work as expected when the cluster is not in Running-Healthy state. Current state: %s", currentState),
			})
		}

		// Validate system repave approval if review_repave_state changed
		if d.HasChange("review_repave_state") {
			if err := validateSystemRepaveApproval(d, c); err != nil {
				return diag.FromErr(err)
			}
		}

		// Handle machine_pool changes for node actions
		if d.HasChange("machine_pool") {
			cloudConfigId := d.Get("cloud_config_id").(string)
			if cloudConfigId == "" {
				return diag.Errorf("cloud_config_id is required for machine_pool operations but is not available. Please ensure the cluster has been imported and is Running-Healthy.")
			}

			cloudType := d.Get("cloud_type").(string)
			getNodeMaintenanceStatusFn := getNodeMaintenanceStatusForCloudType(c, cloudType)
			if getNodeMaintenanceStatusFn == nil {
				return diag.Errorf("node maintenance operations are not supported for cloud_type: %s", cloudType)
			}

			// Get new machine pools
			nraw := d.Get("machine_pool")
			if nraw == nil {
				return diags // No machine pools to process
			}

			ns := nraw.(*schema.Set)
			if ns == nil || ns.Len() == 0 {
				return diags // No machine pools to process
			}

			// Get cloud config kind from cluster
			cloudConfigKind := ""
			if cluster.Spec != nil && cluster.Spec.CloudConfigRef != nil {
				cloudConfigKind = cluster.Spec.CloudConfigRef.Kind
			}
			if cloudConfigKind == "" {
				cloudConfigKind = cloudType // Fallback to cloud_type
			}

			// Process all machine pools
			for _, mp := range ns.List() {
				machinePool := mp.(map[string]interface{})
				machinePoolName := machinePool["name"].(string)
				if machinePoolName == "" {
					return diag.Errorf("machine_pool.name is required for node actions")
				}

				nodes := machinePool["node"]
				if nodes == nil {
					continue // Skip machine pools without nodes
				}

				nodeList := nodes.([]interface{})
				if len(nodeList) == 0 {
					continue // Skip machine pools with empty node lists
				}

				// Resolve node_id for each node if not provided
				for _, n := range nodeList {
					node := n.(map[string]interface{})
					nodeID, hasNodeID := node["node_id"].(string)
					nodeName, hasNodeName := node["node_name"].(string)

					// If node_id is not provided but node_name is provided, resolve it
					if (!hasNodeID || nodeID == "") && hasNodeName && nodeName != "" {
						resolvedNodeID, err := resolveNodeID(c, cloudType, cloudConfigId, machinePoolName, nodeName)
						if err != nil {
							return diag.FromErr(fmt.Errorf("failed to resolve node_id for node '%s' in machine pool '%s': %w", nodeName, machinePoolName, err))
						}
						node["node_id"] = resolvedNodeID
					} else if (!hasNodeID || nodeID == "") && (!hasNodeName || nodeName == "") {
						return diag.Errorf("either node_id or node_name must be provided for each node in machine_pool '%s'", machinePoolName)
					}
				}

				// Create a machine pool structure with all nodes
				machinePoolForAction := map[string]interface{}{
					"node": nodeList,
				}

				// Call resourceNodeAction for node maintenance operations
				// Use machinePoolName (from machine_pool.name) as the MachineName parameter
				if err := resourceNodeAction(c, ctx, machinePoolForAction, getNodeMaintenanceStatusFn, cloudConfigKind, cloudConfigId, machinePoolName); err != nil {
					return diag.FromErr(fmt.Errorf("failed to perform node action on machine pool %s: %w", machinePoolName, err))
				}
			}
		}
	}

	// Update common fields for Day-2 operations
	updateDiags, done := updateCommonFields(d, c)
	if done {
		return updateDiags
	}
	diags = append(diags, updateDiags...)

	// Refresh state
	readDiags := resourceClusterBrownfieldRead(ctx, d, m)
	diags = append(diags, readDiags...)

	return diags
}

// Helper Functions

// toBrownfieldClusterMetadata converts Terraform schema to V1ObjectMetaInputEntity
func toBrownfieldClusterMetadata(d *schema.ResourceData) *models.V1ObjectMetaInputEntity {
	metadata := &models.V1ObjectMetaInputEntity{
		Name: d.Get("name").(string),
	}
	return metadata
}

// toBrownfieldClusterSpecGeneric converts Terraform schema to V1SpectroGenericClusterImportEntitySpec
func toBrownfieldClusterSpecGeneric(d *schema.ResourceData) *models.V1SpectroGenericClusterImportEntitySpec {
	spec := &models.V1SpectroGenericClusterImportEntitySpec{}
	spec.ClusterConfig = toImportClusterConfig(d)
	return spec
}

// toBrownfieldClusterSpecCloudStack converts Terraform schema to V1SpectroCloudStackClusterImportEntitySpec
func toBrownfieldClusterSpecCloudStack(d *schema.ResourceData) *models.V1SpectroCloudStackClusterImportEntitySpec {
	spec := &models.V1SpectroCloudStackClusterImportEntitySpec{}
	spec.ClusterConfig = toImportClusterConfig(d)
	return spec
}

// toBrownfieldClusterSpecMaas converts Terraform schema to V1SpectroMaasClusterImportEntitySpec
func toBrownfieldClusterSpecMaas(d *schema.ResourceData) *models.V1SpectroMaasClusterImportEntitySpec {
	spec := &models.V1SpectroMaasClusterImportEntitySpec{}
	spec.ClusterConfig = toImportClusterConfig(d)
	return spec
}

// toBrownfieldClusterSpecEdgeNative converts Terraform schema to V1SpectroEdgeNativeClusterImportEntitySpec
func toBrownfieldClusterSpecEdgeNative(d *schema.ResourceData) *models.V1SpectroEdgeNativeClusterImportEntitySpec {
	spec := &models.V1SpectroEdgeNativeClusterImportEntitySpec{}
	spec.ClusterConfig = toImportClusterConfig(d)
	return spec
}

// toBrownfieldClusterSpecAws converts Terraform schema to V1SpectroAwsClusterImportEntitySpec
func toBrownfieldClusterSpecAws(d *schema.ResourceData) *models.V1SpectroAwsClusterImportEntitySpec {
	spec := &models.V1SpectroAwsClusterImportEntitySpec{}
	spec.ClusterConfig = toImportClusterConfig(d)
	return spec
}

// toBrownfieldClusterSpecAzure converts Terraform schema to V1SpectroAzureClusterImportEntitySpec
func toBrownfieldClusterSpecAzure(d *schema.ResourceData) *models.V1SpectroAzureClusterImportEntitySpec {
	spec := &models.V1SpectroAzureClusterImportEntitySpec{}
	spec.ClusterConfig = toImportClusterConfig(d)
	return spec
}

// toBrownfieldClusterSpecGcp converts Terraform schema to V1SpectroGcpClusterImportEntitySpec
func toBrownfieldClusterSpecGcp(d *schema.ResourceData) *models.V1SpectroGcpClusterImportEntitySpec {
	spec := &models.V1SpectroGcpClusterImportEntitySpec{}
	spec.ClusterConfig = toImportClusterConfig(d)
	return spec
}

// toBrownfieldClusterSpecVsphere converts Terraform schema to V1SpectroVsphereClusterImportEntitySpec
func toBrownfieldClusterSpecVsphere(d *schema.ResourceData) *models.V1SpectroVsphereClusterImportEntitySpec {
	spec := &models.V1SpectroVsphereClusterImportEntitySpec{}
	spec.ClusterConfig = toImportClusterConfig(d)
	return spec
}

// / toImportClusterConfig converts Terraform schema to V1ImportClusterConfig
func toImportClusterConfig(d *schema.ResourceData) *models.V1ImportClusterConfig {
	config := &models.V1ImportClusterConfig{}

	// Set ImportMode if provided
	if importMode, ok := d.GetOk("import_mode"); ok {
		mode := importMode.(string)
		// Convert "read_only" to "read-only" for API
		switch mode {
		case "read_only":
			config.ImportMode = "read-only"
		case "full":
			// API expects empty string (or not set) for full mode
			// Leave config.ImportMode as empty string (default)
			config.ImportMode = ""
		}
	} else {
		// Default is "full" - API expects empty string
		config.ImportMode = ""
	}

	// Set Proxy if any proxy-related fields are provided (for vsphere and openshift clusters)
	_, hasProxy := d.GetOk("proxy")
	_, hasNoProxy := d.GetOk("no_proxy")
	_, hasHostPath := d.GetOk("host_path")
	_, hasContainerMountPath := d.GetOk("container_mount_path")

	if hasProxy || hasNoProxy || hasHostPath || hasContainerMountPath {
		proxySpec := &models.V1ClusterProxySpec{}

		if httpProxy, ok := d.GetOk("proxy"); ok {
			proxySpec.HTTPProxy = httpProxy.(string)
		}

		if noProxy, ok := d.GetOk("no_proxy"); ok {
			proxySpec.NoProxy = noProxy.(string)
		}

		if hostPath, ok := d.GetOk("host_path"); ok {
			proxySpec.CaHostPath = hostPath.(string)
		}

		if containerMountPath, ok := d.GetOk("container_mount_path"); ok {
			proxySpec.CaContainerMountPath = containerMountPath.(string)
		}

		// Only set proxy if at least one field is set
		if proxySpec.HTTPProxy != "" || proxySpec.NoProxy != "" || proxySpec.CaHostPath != "" || proxySpec.CaContainerMountPath != "" {
			config.Proxy = proxySpec
		}
	}

	return config
}

// readCommonFieldsBrownfield wraps readCommonFields to skip fields that don't exist in brownfield schema
func readCommonFieldsBrownfield(c *client.V1Client, d *schema.ResourceData, cluster *models.V1SpectroCluster) (diag.Diagnostics, bool) {
	// Set tags (always present)
	if err := d.Set("tags", flattenTags(cluster.Metadata.Labels)); err != nil {
		return diag.FromErr(err), true
	}

	// Set backup_policy if field exists
	if _, ok := d.GetOk("backup_policy"); ok {
		if policy, err := c.GetClusterBackupConfig(d.Id()); err != nil {
			return diag.FromErr(err), true
		} else if policy != nil && policy.Spec.Config != nil {
			if err := d.Set("backup_policy", flattenBackupPolicy(policy.Spec.Config, d)); err != nil {
				return diag.FromErr(err), true
			}
		}
	}

	// Set scan_policy if field exists
	if _, ok := d.GetOk("scan_policy"); ok {
		if policy, err := c.GetClusterScanConfig(d.Id()); err != nil {
			return diag.FromErr(err), true
		} else if policy != nil && policy.Spec.DriverSpec != nil {
			if err := d.Set("scan_policy", flattenScanPolicy(policy.Spec.DriverSpec)); err != nil {
				return diag.FromErr(err), true
			}
		}
	}

	// Set cluster_rbac_binding if field exists
	if _, ok := d.GetOk("cluster_rbac_binding"); ok {
		if rbac, err := c.GetClusterRbacConfig(d.Id()); err != nil {
			return diag.FromErr(err), true
		} else if rbac != nil && rbac.Items != nil {
			if err := d.Set("cluster_rbac_binding", flattenClusterRBAC(rbac.Items)); err != nil {
				return diag.FromErr(err), true
			}
		}
	}

	// Set namespaces if field exists
	if _, ok := d.GetOk("namespaces"); ok {
		if namespace, err := c.GetClusterNamespaceConfig(d.Id()); err != nil {
			return diag.FromErr(err), true
		} else if namespace != nil && namespace.Items != nil {
			if err := d.Set("namespaces", flattenClusterNamespaces(namespace.Items)); err != nil {
				return diag.FromErr(err), true
			}
		}
	}

	// Set cluster_timezone if field exists
	if cluster.Spec.ClusterConfig.Timezone != "" {
		if err := d.Set("cluster_timezone", cluster.Spec.ClusterConfig.Timezone); err != nil {
			return diag.FromErr(err), true
		}
	}

	// Set host_config if field exists
	if _, ok := d.GetOk("host_config"); ok {
		hostConfig := cluster.Spec.ClusterConfig.HostClusterConfig
		if hostConfig != nil && *hostConfig.IsHostCluster {
			flattenHostConfig := flattenHostConfig(hostConfig)
			if len(flattenHostConfig) > 0 {
				if err := d.Set("host_config", flattenHostConfig); err != nil {
					return diag.FromErr(err), true
				}
			}
		}
	}

	// Set review_repave_state if field exists
	if _, ok := d.GetOk("review_repave_state"); ok {
		if err := d.Set("review_repave_state", cluster.Status.Repave.State); err != nil {
			return diag.FromErr(err), true
		}
	}

	// Set pause_agent_upgrades - always set during read
	if err := d.Set("pause_agent_upgrades", getSpectroComponentsUpgrade(cluster)); err != nil {
		return diag.FromErr(err), true
	}

	// Set location_config (computed field - always set if available)
	if clusterStatus, err := c.GetClusterWithoutStatus(d.Id()); err != nil {
		return diag.FromErr(err), true
	} else if clusterStatus != nil && clusterStatus.Status != nil && clusterStatus.Status.Location != nil {
		if err := d.Set("location_config", flattenLocationConfig(clusterStatus.Status.Location)); err != nil {
			return diag.FromErr(err), true
		}
	}

	return diag.Diagnostics{}, false
}

// isClusterRunningHealthy checks if cluster is in Running-Healthy state
func isClusterRunningHealthy(cluster *models.V1SpectroCluster, c *client.V1Client) (bool, string) {
	if cluster == nil || cluster.Status == nil {
		return false, "Unknown"
	}

	state := cluster.Status.State
	if state != "Running" {
		return false, state
	}

	// Check health status
	clusterSummary, err := c.GetClusterOverview(cluster.Metadata.UID)
	if err != nil {
		// If we can't get overview, assume Running is enough
		return true, state
	}

	if clusterSummary != nil && clusterSummary.Status != nil && clusterSummary.Status.Health != nil {
		healthState := clusterSummary.Status.Health.State
		if healthState == "Healthy" {
			return true, state + "-" + healthState
		}
		return false, state + "-" + healthState
	}

	// If health is not available, Running is considered acceptable
	return true, state
}

// validateDay1FieldsImmutable validates that Day-1 fields are not being changed
func validateDay1FieldsImmutable(d *schema.ResourceData) diag.Diagnostics {
	var diags diag.Diagnostics

	day1Fields := []string{
		"name", "cloud_type", "import_mode", "host_path",
		"container_mount_path", "context", "proxy", "no_proxy",
	}

	changedFields := []string{}
	for _, field := range day1Fields {
		if d.HasChange(field) {
			changedFields = append(changedFields, field)
		}
	}

	if len(changedFields) > 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Day-1 fields cannot be updated",
			Detail:   fmt.Sprintf("The following Day-1 fields cannot be updated after creation: %v. These fields are immutable. If you need to change these fields, delete and recreate the resource.", changedFields),
		})
	}

	return diags
}

// getNodeMaintenanceStatusForCloudType returns the appropriate GetNodeMaintenanceStatus function based on cloud_type
func getNodeMaintenanceStatusForCloudType(c *client.V1Client, cloudType string) GetMaintenanceStatus {
	switch cloudType {
	case "aws":
		return c.GetNodeMaintenanceStatusAws
	case "azure":
		return c.GetNodeMaintenanceStatusAzure
	case "gcp":
		return c.GetNodeMaintenanceStatusGcp
	case "vsphere", "openshift":
		return c.GetNodeMaintenanceStatusVsphere
	case "generic", "eks-anywhere":
		return c.GetNodeMaintenanceStatusGeneric
	case "apache-cloudstack":
		return c.GetNodeMaintenanceStatusCloudStack
	case "maas":
		return c.GetNodeMaintenanceStatusMaas
	case "edge-native":
		return c.GetNodeMaintenanceStatusEdgeNative
	case "openstack":
		return c.GetNodeMaintenanceStatusOpenStack
	default:
		return nil
	}
}

// getMachinesListForCloudType returns the appropriate GetMachinesList function based on cloud_type
func getMachinesListForCloudType(c *client.V1Client, cloudType string) func(string, string) (map[string]string, error) {
	switch cloudType {
	case "aws":
		return c.GetMachinesListAws
	case "azure":
		return c.GetMachinesListAzure
	case "gcp":
		return c.GetMachinesListGcp
	case "vsphere", "openshift":
		return c.GetMachinesListVsphere
	case "generic", "eks-anywhere":
		return c.GetMachinesListGeneric
	case "apache-cloudstack":
		return c.GetMachinesListApacheCloudstack
	case "maas":
		return c.GetMachinesListMaas
	case "edge-native":
		return c.GetMachinesListEdgeNative
	case "openstack":
		return c.GetMachinesListOpenStack
	default:
		return nil
	}
}

// resolveNodeID resolves node_id from node_name by listing machines from the cloud config
// Returns the node_id (UID) if found, or an error if not found or if resolution fails
func resolveNodeID(c *client.V1Client, cloudType, cloudConfigUID, machinePoolName, nodeName string) (string, error) {
	getMachinesListFn := getMachinesListForCloudType(c, cloudType)
	if getMachinesListFn == nil {
		return "", fmt.Errorf("node_id resolution is not supported for cloud_type: %s", cloudType)
	}

	machinesMap, err := getMachinesListFn(cloudConfigUID, machinePoolName)
	if err != nil {
		return "", fmt.Errorf("failed to list machines: %w", err)
	}

	nodeID, found := machinesMap[nodeName]
	if !found {
		return "", fmt.Errorf("node with name '%s' not found in machine pool '%s'", nodeName, machinePoolName)
	}

	return nodeID, nil
}

// getClusterImportInfo extracts the kubectl command and manifest URL from a cluster object.
// Returns kubectl_command, manifest_url, and an error if the import link is not available.
func getClusterImportInfo(cluster *models.V1SpectroCluster) (kubectlCommand, manifestURL string, err error) {
	if cluster == nil {
		return "", "", fmt.Errorf("cluster is nil")
	}

	if cluster.Status == nil {
		return "", "", fmt.Errorf("cluster status is not available")
	}

	if cluster.Status.ClusterImport == nil {
		return "", "", fmt.Errorf("cluster import information is not available")
	}

	kubectlCommand = cluster.Status.ClusterImport.ImportLink
	if kubectlCommand == "" {
		return "", "", fmt.Errorf("import link is empty")
	}

	// Extract manifest URL from importLink
	// importLink format: "kubectl apply -f https://api.dev.spectrocloud.com/v1/spectroclusters/{uid}/import/manifest"
	manifestURL = extractManifestURL(kubectlCommand)

	return kubectlCommand, manifestURL, nil
}

// extractManifestURL extracts the manifest URL from the importLink string.
// importLink format: "kubectl apply -f https://api.dev.spectrocloud.com/v1/spectroclusters/{uid}/import/manifest"
// Returns: "https://api.dev.spectrocloud.com/v1/spectroclusters/{uid}/import/manifest"
func extractManifestURL(importLink string) string {
	// Remove "kubectl apply -f" prefix and trim whitespace
	prefix := "kubectl apply -f"
	if strings.HasPrefix(importLink, prefix) {
		return strings.TrimSpace(strings.TrimPrefix(importLink, prefix))
	}
	// If already a URL or no prefix, return as-is
	return strings.TrimSpace(importLink)
}

// resourceClusterBrownfieldImport imports an existing brownfield cluster into Terraform state
func resourceClusterBrownfieldImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	clusterID, scope, customCloudName, err := ParseResourceCustomCloudImportID(d)
	if err != nil {
		return nil, err
	}
	d.SetId(clusterID + ":" + scope)
	_ = d.Set("cloud_type", customCloudName)
	c, err := GetCommonCluster(d, m)
	if err != nil {
		return nil, err
	}

	diags := resourceClusterBrownfieldRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read cluster for import: %v", diags)
	}

	// cluster profile and common default cluster attribute is get set here
	err = flattenCommonAttributeForBrownfieldClusterImport(c, d)
	if err != nil {
		return nil, err
	}

	// Return the resource data. In most cases, this method is only used to
	// import one resource at a time, so you should return the resource data
	// in a slice with a single element.
	return []*schema.ResourceData{d}, nil
}

func flattenCommonAttributeForBrownfieldClusterImport(c *client.V1Client, d *schema.ResourceData) error {
	clusterProfiles, err := flattenClusterProfileForImport(c, d)
	if err != nil {
		return err
	}
	err = d.Set("cluster_profile", clusterProfiles)
	if err != nil {
		return err
	}

	var diags diag.Diagnostics
	cluster, err := resourceClusterRead(d, c, diags)
	if err != nil {
		return err
	}

	if cluster.Spec.ClusterConfig.Timezone != "" {
		if err := d.Set("cluster_timezone", cluster.Spec.ClusterConfig.Timezone); err != nil {
			return err
		}
	}

	if cluster.Metadata.Annotations["description"] != "" {
		if err := d.Set("description", cluster.Metadata.Annotations["description"]); err != nil {
			return err
		}
	}

	if cluster.Status.SpcApply != nil {
		err = d.Set("apply_setting", cluster.Status.SpcApply.ActionType)
		if err != nil {
			return err
		}
	}

	err = d.Set("pause_agent_upgrades", getSpectroComponentsUpgrade(cluster))
	if err != nil {
		return err
	}
	if cluster.Status.Repave != nil {
		if err = d.Set("review_repave_state", cluster.Status.Repave.State); err != nil {
			return err
		}
	}
	err = d.Set("force_delete", false)
	if err != nil {
		return err
	}
	err = d.Set("force_delete_delay", 20)
	if err != nil {
		return err
	}
	err = d.Set("skip_completion", false)
	if err != nil {
		return err
	}
	return nil
}

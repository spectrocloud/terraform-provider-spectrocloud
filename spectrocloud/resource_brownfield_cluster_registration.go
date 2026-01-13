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
)

func resourceBrownfieldClusterRegistration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBrownfieldClusterRegistrationCreate,
		ReadContext:   resourceBrownfieldClusterRegistrationRead,
		UpdateContext: resourceBrownfieldClusterRegistrationUpdate,
		DeleteContext: resourceBrownfieldClusterRegistrationDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Description: "Register an existing Kubernetes cluster (brownfield) with Spectro Cloud. This resource creates a cluster registration and provides the import link and manifest URL needed to complete the cluster import process.",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew: true
				Description: "The name of the cluster to be registered.",
			},
			"cloud_type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"aws",
					"eksa",
					"azure",
					"gcp",
					"vsphere",
					"openshift",
					"generic",
				}, false),
				Description: "The cloud type of the cluster. Supported values: `aws`, `eksa`, `azure`, `gcp`, `vsphere`, `openshift`, `generic`.",
			},
			"import_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "full",
				ValidateFunc: validation.StringInSlice([]string{"read_only", "full"}, false),
				Description:  "The import mode for the cluster. Allowed values are `read_only` (imports cluster with read-only permissions) or `full` (imports cluster with full permissions). Defaults to `full`.",
			},
			"host_path": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Location for Proxy CA cert on host nodes. This is the file path on the host where the Proxy CA certificate is stored.",
			},
			"container_mount_path": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Location to mount Proxy CA cert inside container. This is the file path inside the container where the Proxy CA certificate will be mounted.",
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"project", "tenant"}, false),
				Description:  "The context for the cluster registration. Allowed values are `project` or `tenant`. Defaults to `project`." + PROJECT_NAME_NUANCE,
			},
			"proxy": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Location to mount Proxy CA cert inside container.This field is an supports for vsphere and openshift clusters",
			},
			"no_proxy": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Location to mount Proxy CA cert inside container. This field is an supports for vsphere and openshift clusters.",
			},
			"manifest_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The URL of the import manifest. This is the actual manifest URL extracted from the kubectl_command.",
			},
			"kubectl_command": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The kubectl command to import the cluster. Format: `kubectl apply -f <manifest_url>`.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The current operational state of the cluster. Possible values include: `Pending`, `Provisioning`, `Running`, `Deleting`, `Deleted`, `Error`, `Importing`.",
			},
		},
	}
}

func resourceBrownfieldClusterRegistrationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
	case "vsphere":
		entity := &models.V1SpectroVsphereClusterImportEntity{
			Metadata: metadata,
			Spec:     toBrownfieldClusterSpecVsphere(d),
		}
		clusterUID, err = c.ImportSpectroVsphereCluster(entity)
	case "generic":
		entity := &models.V1SpectroGenericClusterImportEntity{
			Metadata: metadata,
			Spec:     toBrownfieldClusterSpecGeneric(d),
		}
		clusterUID, err = c.ImportSpectroClusterGeneric(entity)
	case "eksa", "openshift":
		// For EKS-Anywhere and OpenShift, use Generic import
		entity := &models.V1SpectroGenericClusterImportEntity{
			Metadata: metadata,
			Spec:     toBrownfieldClusterSpecGeneric(d),
		}
		clusterUID, err = c.ImportSpectroClusterGeneric(entity)
	default:
		return diag.FromErr(fmt.Errorf("unsupported cloud type: %s", cloudType))
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to register brownfield cluster: %w", err))
	}

	// Set the cluster UID as the resource ID
	registrationClusterUID := fmt.Sprintf("registration_%s", clusterUID)
	d.SetId(registrationClusterUID)
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
	}

	return diags
}

// resourceBrownfieldClusterRegistrationUpdate handles update operations for brownfield cluster registration.
// Day-2 operations are not supported - updates are not allowed. Returns a warning and refreshes state.
func resourceBrownfieldClusterRegistrationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Check if any fields have changed
	if d.HasChanges("name", "cloud_type", "context", "import_mode", "host_path", "container_mount_path", "proxy", "no_proxy") {
		// Return warning that updates are not supported
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Day-2 operation not supported for update",
			Detail:   "Day-2 operation not supported for update, if required delete and recreate it again. The resource state will be refreshed to match the current cluster state.",
		})
	}

	// Refresh the state by reading the current cluster state
	readDiags := resourceBrownfieldClusterRegistrationRead(ctx, d, m)
	diags = append(diags, readDiags...)

	return diags
}

// Helper functions
// extractClusterUIDFromResourceID extracts the actual cluster UID from the resource ID.
// Resource ID format: "registration_<clusterUID>"
// Returns the clusterUID and an error if the format is invalid.
func extractClusterUIDFromResourceID(resourceID string) (string, error) {
	prefix := "registration_"
	if !strings.HasPrefix(resourceID, prefix) {
		return "", fmt.Errorf("invalid resource ID format: expected 'registration_<clusterUID>', got: %s", resourceID)
	}
	clusterUID := strings.TrimPrefix(resourceID, prefix)
	if clusterUID == "" {
		return "", fmt.Errorf("invalid resource ID format: cluster UID is empty")
	}
	return clusterUID, nil
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

// Helper functions

// toBrownfieldClusterMetadata converts Terraform schema to V1ObjectMetaInputEntity
func toBrownfieldClusterMetadata(d *schema.ResourceData) *models.V1ObjectMetaInputEntity {
	metadata := &models.V1ObjectMetaInputEntity{
		Name: d.Get("name").(string),
	}
	return metadata
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

// toBrownfieldClusterSpecGeneric converts Terraform schema to V1SpectroGenericClusterImportEntitySpec
func toBrownfieldClusterSpecGeneric(d *schema.ResourceData) *models.V1SpectroGenericClusterImportEntitySpec {
	spec := &models.V1SpectroGenericClusterImportEntitySpec{}
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
		if mode == "read_only" {
			mode = "read-only"
		}
		config.ImportMode = mode
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

// Read function - reads the current state of the cluster
func resourceBrownfieldClusterRegistrationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)
	var diags diag.Diagnostics

	resourceID := d.Id()
	clusterUID, err := extractClusterUIDFromResourceID(resourceID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("invalid resource ID: %w", err))
	}
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

	// Set basic fields
	if err := d.Set("name", cluster.Metadata.Name); err != nil {
		return diag.FromErr(err)
	}

	// Set cloud_type from cluster spec
	if cluster.Spec != nil && cluster.Spec.CloudType != "" {
		if err := d.Set("cloud_type", cluster.Spec.CloudType); err != nil {
			return diag.FromErr(err)
		}
	}

	// Set status if available
	if cluster.Status != nil && cluster.Status.State != "" {
		if err := d.Set("status", cluster.Status.State); err != nil {
			return diag.FromErr(err)
		}
	}

	// Get the import link and manifest URL from cluster object
	kubectlCommand, manifestURL, err := getClusterImportInfo(cluster)
	if err != nil {
		// Import link may not be available - this is not necessarily an error
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "kubectl_command not available",
			Detail:   fmt.Sprintf("kubectl_command is not yet available for cluster %s: %v", clusterUID, err),
		})
	} else {
		if err := d.Set("kubectl_command", kubectlCommand); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("manifest_url", manifestURL); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

// Delete function - deletes the cluster registration
func resourceBrownfieldClusterRegistrationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)
	var diags diag.Diagnostics

	resourceID := d.Id()
	clusterUID, err := extractClusterUIDFromResourceID(resourceID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("invalid resource ID: %w", err))
	}

	// Delete the cluster registration
	err = c.DeleteCluster(clusterUID)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return diags
}

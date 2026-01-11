package spectrocloud

import (
	"context"
	"fmt"
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
		Importer: &schema.ResourceImporter{
			StateContext: resourceBrownfieldClusterRegistrationImport,
		},
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
				Description: "The name of the cluster to be registered.",
			},
			"cloud_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"AWS",
					"EKS-Anywhere",
					"Azure",
					"Google Cloud",
					"VMWare VSphere",
					"OpenShift",
					"Generic",
				}, false),
				Description: "The cloud type of the cluster. Supported values: `AWS`, `EKS-Anywhere`, `Azure`, `Google Cloud`, `VMWare VSphere`, `OpenShift`, `Generic`.",
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
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Proxy configuration for the cluster import.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ca_host_path": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Location for Proxy CA cert on host nodes. This is the file path on the host where the Proxy CA certificate is stored.",
						},
						"ca_container_mount_path": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Location to mount Proxy CA cert inside container. This is the file path inside the container where the Proxy CA certificate will be mounted.",
						},
						"http_proxy": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "URL for HTTP requests unless overridden by NoProxy.",
						},
						"https_proxy": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "URL for HTTPS requests unless overridden by NoProxy.",
						},
						"no_proxy": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Comma-separated list of hostnames and/or CIDRs that should not use the proxy. Represents the NO_PROXY or no_proxy environment variable.",
						},
					},
				},
			},

			// "spec": {
			// 	Type:        schema.TypeList,
			// 	Optional:    true,
			// 	MaxItems:    1,
			// 	Description: "Specification for the cluster import. Structure varies by cloud type.",
			// 	Elem: &schema.Resource{
			// 		Schema: map[string]*schema.Schema{
			// 			"cluster_config": {
			// 				Type:        schema.TypeList,
			// 				Optional:    true,
			// 				MaxItems:    1,
			// 				Description: "Cluster configuration for import.",
			// 				Elem: &schema.Resource{
			// 					Schema: map[string]*schema.Schema{
			// 						"endpoint": {
			// 							Type:        schema.TypeString,
			// 							Optional:    true,
			// 							Description: "The Kubernetes API endpoint URL.",
			// 						},
			// 						"ca_cert": {
			// 							Type:        schema.TypeString,
			// 							Optional:    true,
			// 							Sensitive:   true,
			// 							Description: "The CA certificate for the Kubernetes cluster.",
			// 						},
			// 						"token": {
			// 							Type:        schema.TypeString,
			// 							Optional:    true,
			// 							Sensitive:   true,
			// 							Description: "The authentication token for the Kubernetes cluster.",
			// 						},
			// 					},
			// 				},
			// 			},
			// 		},
			// 	},
			// },
			"tags": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "A list of tags to be applied to the cluster. Tags are typically in the form of `key:value`.",
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
			"cluster_uid": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The unique identifier (UID) of the registered cluster.",
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
		clusterUID, err = c.PostSpectroClusterAwsImport(entity)
	case "azure":
		entity := &models.V1SpectroAzureClusterImportEntity{
			Metadata: metadata,
			Spec:     toBrownfieldClusterSpecAzure(d),
		}
		clusterUID, err = c.PostSpectroClusterAzureImport(entity)
	case "gcp":
		entity := &models.V1SpectroGcpClusterImportEntity{
			Metadata: metadata,
			Spec:     toBrownfieldClusterSpecGcp(d),
		}
		clusterUID, err = c.PostSpectroClusterGcpImport(entity)
	case "vsphere":
		entity := &models.V1SpectroVsphereClusterImportEntity{
			Metadata: metadata,
			Spec:     toBrownfieldClusterSpecVsphere(d),
		}
		clusterUID, err = c.PostSpectroVsphereClusterImport(entity)
	case "generic":
		entity := &models.V1SpectroGenericClusterImportEntity{
			Metadata: metadata,
			Spec:     toBrownfieldClusterSpecGeneric(d),
		}
		clusterUID, err = c.PostSpectroClusterGenericImport(entity)
	case "EKS-Anywhere", "OpenShift":
		// For EKS-Anywhere and OpenShift, use Generic import
		entity := &models.V1SpectroGenericClusterImportEntity{
			Metadata: metadata,
			Spec:     toBrownfieldClusterSpecGeneric(d),
		}
		clusterUID, err = c.PostSpectroClusterGenericImport(entity)
	default:
		return diag.FromErr(fmt.Errorf("unsupported cloud type: %s", cloudType))
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to register brownfield cluster: %w", err))
	}

	// Set the cluster UID as the resource ID
	d.SetId(clusterUID)

	// Get the import link and manifest URL
	// Note: The import link may not be immediately available, so we may need to retry
	kubectl_command, manifestURL, err := c.GetClusterImportLink(clusterUID)
	if err != nil {
		// Log warning but don't fail - import link may not be available immediately
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Import link not immediately available",
			Detail:   fmt.Sprintf("Cluster registered successfully, but import link is not yet available: %v. You may need to run 'terraform refresh' to get the import link.", err),
		})
	} else {
		if err := d.Set("kubectl_command", kubectl_command); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("manifest_url", manifestURL); err != nil {
			return diag.FromErr(err)
		}
	}

	if err := d.Set("cluster_uid", clusterUID); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceBrownfieldClusterRegistrationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	// Set basic fields
	if err := d.Set("name", cluster.Metadata.Name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("cluster_uid", clusterUID); err != nil {
		return diag.FromErr(err)
	}

	// Get the import link and manifest URL
	kubectl_command, manifestURL, err := c.GetClusterImportLink(clusterUID)
	if err != nil {
		// Import link may not be available yet - this is not necessarily an error
		// Just log it and continue
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "kubectl_command not available",
			Detail:   fmt.Sprintf("kubectl_command is not yet available for cluster %s: %v", clusterUID, err),
		})
	} else {
		if err := d.Set("kubectl_command", kubectl_command); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("manifest_url", manifestURL); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

// func resourceBrownfieldClusterRegistrationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
// 	// Brownfield cluster registration is typically immutable after creation
// 	// Most fields cannot be updated. Only metadata annotations/labels might be updatable.
// 	// For now, we'll just read the current state
// 	return resourceBrownfieldClusterRegistrationRead(ctx, d, m)
// }

func resourceBrownfieldClusterRegistrationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)
	var diags diag.Diagnostics

	clusterUID := d.Id()

	// Delete the cluster registration
	err := c.DeleteCluster(clusterUID)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return diags
}

// Imports an existing cluster registration into Terraform state
func resourceBrownfieldClusterRegistrationImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	// Import format: cluster_uid:context (optional context)
	importID := d.Id()
	if importID == "" {
		return nil, fmt.Errorf("cluster UID is required for import")
	}

	// Parse import ID (format: cluster_uid or cluster_uid:context)
	var clusterUID, context string
	parts := splitImportID(importID)
	if len(parts) == 2 {
		clusterUID = parts[0]
		context = parts[1]
	} else {
		clusterUID = importID
		context = "project" // default
	}

	// Set the context if provided
	if context != "" {
		if err := d.Set("context", context); err != nil {
			return nil, err
		}
	}

	// Set the cluster UID
	d.SetId(clusterUID)

	// Read the cluster data
	diags := resourceBrownfieldClusterRegistrationRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read brownfield cluster registration for import: %v", diags)
	}

	return []*schema.ResourceData{d}, nil
}

// Helper functions

func toBrownfieldClusterMetadata(d *schema.ResourceData) *models.V1ObjectMetaInputEntity {
	metadata := &models.V1ObjectMetaInputEntity{
		Name: d.Get("name").(string),
	}

	if v, ok := d.GetOk("metadata"); ok && len(v.([]interface{})) > 0 {
		metaMap := v.([]interface{})[0].(map[string]interface{})

		if labels, ok := metaMap["labels"].(map[string]interface{}); ok && len(labels) > 0 {
			metadata.Labels = make(map[string]string)
			for k, v := range labels {
				metadata.Labels[k] = v.(string)
			}
		}

		if annotations, ok := metaMap["annotations"].(map[string]interface{}); ok && len(annotations) > 0 {
			metadata.Annotations = make(map[string]string)
			for k, v := range annotations {
				metadata.Annotations[k] = v.(string)
			}
		}
	}

	return metadata
}

func toBrownfieldClusterSpecAws(d *schema.ResourceData) *models.V1SpectroAwsClusterImportEntitySpec {
	spec := &models.V1SpectroAwsClusterImportEntitySpec{}
	if v, ok := d.GetOk("spec"); ok && len(v.([]interface{})) > 0 {
		specMap := v.([]interface{})[0].(map[string]interface{})
		if clusterConfig, ok := specMap["cluster_config"].([]interface{}); ok && len(clusterConfig) > 0 {
			spec.ClusterConfig = toImportClusterConfig(clusterConfig[0].(map[string]interface{}))
		}
	}
	return spec
}

func toBrownfieldClusterSpecAzure(d *schema.ResourceData) *models.V1SpectroAzureClusterImportEntitySpec {
	spec := &models.V1SpectroAzureClusterImportEntitySpec{}
	if v, ok := d.GetOk("spec"); ok && len(v.([]interface{})) > 0 {
		specMap := v.([]interface{})[0].(map[string]interface{})
		if clusterConfig, ok := specMap["cluster_config"].([]interface{}); ok && len(clusterConfig) > 0 {
			spec.ClusterConfig = toImportClusterConfig(clusterConfig[0].(map[string]interface{}))
		}
	}
	return spec
}

func toBrownfieldClusterSpecGcp(d *schema.ResourceData) *models.V1SpectroGcpClusterImportEntitySpec {
	spec := &models.V1SpectroGcpClusterImportEntitySpec{}
	if v, ok := d.GetOk("spec"); ok && len(v.([]interface{})) > 0 {
		specMap := v.([]interface{})[0].(map[string]interface{})
		if clusterConfig, ok := specMap["cluster_config"].([]interface{}); ok && len(clusterConfig) > 0 {
			spec.ClusterConfig = toImportClusterConfig(clusterConfig[0].(map[string]interface{}))
		}
	}
	return spec
}

func toBrownfieldClusterSpecVsphere(d *schema.ResourceData) *models.V1SpectroVsphereClusterImportEntitySpec {
	spec := &models.V1SpectroVsphereClusterImportEntitySpec{}
	if v, ok := d.GetOk("spec"); ok && len(v.([]interface{})) > 0 {
		specMap := v.([]interface{})[0].(map[string]interface{})
		if clusterConfig, ok := specMap["cluster_config"].([]interface{}); ok && len(clusterConfig) > 0 {
			spec.ClusterConfig = toImportClusterConfig(clusterConfig[0].(map[string]interface{}))
		}
	}
	return spec
}

func toBrownfieldClusterSpecGeneric(d *schema.ResourceData) *models.V1SpectroGenericClusterImportEntitySpec {
	spec := &models.V1SpectroGenericClusterImportEntitySpec{}
	if v, ok := d.GetOk("spec"); ok && len(v.([]interface{})) > 0 {
		specMap := v.([]interface{})[0].(map[string]interface{})
		if clusterConfig, ok := specMap["cluster_config"].([]interface{}); ok && len(clusterConfig) > 0 {
			spec.ClusterConfig = toImportClusterConfig(clusterConfig[0].(map[string]interface{}))
		}
	}
	return spec
}

func toImportClusterConfig(configMap map[string]interface{}) *models.V1ImportClusterConfig {
	config := &models.V1ImportClusterConfig{}

	if endpoint, ok := configMap["endpoint"].(string); ok && endpoint != "" {
		config.Endpoint = &endpoint
	}

	if caCert, ok := configMap["ca_cert"].(string); ok && caCert != "" {
		config.CaCert = &caCert
	}

	if token, ok := configMap["token"].(string); ok && token != "" {
		config.Token = &token
	}

	return config
}

func splitImportID(importID string) []string {
	// Simple split by colon - can be enhanced if needed
	parts := []string{}
	lastIndex := 0
	for i, char := range importID {
		if char == ':' {
			parts = append(parts, importID[lastIndex:i])
			lastIndex = i + 1
		}
	}
	if lastIndex < len(importID) {
		parts = append(parts, importID[lastIndex:])
	}
	return parts
}

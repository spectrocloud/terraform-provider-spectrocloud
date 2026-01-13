package spectrocloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

func dataSourceBrownfieldClusterRegistration() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceBrownfieldClusterRegistrationRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:          schema.TypeString,
				Computed:      true,
				Optional:      true,
				ConflictsWith: []string{"name"},
				Description:   "The unique identifier (UID) of the registered cluster. Either `id` or `name` must be provided.",
			},
			"name": {
				Type:          schema.TypeString,
				Computed:      true,
				Optional:      true,
				ConflictsWith: []string{"id"},
				Description:   "The name of the cluster. Either `id` or `name` must be provided.",
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"project", "tenant"}, false),
				Description:  "The context for the cluster registration. Allowed values are `project` or `tenant`. Defaults to `project`." + PROJECT_NAME_NUANCE,
			},
			"cloud_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The cloud type of the cluster. Supported values: `aws`, `eksa`, `azure`, `gcp`, `vsphere`, `openshift`, `generic`.",
			},
			"import_mode": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The import mode for the cluster. Possible values: `read_only` or `full`.",
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
			"cluster_uid": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The unique identifier (UID) of the registered cluster.",
			},
		},
	}
}

func dataSourceBrownfieldClusterRegistrationRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)
	var diags diag.Diagnostics

	var cluster *models.V1SpectroCluster
	var err error

	// Lookup by ID if provided
	if clusterUID, ok := d.GetOk("id"); ok {
		cluster, err = c.GetCluster(clusterUID.(string))
		if err != nil {
			return handleReadError(d, err, diags)
		}
		if cluster == nil {
			// Cluster not found
			d.SetId("")
			return diags
		}
		d.SetId(cluster.Metadata.UID)
	} else if name, ok := d.GetOk("name"); ok {
		// Lookup by name (brownfield clusters are not virtual)
		cluster, err = c.GetClusterByName(name.(string), false)
		if err != nil {
			return handleReadError(d, err, diags)
		}
		if cluster == nil {
			// Cluster not found
			d.SetId("")
			return diags
		}
		d.SetId(cluster.Metadata.UID)
	} else {
		return diag.FromErr(fmt.Errorf("either 'id' or 'name' must be provided"))
	}

	// Set basic fields
	if err := d.Set("name", cluster.Metadata.Name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("cluster_uid", cluster.Metadata.UID); err != nil {
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
		// Just log it and continue
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "kubectl_command not available",
			Detail:   fmt.Sprintf("kubectl_command is not yet available for cluster %s: %v", cluster.Metadata.UID, err),
		})
	} else {
		if err := d.Set("kubectl_command", kubectlCommand); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("manifest_url", manifestURL); err != nil {
			return diag.FromErr(err)
		}
	}

	// Set import_mode from cluster config if available
	if cluster.Spec != nil && cluster.Spec.ClusterConfig != nil {
		// Check if there's import mode information in the cluster config
		// Note: This may need to be adjusted based on actual API response structure
		// The import_mode might be stored in a different location
	}

	return diags
}

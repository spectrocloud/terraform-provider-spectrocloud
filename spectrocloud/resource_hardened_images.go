package spectrocloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
)

// hardenedImagesID is the fixed identifier for the tenant hardened images singleton resource.
const hardenedImagesID = "default-hardened-images-id"

// hardenedImagesPageSize is the maximum page size the propagation status API accepts.
const hardenedImagesPageSize int64 = 50

// hardenedImagesMaxPages bounds the paging loop so a server that fails to advance its offset
// cannot spin forever. At 50 clusters per page this covers 50,000 clusters.
const hardenedImagesMaxPages = 1000

func resourceHardenedImages() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHardenedImagesCreate,
		ReadContext:   resourceHardenedImagesRead,
		UpdateContext: resourceHardenedImagesUpdate,
		DeleteContext: resourceHardenedImagesDelete,
		Description: "Resource for managing hardened image adoption across a tenant in Spectro Cloud. " +
			"Palette propagates the Spectro Cloud registry image pull secret to every managed workload cluster so " +
			"that clusters can pull hardened images. Use this resource to opt individual clusters out of that " +
			"propagation and to observe the rollout status across the tenant.",
		Importer: &schema.ResourceImporter{
			StateContext: resourceHardenedImagesImport,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"excluded_cluster_uids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringIsNotEmpty,
				},
				Description: "Unique identifiers of the clusters that must not receive the hardened image pull " +
					"secret. Palette keeps every other cluster in the tenant enrolled. Dropping an identifier from " +
					"this set re-enrolls that cluster on the next apply.",
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
				Description: "Aggregated rollout progress for the tenant. Palette reports `Completed` once every " +
					"enrolled cluster holds the secret, `InProgress` while the rollout is running, and `Failed` " +
					"when at least one cluster could not be updated.",
			},
			"failed_clusters_count": {
				Type:     schema.TypeInt,
				Computed: true,
				Description: "Number of enrolled clusters that Palette could not deliver the image pull secret to. " +
					"Inspect the `clusters` attribute for the per-cluster failure reason.",
			},
			"clusters": {
				Type:     schema.TypeList,
				Computed: true,
				Description: "One entry per cluster in the tenant, reporting how far the hardened image pull " +
					"secret rollout has progressed on that cluster. Clusters that are currently opted out are " +
					"listed too.",
				Elem: &schema.Resource{
					Schema: hardenedImagesClusterSchema(),
				},
			},
		},
	}
}

// hardenedImagesClusterSchema describes one row of the tenant rollout report. It is shared by the
// resource and the data source so both expose an identical shape.
func hardenedImagesClusterSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"uid": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Unique identifier Palette assigned to the cluster.",
		},
		"name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Display name of the cluster as it appears in Palette.",
		},
		"project_uid": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Unique identifier of the project the cluster belongs to. Empty for tenant-scoped clusters.",
		},
		"project_name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Display name of the project the cluster belongs to. Empty for tenant-scoped clusters.",
		},
		"exclude": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "Set to `true` when the cluster is opted out of hardened image pull secret propagation.",
		},
		"state": {
			Type:     schema.TypeString,
			Computed: true,
			Description: "Rollout progress for this cluster. Palette reports `Completed`, `InProgress`, or " +
				"`Failed`.",
		},
		"reason": {
			Type:     schema.TypeString,
			Computed: true,
			Description: "Short machine-readable cause when the rollout did not succeed, such as " +
				"`ConnectivityIssue`, `PauseAgentUpgrades`, or `LocalPropagationFailure`.",
		},
		"message": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Human-readable detail that accompanies the failure reason, when Palette supplies one.",
		},
	}
}

// getHardenedImagesStatus walks every page of the tenant propagation status API and returns the
// aggregate state together with the full cluster list. clusterName and projectUID are optional
// server-side filters; pass empty strings to retrieve the whole tenant.
func getHardenedImagesStatus(c *client.V1Client, clusterName, projectUID string) (*models.V1ImagePullSecretTenantPropagationStatus, error) {
	aggregate := &models.V1ImagePullSecretTenantPropagationStatus{}

	var offset int64
	for page := 0; page < hardenedImagesMaxPages; page++ {
		status, err := c.GetImagePullSecretStatus(clusterName, projectUID, hardenedImagesPageSize, offset)
		if err != nil {
			return nil, err
		}
		if status == nil {
			break
		}

		// The tenant-wide fields are identical on every page, so the last page read wins.
		aggregate.State = status.State
		aggregate.Clusters = status.Clusters
		aggregate.Items = append(aggregate.Items, status.Items...)

		if int64(len(status.Items)) < hardenedImagesPageSize {
			return aggregate, nil
		}
		offset += int64(len(status.Items))
	}

	return nil, fmt.Errorf("hardened image propagation status exceeded %d pages; refusing to keep paging", hardenedImagesMaxPages)
}

// flattenHardenedImagesClusters converts the API rollout report into the `clusters` attribute shape.
func flattenHardenedImagesClusters(items []*models.V1ImagePullSecretTenantPropagationClusterStatus) []interface{} {
	clusters := make([]interface{}, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		cluster := map[string]interface{}{
			"exclude": item.Exclude,
			"state":   string(item.State),
			"reason":  item.Reason,
			"message": item.Message,
		}
		if item.Cluster != nil {
			cluster["uid"] = item.Cluster.UID
			cluster["name"] = item.Cluster.Name
		}
		if item.Project != nil {
			cluster["project_uid"] = item.Project.UID
			cluster["project_name"] = item.Project.Name
		}
		clusters = append(clusters, cluster)
	}
	return clusters
}

// excludedClusterUIDs returns the identifiers of the clusters the API currently reports as opted out.
func excludedClusterUIDs(items []*models.V1ImagePullSecretTenantPropagationClusterStatus) []string {
	uids := make([]string, 0)
	for _, item := range items {
		if item == nil || !item.Exclude || item.Cluster == nil {
			continue
		}
		uids = append(uids, item.Cluster.UID)
	}
	return uids
}

// flattenHardenedImages writes a propagation status report into state.
func flattenHardenedImages(status *models.V1ImagePullSecretTenantPropagationStatus, d *schema.ResourceData) error {
	if err := d.Set("state", string(status.State)); err != nil {
		return err
	}

	var failed int64
	if status.Clusters != nil {
		failed = status.Clusters.Failed
	}
	if err := d.Set("failed_clusters_count", failed); err != nil {
		return err
	}

	return d.Set("clusters", flattenHardenedImagesClusters(status.Items))
}

// applyHardenedImagesExclusions converges the tenant on the exclusion set held in the configuration.
// It reads the current server-side set, excludes the clusters that were added, and re-enrolls the
// clusters that were dropped.
func applyHardenedImagesExclusions(d *schema.ResourceData, c *client.V1Client) error {
	status, err := getHardenedImagesStatus(c, "", "")
	if err != nil {
		return err
	}

	current := make(map[string]bool)
	for _, uid := range excludedClusterUIDs(status.Items) {
		current[uid] = true
	}

	desired := make(map[string]bool)
	for _, raw := range d.Get("excluded_cluster_uids").(*schema.Set).List() {
		desired[raw.(string)] = true
	}

	toExclude := make([]string, 0)
	for uid := range desired {
		if !current[uid] {
			toExclude = append(toExclude, uid)
		}
	}

	toInclude := make([]string, 0)
	for uid := range current {
		if !desired[uid] {
			toInclude = append(toInclude, uid)
		}
	}

	if len(toExclude) > 0 {
		if err := c.UpdateImagePullSecretClusterExclusion(toExclude, true); err != nil {
			return err
		}
	}
	if len(toInclude) > 0 {
		if err := c.UpdateImagePullSecretClusterExclusion(toInclude, false); err != nil {
			return err
		}
	}
	return nil
}

func resourceHardenedImagesCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, tenantString)
	// Hardened image propagation is always on for a tenant, so creation only applies the exclusions.
	if err := applyHardenedImagesExclusions(d, c); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(hardenedImagesID)
	return resourceHardenedImagesRead(ctx, d, m)
}

func resourceHardenedImagesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, tenantString)
	var diags diag.Diagnostics

	status, err := getHardenedImagesStatus(c, "", "")
	if err != nil {
		return handleReadError(d, err, diags)
	}
	// handling case for cross-plane for singleton resource
	if d.Id() != hardenedImagesID {
		d.SetId("")
		return diags
	}
	if err := d.Set("excluded_cluster_uids", excludedClusterUIDs(status.Items)); err != nil {
		return diag.FromErr(err)
	}
	if err := flattenHardenedImages(status, d); err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceHardenedImagesUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, tenantString)
	if err := applyHardenedImagesExclusions(d, c); err != nil {
		return diag.FromErr(err)
	}
	return resourceHardenedImagesRead(ctx, d, m)
}

func resourceHardenedImagesDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	// Palette has no delete operation for hardened image propagation, and destroying this resource
	// deliberately leaves the tenant exactly as it is: clusters that were opted out stay opted out.
	// This differs from the other tenant singletons, which reset to the Palette defaults on destroy.
	// Re-enroll a cluster by dropping it from excluded_cluster_uids and applying before destroying.
	d.SetId("")
	return diags
}

func resourceHardenedImagesImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c := getV1ClientWithResourceContext(m, tenantString)

	// Resolve org name → tenant UID if a name was provided instead of a UID
	resolvedTenantUID, err := resolveUidorNameToContextID(m, c, d.Id())
	if err != nil {
		return nil, err
	}

	actualTenantId, err := c.GetTenantUID()
	if err != nil {
		return nil, err
	}
	if resolvedTenantUID != actualTenantId {
		return nil, fmt.Errorf("invalid import: tenant %q does not match your authorized tenant UID %q", d.Id(), actualTenantId)
	}

	// Set the canonical ID so resourceHardenedImagesRead passes its singleton check
	d.SetId(hardenedImagesID)

	diags := resourceHardenedImagesRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read hardened images configuration for import: %v", diags)
	}
	return []*schema.ResourceData{d}, nil
}

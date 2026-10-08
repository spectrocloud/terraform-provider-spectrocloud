package spectrocloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceHardenedImages() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHardenedImagesRead,
		Description: "Provides the hardened image rollout status for a tenant. Palette propagates the Spectro " +
			"Cloud registry image pull secret to managed workload clusters so they can pull hardened images, and " +
			"this data source reports how far that rollout has progressed.",
		Schema: map[string]*schema.Schema{
			"cluster_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Restrict the report to clusters whose name matches this value.",
			},
			"project_uid": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Restrict the report to clusters that belong to the project with this unique identifier.",
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
			"excluded_cluster_uids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Description: "Unique identifiers of the clusters that are currently opted out of hardened image " +
					"pull secret propagation.",
			},
			"clusters": {
				Type:     schema.TypeList,
				Computed: true,
				Description: "One entry per cluster the filters matched, reporting how far the hardened image " +
					"pull secret rollout has progressed on that cluster. Clusters that are currently opted out " +
					"are listed too.",
				Elem: &schema.Resource{
					Schema: hardenedImagesClusterSchema(),
				},
			},
		},
	}
}

func dataSourceHardenedImagesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, tenantString)
	var diags diag.Diagnostics

	tenantUID, err := c.GetTenantUID()
	if err != nil {
		return diag.FromErr(err)
	}

	status, err := getHardenedImagesStatus(c, d.Get("cluster_name").(string), d.Get("project_uid").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("excluded_cluster_uids", excludedClusterUIDs(status.Items)); err != nil {
		return diag.FromErr(err)
	}
	if err := flattenHardenedImages(status, d); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(tenantUID)
	return diags
}

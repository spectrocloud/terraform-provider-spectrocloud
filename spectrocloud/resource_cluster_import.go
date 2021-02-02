package spectrocloud

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceClusterImport() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCloudClusterImport,
		ReadContext:   resourceCloudClusterRead,
		UpdateContext: resourceCloudClusterUpdate,
		DeleteContext: resourceClusterDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 2,
		Schema: map[string]*schema.Schema{
			name: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			cloud: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			cloud_config_id: {
				Type:     schema.TypeString,
				Computed: true,
			},
			cluster_import_manifest_url: {
				Type:     schema.TypeString,
				Computed: true,
			},
			cluster_import_manifest: {
				Type:     schema.TypeString,
				Computed: true,
			},
			cluster_prrofile_id: {
				Type:     schema.TypeString,
				Optional: true,
			},
			pack: {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      resourcePackHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						name: {
							Type:     schema.TypeString,
							Required: true,
						},
						tag: {
							Type:     schema.TypeString,
							Required: true,
						},
						values: {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

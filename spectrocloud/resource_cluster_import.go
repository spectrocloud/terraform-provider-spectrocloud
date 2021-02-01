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
			Name: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			Cloud: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			CloudConfigId: {
				Type:     schema.TypeString,
				Computed: true,
			},
			ClusterImportManifestUrl: {
				Type:     schema.TypeString,
				Computed: true,
			},
			ClusterImportManifest: {
				Type:     schema.TypeString,
				Computed: true,
			},
			ClusterProfileId: {
				Type:     schema.TypeString,
				Optional: true,
			},
			Pack: {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      resourcePackHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						Name: {
							Type:     schema.TypeString,
							Required: true,
						},
						Tag: {
							Type:     schema.TypeString,
							Required: true,
						},
						Values: {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

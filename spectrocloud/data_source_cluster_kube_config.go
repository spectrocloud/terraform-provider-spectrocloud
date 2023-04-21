package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceClusterKubeConfig() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceClusterKubeConfigRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"kube_config": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"project", "tenant"}, false),
				Description:  "Cluster context can be 'project' or 'tenant'. Defaults to 'project'.",
			},
		},
	}
}

func dataSourceClusterKubeConfigRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	var diags diag.Diagnostics
	ClusterContext := d.Get("context").(string)
	if name, okName := d.GetOk("name"); okName {
		cluster, err := c.GetClusterByName(name.(string), ClusterContext)
		if err != nil {
			return diag.FromErr(err)
		}
		if cluster != nil {
			kubeConfig, _ := c.GetClusterKubeConfig(cluster.Metadata.UID)
			d.SetId(cluster.Metadata.UID)
			if err := d.Set("kube_config", kubeConfig); err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set("name", cluster.Metadata.Name); err != nil {
				return diag.FromErr(err)
			}
		}
	}
	return diags
}

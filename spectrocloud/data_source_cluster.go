package spectrocloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceCluster() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceClusterRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"admin_kube_config": {
				Type:     schema.TypeString,
				Computed: true,
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
				Description: "The context of the cluster. Allowed values are `project` or `tenant`. " +
					"Defaults to `project`." + PROJECT_NAME_NUANCE,
			},
			"virtual": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "If set to true, the cluster will treated as a virtual cluster. Defaults to `false`.",
			},
		},
	}
}

func dataSourceClusterRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	var diags diag.Diagnostics
	if name, okName := d.GetOk("name"); okName {
		ClusterContext := d.Get("context").(string)
		cluster, err := c.GetClusterByName(name.(string), ClusterContext, d.Get("virtual").(bool))
		if err != nil {
			return diag.FromErr(err)
		}
		if cluster != nil {
			d.SetId(cluster.Metadata.UID)
			kubeConfig, _ := c.GetClusterKubeConfig(cluster.Metadata.UID, ClusterContext)
			if err := d.Set("kube_config", kubeConfig); err != nil {
				return diag.FromErr(err)
			}
			adminKubeConfig, _ := c.GetClusterAdminKubeConfig(cluster.Metadata.UID, ClusterContext)
			if adminKubeConfig != "" {
				if err := d.Set("admin_kube_config", adminKubeConfig); err != nil {
					return diag.FromErr(err)
				}
			}
			d.SetId(cluster.Metadata.UID)
			if err := d.Set("name", cluster.Metadata.Name); err != nil {
				return diag.FromErr(err)
			}
		}
	}
	return diags
}

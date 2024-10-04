package spectrocloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceCluster() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceClusterRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the cluster.",
			},
			"admin_kube_config": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The admin kubeconfig file for accessing the cluster. This is computed automatically.",
			},
			"kube_config": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The kubeconfig file for accessing the cluster as a non-admin user. This is computed automatically.",
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
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)
	var diags diag.Diagnostics
	if name, okName := d.GetOk("name"); okName {

		cluster, err := c.GetClusterByName(name.(string), d.Get("virtual").(bool))
		if err != nil {
			return diag.FromErr(err)
		}
		if cluster != nil {
			d.SetId(cluster.Metadata.UID)
			kubeConfig, _ := c.GetClusterKubeConfig(cluster.Metadata.UID)
			if err := d.Set("kube_config", kubeConfig); err != nil {
				return diag.FromErr(err)
			}
			adminKubeConfig, _ := c.GetClusterAdminKubeConfig(cluster.Metadata.UID)
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

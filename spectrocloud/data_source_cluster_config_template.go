package spectrocloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceClusterConfigTemplate() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceClusterConfigTemplateRead,
		Description: "Data source for retrieving information about a cluster config template.",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the cluster config template.",
			},
			"cloud_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The cloud type for the cluster template.",
			},
			"profiles": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of cluster profile references.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"uid": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "UID of the cluster profile.",
						},
					},
				},
			},
			"policies": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of policy references.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"uid": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "UID of the policy.",
						},
						"kind": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Kind of the policy.",
						},
					},
				},
			},
		},
	}
}

func dataSourceClusterConfigTemplateRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "")
	var diags diag.Diagnostics

	name := d.Get("name").(string)
	templateSummary, err := c.GetClusterConfigTemplateByName(name)
	if err != nil {
		return handleReadError(d, err, diags)
	}

	template, err := c.GetClusterConfigTemplate(templateSummary.Metadata.UID)
	if err != nil {
		return handleReadError(d, err, diags)
	}

	d.SetId(template.Metadata.UID)

	if err := d.Set("name", template.Metadata.Name); err != nil {
		return diag.FromErr(err)
	}

	if template.Spec != nil {
		if err := d.Set("cloud_type", template.Spec.CloudType); err != nil {
			return diag.FromErr(err)
		}

		if err := d.Set("profiles", flattenClusterTemplateProfiles(template.Spec.Profiles)); err != nil {
			return diag.FromErr(err)
		}

		if err := d.Set("policies", flattenClusterTemplatePolicies(template.Spec.Policies)); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

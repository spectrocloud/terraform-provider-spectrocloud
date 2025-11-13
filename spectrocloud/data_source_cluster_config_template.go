package spectrocloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceClusterConfigTemplate() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceClusterConfigTemplateRead,
		Description: "Data source for retrieving information about a cluster config template. **Tech Preview**: This data source is in tech preview and may undergo changes.",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the cluster config template.",
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"project", "tenant"}, false),
				Description: "The context of the cluster config template. Allowed values are `project` or `tenant`. " +
					"Default value is `project`. " + PROJECT_NAME_NUANCE,
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The description of the cluster config template.",
			},
			"tags": {
				Type:     schema.TypeSet,
				Computed: true,
				Set:      schema.HashString,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Tags assigned to the cluster config template.",
			},
			"cloud_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The cloud type for the cluster template.",
			},
			"profiles": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Set of cluster profile references.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"uid": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "UID of the cluster profile.",
						},
						"variables": {
							Type:        schema.TypeSet,
							Computed:    true,
							Description: "Set of profile variable values and assignment strategies.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Name of the variable.",
									},
									"value": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Value of the variable to be applied to all clusters launched from this template.",
									},
									"assign_strategy": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Assignment strategy for the variable. Possible values: `all` or `cluster`.",
									},
								},
							},
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
			"attached_cluster": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of clusters attached to this template.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cluster_uid": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "UID of the attached cluster.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the attached cluster.",
						},
					},
				},
			},
			"execution_state": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Current execution state of the cluster template. Possible values: `Pending`, `Applied`, `Failed`, `PartiallyApplied`.",
			},
		},
	}
}

func dataSourceClusterConfigTemplateRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, d.Get("context").(string))
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

	if err := d.Set("tags", flattenTags(template.Metadata.Labels)); err != nil {
		return diag.FromErr(err)
	}

	// Get description from annotations if it exists
	if template.Metadata.Annotations != nil {
		if description, found := template.Metadata.Annotations["description"]; found {
			if err := d.Set("description", description); err != nil {
				return diag.FromErr(err)
			}
		}
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

		// Set attached clusters
		if err := d.Set("attached_cluster", flattenAttachedClusters(template.Spec.Clusters)); err != nil {
			return diag.FromErr(err)
		}
	}

	// Set execution state from status
	if template.Status != nil {
		if err := d.Set("execution_state", template.Status.State); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

package spectrocloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceFilter() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFilterRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the filter.",
			},
			"metadata": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the filter.",
						},
						"annotations": {
							Type:        schema.TypeMap,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Annotations related to the filter, represented as key-value pairs.",
						},
						"labels": {
							Type:        schema.TypeMap,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Labels associated with the filter, represented as key-value pairs.",
						},
					},
				},
			},
			"spec": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"filter_group": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"conjunction": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The conjunction used to combine filter groups. Common values: `AND`, `OR`.",
									},
									"filters": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"key": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "The key for the filter condition.",
												},
												"negation": {
													Type:        schema.TypeBool,
													Computed:    true,
													Description: "A flag indicating whether the filter is negated.",
												},
												"operator": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "The operator used in the filter condition. Examples: `=`, `!=`.",
												},
												"values": {
													Type:        schema.TypeList,
													Computed:    true,
													Elem:        &schema.Schema{Type: schema.TypeString},
													Description: "A list of values to compare against in the filter condition.",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceFilterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "")
	var diags diag.Diagnostics

	name := d.Get("name").(string)
	filter, err := c.GetTagFilterByName(name)
	if err != nil {
		return handleReadError(d, err, diags)
	}

	FilterSummary, err := c.GetTagFilter(filter.Metadata.UID)
	if err != nil {
		return handleReadError(d, err, diags)
	}

	d.SetId(FilterSummary.Metadata.UID)
	err = d.Set("metadata", []interface{}{map[string]interface{}{
		"name":        FilterSummary.Metadata.Name,
		"annotations": FilterSummary.Metadata.Annotations,
		"labels":      FilterSummary.Metadata.Labels,
	}})
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("spec", []interface{}{map[string]interface{}{
		"filter_group": flattenFilterGroup(FilterSummary.Spec.FilterGroup),
	}})
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

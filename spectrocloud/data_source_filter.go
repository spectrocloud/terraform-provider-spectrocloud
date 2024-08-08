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
							Type:     schema.TypeString,
							Computed: true,
						},
						"annotations": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"labels": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
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
										Type:     schema.TypeString,
										Computed: true,
									},
									"filters": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"key": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"negation": {
													Type:     schema.TypeBool,
													Computed: true,
												},
												"operator": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"values": {
													Type:     schema.TypeList,
													Computed: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
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
		return diag.FromErr(err)
	}

	FilterSummary, err := c.GetTagFilter(filter.Metadata.UID)
	if err != nil {
		return diag.FromErr(err)
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

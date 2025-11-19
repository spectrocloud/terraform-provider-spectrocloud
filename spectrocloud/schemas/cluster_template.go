package schemas

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func ClusterTemplateSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		MaxItems:    1,
		Description: "The cluster template of the cluster.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"id": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The ID of the cluster template.",
				},
				"name": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The name of the cluster template.",
				},
				"cluster_profile": {
					Type:        schema.TypeSet,
					Optional:    true,
					Description: "The cluster profile of the cluster template.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"id": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "The UID of the cluster profile.",
							},
							"variables": {
								Type:        schema.TypeMap,
								Optional:    true,
								Description: "A map of cluster profile variables, specified as key-value pairs. For example: `priority = \"5\"`.",
								Elem: &schema.Schema{
									Type: schema.TypeString,
								},
							},
						},
					},
				},
			},
		},
	}
}

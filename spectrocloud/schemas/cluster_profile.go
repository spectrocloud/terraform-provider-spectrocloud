package schemas

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ClusterProfileSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"id": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The ID of the cluster profile.",
				},
				"pack": PackSchema(),
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
	}
}

func ClusterProfileSchemaV2() *schema.Schema {
	elem := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the cluster profile.",
			},
			"pack": PackSchema(),
			"variables": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "A map of cluster profile variables, specified as key-value pairs. For example: `priority = \"5\"`.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}

	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		// Hash on the full object to ensure changes to any nested attribute
		// (id/pack/variables) are treated as element changes in the set.
		Set:  schema.HashResource(elem),
		Elem: elem,
	}
}

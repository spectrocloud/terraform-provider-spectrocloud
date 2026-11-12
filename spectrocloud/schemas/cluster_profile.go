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
			// Computed-only, so it is excluded from the set hash. It also makes a changed element's
			// planned value unknown, which stops terraform-plugin-sdk from overwriting the
			// provider's rollback with the planned value when an update fails (PLT-2508).
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the cluster profile, read from Palette.",
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

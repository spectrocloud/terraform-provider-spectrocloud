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
					Description: "A map of cluster profile variables, defined as key-value pairs. Example: `priority: 5`.",
				},
			},
		},
	}
}

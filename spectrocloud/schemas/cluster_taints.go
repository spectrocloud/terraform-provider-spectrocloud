package schemas

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func ClusterTaintsSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"key": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The key of the taint.",
				},
				"value": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The value of the taint.",
				},
				"effect": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The effect of the taint.",
				},
			},
		},
	}
}

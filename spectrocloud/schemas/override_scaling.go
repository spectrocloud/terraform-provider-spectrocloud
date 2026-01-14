package schemas

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func OverrideScalingSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		MaxItems:    1,
		Description: "Rolling update strategy for the machine pool.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"max_surge": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: "Max extra nodes during rolling update. Integer or percentage (e.g., '1' or '20%'). Only valid when type=OverrideScaling. Both maxSurge and maxUnavailable are required.",
				},
				"max_unavailable": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: "Max unavailable nodes during rolling update. Integer or percentage (e.g., '0' or '10%'). Only valid when type=OverrideScaling. Both maxSurge and maxUnavailable are required.",
				},
			},
		},
	}
}

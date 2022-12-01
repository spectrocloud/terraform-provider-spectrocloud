package schemas

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func ClusterLocationSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"country_code": {
					Type:     schema.TypeString,
					Optional: true,
					Default:  "",
				},
				"country_name": {
					Type:     schema.TypeString,
					Description: "The name of the country.",
					Optional: true,
					Default:  "",
				},
				"region_code": {
					Type:     schema.TypeString,
					Optional: true,
					Default:  "",
				},
				"region_name": {
					Type:     schema.TypeString,
					Optional: true,
					Default:  "",
				},
				"latitude": {
					Type:     schema.TypeFloat,
					Required: true,
				},
				"longitude": {
					Type:     schema.TypeFloat,
					Required: true,
				},
			},
		},
	}
}

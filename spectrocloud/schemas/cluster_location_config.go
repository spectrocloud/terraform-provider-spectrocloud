package schemas

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func ClusterLocationSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"country_code": {
					Type:        schema.TypeString,
					Description: "The country code of the country the cluster is located in.",
					Optional:    true,
					Default:     "",
				},
				"country_name": {
					Type:        schema.TypeString,
					Description: "The name of the country.",
					Optional:    true,
					Default:     "",
				},
				"region_code": {
					Type:        schema.TypeString,
					Description: "The region code of where the cluster is located in.",
					Optional:    true,
					Default:     "",
				},
				"region_name": {
					Type:        schema.TypeString,
					Description: "The name of the region.",
					Optional:    true,
					Default:     "",
				},
				"latitude": {
					Type:        schema.TypeFloat,
					Description: "The latitude coordinates value.",
					Required:    true,
				},
				"longitude": {
					Type:        schema.TypeFloat,
					Required:    true,
					Description: "The longitude coordinates value.",
				},
			},
		},
	}
}

func ClusterLocationSchemaComputed() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"country_code": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"country_name": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"region_code": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"region_name": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"latitude": {
					Type:     schema.TypeFloat,
					Computed: true,
				},
				"longitude": {
					Type:     schema.TypeFloat,
					Computed: true,
				},
			},
		},
	}
}

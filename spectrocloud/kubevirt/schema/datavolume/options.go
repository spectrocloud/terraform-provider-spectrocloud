package datavolume

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataVolumeOptionsSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "DataVolumeSpec defines our specification for a DataVolume type",
		Required:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:     schema.TypeString,
					Required: true,
				},
				"disk": {
					Type:     schema.TypeList,
					Required: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"name": {
								Type:     schema.TypeString,
								Required: true,
							},
							"bus": {
								Type:     schema.TypeString,
								Required: true,
							},
						},
					},
				},
				"volume_source": {
					Type:     schema.TypeList,
					Required: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"data_volume": {
								Type:     schema.TypeList,
								Required: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"name": {
											Type:     schema.TypeString,
											Required: true,
										},
										"hotpluggable": {
											Type:     schema.TypeBool,
											Optional: true,
											Default:  true,
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

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
					Type:        schema.TypeString,
					Required:    true,
					Description: "Name of the volume attachment in the virtual machine spec.",
				},
				"disk": {
					Type:     schema.TypeList,
					Required: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"name": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Name of the disk definition in the virtual machine spec.",
							},
							"bus": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Disk bus type used by the attached data volume (for example, `virtio` or `sata`).",
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
											Type:        schema.TypeString,
											Required:    true,
											Description: "Name of the data volume source referenced by this attachment.",
										},
										"hotpluggable": {
											Type:        schema.TypeBool,
											Optional:    true,
											Default:     true,
											Description: "Whether this data volume can be hot-plugged while the virtual machine is running.",
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

package schemas

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func VMDeviceSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"disk": {
					Type:     schema.TypeList,
					Required: true,
					Elem:     VMDiskSchema(),
				},
				"interface": {
					Type:     schema.TypeList,
					Required: true,
					Elem:     VMInterfaceSchema(),
				},
			},
		},
	}
}

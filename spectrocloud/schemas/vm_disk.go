package schemas

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func VMDiskSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The name of the disk. This is the name that will be used to identify the disk in the guest OS.",
				},
				"bus": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The bus type of the disk. This is the name that will be used to identify the disk in the guest OS.",
				},
			},
		},
	}
}

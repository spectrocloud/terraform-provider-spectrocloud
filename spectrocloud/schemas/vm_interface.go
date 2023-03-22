package schemas

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func VMInterfaceSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The name of the interface. This is the name that will be used to identify the device interface in the guest OS.",
				},
				"type": {
					Type:         schema.TypeString,
					Optional:     true,
					Default:      "masquerade",
					ValidateFunc: validation.StringInSlice([]string{"masquerade", "bridge", "macvtap"}, false),
					Description:  "The type of the interface. Can be one of `masquerade`, `bridge`, or `macvtap`. Defaults to `masquerade`.",
				},
				"model": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "virtio",
					Description: "The model of the interface.",
				},
			},
		},
	}
}

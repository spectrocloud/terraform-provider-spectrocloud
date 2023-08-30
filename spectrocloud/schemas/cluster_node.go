package schemas

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func NodeSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"node_id": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The node_id of the node.",
				},
				"action": {
					Type:         schema.TypeString,
					Required:     true,
					Description:  "The action to perform on the node. Valid values are: `cordon`, `uncordon`.",
					ValidateFunc: validation.StringInSlice([]string{"cordon", "uncordon"}, false),
				},
			},
		},
	}
}

package schemas

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func MachinePoolArchTypeSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		Default:      "amd64",
		Description:  "Architecture type of the machine pool. Allowed values are `amd64` and `arm64`. Default is `amd64`.",
		ValidateFunc: validation.StringInSlice([]string{"amd64", "arm64"}, false),
	}
}

func NodeSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"node_id": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The node_id of the node, For example `i-07f899a33dee624f7`",
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

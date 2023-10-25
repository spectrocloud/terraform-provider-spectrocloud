package schemas

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SubnetSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "Name of the subnet.",
				},
				"cidr_block": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "CidrBlock is the CIDR block to be used when the provider creates a managed virtual network.",
				},
				"security_group_name": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Network Security Group(NSG) to be attached to subnet.",
				},
			},
		},
	}
}

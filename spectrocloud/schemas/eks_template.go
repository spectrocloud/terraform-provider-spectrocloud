package schemas

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func EksLaunchTemplate() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"ami_id": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The ID of the custom Amazon Machine Image (AMI).",
				},
				"root_volume_type": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The type of the root volume.",
				},
				"root_volume_iops": {
					Type:        schema.TypeInt,
					Optional:    true,
					Description: "The number of input/output operations per second (IOPS) for the root volume.",
				},
				"root_volume_throughput": {
					Type:        schema.TypeInt,
					Optional:    true,
					Description: "The throughput of the root volume in MiB/s.",
				},
			},
		},
	}
}

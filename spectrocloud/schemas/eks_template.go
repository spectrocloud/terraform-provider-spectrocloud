package schemas

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func AwsLaunchTemplate() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"ami_id": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The ID of the custom Amazon Machine Image (AMI). If you do not set an `ami_id`, Palette will repave the cluster when it automatically updates the EKS AMI.",
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
				"additional_security_groups": {
					Type: schema.TypeSet,
					Set:  schema.HashString,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
					Optional:    true,
					Description: "Additional security groups to attach to the instance.",
				},
			},
		},
	}
}

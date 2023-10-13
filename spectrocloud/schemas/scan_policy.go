package schemas

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func ScanPolicySchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		MaxItems:    1,
		Description: "The scan policy for the cluster.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"configuration_scan_schedule": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The schedule for configuration scan.",
				},
				"penetration_scan_schedule": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The schedule for penetration scan.",
				},
				"conformance_scan_schedule": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The schedule for conformance scan.",
				},
			},
		},
	}
}

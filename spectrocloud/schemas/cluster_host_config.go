package schemas

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ClusterHostConfigSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		Description: "The host configuration for the cluster.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"host_endpoint_type": {
					Type:         schema.TypeString,
					Optional:     true,
					Default:      "Ingress",
					ValidateFunc: validation.StringInSlice([]string{"Ingress", "LoadBalancer"}, false),
					Description:  "The type of endpoint for the cluster. Can be either 'Ingress' or 'LoadBalancer'. The default is 'Ingress'.",
				},
				"ingress_host": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The host for the Ingress endpoint. Required if 'host_endpoint_type' is set to 'Ingress'.",
				},
				"external_traffic_policy": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The external traffic policy for the cluster.",
				},
				"load_balancer_source_ranges": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The source ranges for the load balancer. Required if 'host_endpoint_type' is set to 'LoadBalancer'.",
				},
			},
		},
	}
}

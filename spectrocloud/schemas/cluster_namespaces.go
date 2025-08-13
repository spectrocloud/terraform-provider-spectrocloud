package schemas

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func ClusterNamespacesSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		Description: "The namespaces for the cluster.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "Name of the namespace. This is the name of the Kubernetes namespace in the cluster.",
				},
				"resource_allocation": {
					Type:     schema.TypeMap,
					Required: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
					Description: "Resource allocation for the namespace. This is a map containing the resource type and the resource value. For example, `{cpu_cores: '2', memory_MiB: '2048', gpu_limit: '1', gpu_provider: 'nvidia'}`",
				},
			},
		},
	}
}

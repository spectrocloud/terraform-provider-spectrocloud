package schemas

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func WorkspaceNamespacesSchema() *schema.Schema {
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
				"cluster_resource_allocations": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"uid": {
								Type:     schema.TypeString,
								Required: true,
							},
							"resource_allocation": {
								Type:     schema.TypeMap,
								Required: true,
								Elem: &schema.Schema{
									Type: schema.TypeString,
								},
								Description: "Resource allocation for the cluster. This is a map containing the resource type and the resource value. For example, `{cpu_cores: '2', memory_MiB: '2048', gpu_limit: '1'}`. Note: gpu_provider is not supported here; use the default resource_allocation for GPU provider configuration.",
							},
						},
					},
				},
				"images_blacklist": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
					Description: "List of images to disallow for the namespace. For example, `['nginx:latest', 'redis:latest']`",
				},
			},
		},
	}
}

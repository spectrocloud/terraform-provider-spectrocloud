package schemas

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func ClusterTemplateSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Description: "The cluster template of the cluster. If the cluster was created with `cluster_profile` and this " +
			"is populated in a later apply (with `cluster_profile` removed), the cluster is attached to the named " +
			"template (Day 2 attach) - a single API call that binds the cluster to the template; the target profile " +
			"set is applied by the template's batch reconciler at the next maintenance window, not synchronously. " +
			"This is currently one-way: there is no supported way to detach a cluster from a template and revert to " +
			"`cluster_profile`, so removing `cluster_template` after it has been set is rejected.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"id": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The ID of the cluster template.",
				},
				"name": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The name of the cluster template.",
				},
				"cluster_profile": {
					Type:        schema.TypeSet,
					Optional:    true,
					Description: "The cluster profile of the cluster template.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"id": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "The UID of the cluster profile.",
							},
							"variables": {
								Type:        schema.TypeMap,
								Optional:    true,
								Description: "A map of cluster profile variables, specified as key-value pairs. For example: `priority = \"5\"`.",
								Elem: &schema.Schema{
									Type: schema.TypeString,
								},
							},
						},
					},
				},
			},
		},
	}
}

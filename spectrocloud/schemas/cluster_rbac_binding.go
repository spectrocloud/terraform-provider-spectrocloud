package schemas

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ClusterRbacBindingSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"type": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringInSlice([]string{"RoleBinding", "ClusterRoleBinding"}, false),
					Description:  "The type of the RBAC binding. Can be one of the following values: `RoleBinding`, or `ClusterRoleBinding`.",
				},
				"namespace": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The namespace of the RBAC binding. Required if 'type' is set to 'RoleBinding'.",
				},
				"role": {
					Type:     schema.TypeMap,
					Optional: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
					Description: "The role of the RBAC binding. Required if 'type' is set to 'RoleBinding'.",
				},
				"subjects": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"type": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.StringInSlice([]string{"User", "Group", "ServiceAccount"}, false),
								Description:  "The type of the subject. Can be one of the following values: `User`, `Group`, or `ServiceAccount`.",
							},
							"name": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "The name of the subject. Required if 'type' is set to 'User' or 'Group'.",
							},
							"namespace": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "The Kubernetes namespace of the subject. Required if 'type' is set to 'ServiceAccount'.",
							},
						},
					},
				},
			},
		},
	}
}

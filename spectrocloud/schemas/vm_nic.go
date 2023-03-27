package schemas

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func VMNicSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"nic": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"name": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "The name of the network interface.",
							},
							"multus": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"network_name": {
											Type:        schema.TypeString,
											Required:    true,
											Description: "The name of the network attachment definition.",
										},
										"default": {
											Type:        schema.TypeBool,
											Optional:    true,
											Description: "Set this network as the default one for the pod.",
										},
									},
								},
								Description: "The multus configuration for the network interface.",
							},
							"network_type": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "The name of the network to be attached to the virtual machine.",
							},
						},
					},
					Description: "The network specification for the virtual machine.",
				},
			},
		},
	}
}

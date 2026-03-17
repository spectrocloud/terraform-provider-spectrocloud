package virtualmachineinstance

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

func networkFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Description: "Network name.",
			Required:    true,
		},
		"network_source": {
			Type:        schema.TypeList,
			Description: "NetworkSource represents the network type and the source interface that should be connected to the virtual machine.",
			MaxItems:    1,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"pod": {
						Type:        schema.TypeList,
						Description: "Pod network.",
						MaxItems:    1,
						Optional:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"vm_network_cidr": {
									Type:        schema.TypeString,
									Description: "CIDR for vm network.",
									Optional:    true,
								},
							},
						},
					},
					"multus": {
						Type:        schema.TypeList,
						Description: "Multus network.",
						MaxItems:    1,
						Optional:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"network_name": {
									Type:        schema.TypeString,
									Description: "References to a NetworkAttachmentDefinition CRD object. Format: <networkName>, <namespace>/<networkName>. If namespace is not specified, VMI namespace is assumed.",
									Required:    true,
								},
								"default": {
									Type:        schema.TypeBool,
									Description: "Select the default network and add it to the multus-cni.io/default-network annotation.",
									Optional:    true,
								},
							},
						},
					},
				},
			},
		},
	}
}

func NetworksSchema() *schema.Schema {
	fields := networkFields()

	return &schema.Schema{
		Type: schema.TypeList,

		Description: "List of networks that can be attached to a vm's virtual interface.",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: fields,
		},
	}
}

// expandNetworksToVM expands the network schema into []*models.V1VMNetwork for VM spec.
func expandNetworksToVM(networks []interface{}) []*models.V1VMNetwork {
	result := make([]*models.V1VMNetwork, len(networks))
	if len(networks) == 0 || networks[0] == nil {
		return result
	}
	for i, network := range networks {
		in := network.(map[string]interface{})
		item := &models.V1VMNetwork{}
		if v, ok := in["name"].(string); ok {
			item.Name = types.Ptr(v)
		}
		if v, ok := in["network_source"].([]interface{}); ok && len(v) > 0 {
			if src, ok := v[0].(map[string]interface{}); ok {
				if p, ok := src["pod"].([]interface{}); ok && len(p) == 1 {
					item.Pod = expandPodNetworkToVM(p)
				}
				if m, ok := src["multus"].([]interface{}); ok && len(m) == 1 {
					item.Multus = expandMultusNetworkToVM(m)
				}
			}
		}
		result[i] = item
	}
	return result
}

func expandPodNetworkToVM(pod []interface{}) *models.V1VMPodNetwork {
	if len(pod) == 0 || pod[0] == nil {
		return nil
	}
	result := &models.V1VMPodNetwork{}
	in := pod[0].(map[string]interface{})
	if v, ok := in["vm_network_cidr"].(string); ok {
		result.VMNetworkCIDR = v
	}
	if v, ok := in["vm_ipv6_network_cidr"].(string); ok {
		result.VMIPV6NetworkCIDR = v
	}
	return result
}

func expandMultusNetworkToVM(multus []interface{}) *models.V1VMMultusNetwork {
	if len(multus) == 0 || multus[0] == nil {
		return nil
	}
	result := &models.V1VMMultusNetwork{}
	in := multus[0].(map[string]interface{})
	if v, ok := in["network_name"].(string); ok {
		result.NetworkName = types.Ptr(v)
	}
	if v, ok := in["default"].(bool); ok {
		result.Default = v
	}
	return result
}

// func flattenNetworks(in []kubevirtapiv1.Network) []interface{} {
// 	att := make([]interface{}, len(in))

// 	for i, v := range in {
// 		c := make(map[string]interface{})

// 		c["name"] = v.Name
// 		c["network_source"] = flattenNetworkSource(v.NetworkSource)

// 		att[i] = c
// 	}

// 	return att
// }

// flattenNetworksFromVM flattens []*models.V1VMNetwork to the same shape as flattenNetworks.
func flattenNetworksFromVM(in []*models.V1VMNetwork) []interface{} {
	if len(in) == 0 {
		return nil
	}
	att := make([]interface{}, len(in))
	for i, v := range in {
		if v == nil {
			continue
		}
		c := make(map[string]interface{})
		if v.Name != nil {
			c["name"] = *v.Name
		}
		c["network_source"] = flattenNetworkSourceFromVM(v)
		att[i] = c
	}
	return att
}

func flattenNetworkSourceFromVM(in *models.V1VMNetwork) []interface{} {
	if in == nil {
		return []interface{}{map[string]interface{}{}}
	}
	att := make(map[string]interface{})
	if in.Pod != nil {
		att["pod"] = []interface{}{map[string]interface{}{"vm_network_cidr": in.Pod.VMNetworkCIDR, "vm_ipv6_network_cidr": in.Pod.VMIPV6NetworkCIDR}}
	}
	if in.Multus != nil {
		networkName := ""
		if in.Multus.NetworkName != nil {
			networkName = *in.Multus.NetworkName
		}
		att["multus"] = []interface{}{map[string]interface{}{"network_name": networkName, "default": in.Multus.Default}}
	}
	return []interface{}{att}
}

// func flattenNetworkSource(in kubevirtapiv1.NetworkSource) []interface{} {
// 	att := make(map[string]interface{})

// 	if in.Pod != nil {
// 		att["pod"] = flattenPodNetwork(*in.Pod)
// 	}
// 	if in.Multus != nil {
// 		att["multus"] = flattenMultusNetwork(*in.Multus)
// 	}

// 	return []interface{}{att}
// }

// func flattenPodNetwork(in kubevirtapiv1.PodNetwork) []interface{} {
// 	att := make(map[string]interface{})

// 	att["vm_network_cidr"] = in.VMNetworkCIDR

// 	return []interface{}{att}
// }

// func flattenMultusNetwork(in kubevirtapiv1.MultusNetwork) []interface{} {
// 	att := make(map[string]interface{})

// 	att["network_name"] = in.NetworkName
// 	att["default"] = in.Default

// 	return []interface{}{att}
// }

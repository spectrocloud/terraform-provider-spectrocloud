package virtualmachineinstance

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spectrocloud/palette-sdk-go/api/models"
)

func probeFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		// TODO nargaman
	}
}

func ProbeSchema() *schema.Schema {
	fields := probeFields()

	return &schema.Schema{
		Type: schema.TypeList,

		Description: "Specification of the desired behavior of the VirtualMachineInstance on the host.",
		Optional:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: fields,
		},
	}
}

// func expandProbe(probe []interface{}) *kubevirtapiv1.Probe {
// 	if len(probe) == 0 || probe[0] == nil {
// 		return nil
// 	}

// 	result := &kubevirtapiv1.Probe{}

// 	_ = probe[0].(map[string]interface{})

// 	// TODO nargaman

// 	return result
// }

// expandProbeToVM expands the probe schema into Palette models.V1VMProbe for VM spec.
func expandProbeToVM(probe []interface{}) *models.V1VMProbe {
	if len(probe) == 0 || probe[0] == nil {
		return nil
	}
	_ = probe[0].(map[string]interface{})
	// TODO: populate V1VMProbe fields when schema is defined
	return &models.V1VMProbe{}
}

// func flattenProbe(in kubevirtapiv1.Probe) []interface{} {
// 	att := make(map[string]interface{})

// 	// att["spec"] = flattenVirtualMachineInstanceSpecSpec(in.Spec)
// 	// att["status"] = flattenVirtualMachineInstanceSpecStatus(in.Status)
// 	// TODO nargaman

// 	return []interface{}{att}
// }

// flattenProbeFromVM flattens *models.V1VMProbe to the same shape as flattenProbe.
func flattenProbeFromVM(in *models.V1VMProbe) []interface{} {
	if in == nil {
		return []interface{}{map[string]interface{}{}}
	}
	att := make(map[string]interface{})
	return []interface{}{att}
}

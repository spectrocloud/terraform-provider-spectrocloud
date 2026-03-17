package virtualmachine

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/utils"
)

func virtualMachineConditionsFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"type": {
			Type:        schema.TypeString,
			Description: "VirtualMachineConditionType represent the type of the VM as concluded from its VMi status.",
			Optional:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"Failure",
				"Ready",
				"Paused",
				"RenameOperation",
			}, false),
		},
		"status": {
			Type:        schema.TypeString,
			Description: "ConditionStatus represents the status of this VM condition, if the VM currently in the condition.",
			Optional:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"True",
				"False",
				"Unknown",
			}, false),
		},
		// TODO nargaman -  Add following values
		// "last_probe_time": {
		// 	Type:        schema.TypeString,
		// 	Description: "Last probe time.",
		// 	Optional:    true,
		// },
		// "last_transition_time": {
		// 	Type:        schema.TypeString,
		// 	Description: "Last transition time.",
		// 	Optional:    true,
		// },
		"reason": {
			Type:        schema.TypeString,
			Description: "Condition reason.",
			Optional:    true,
		},
		"message": {
			Type:        schema.TypeString,
			Description: "Condition message.",
			Optional:    true,
		},
	}
}

func virtualMachineConditionsSchema() *schema.Schema {
	fields := virtualMachineConditionsFields()

	return &schema.Schema{
		Type: schema.TypeList,

		Description: "Hold the state information of the VirtualMachine and its VirtualMachineInstance.",
		Required:    true,
		Elem: &schema.Resource{
			Schema: fields,
		},
	}
}

func expandVirtualMachineConditions(conditions []interface{}) ([]*models.V1VMVirtualMachineCondition, error) {
	if len(conditions) == 0 || conditions[0] == nil {
		return []*models.V1VMVirtualMachineCondition{}, nil
	}

	result := make([]*models.V1VMVirtualMachineCondition, len(conditions))
	for i, condition := range conditions {
		in := condition.(map[string]interface{})
		cond := &models.V1VMVirtualMachineCondition{}

		if v, ok := in["type"].(string); ok {
			cond.Type = utils.PtrToString(v)
		}
		if v, ok := in["status"].(string); ok {
			cond.Status = utils.PtrToString(v)
		}
		if v, ok := in["reason"].(string); ok {
			cond.Reason = v
		}
		if v, ok := in["message"].(string); ok {
			cond.Message = v
		}

		result[i] = cond
	}

	return result, nil
}

// func flattenVirtualMachineConditions(in []kubevirtapiv1.VirtualMachineCondition) []interface{} {
// 	att := make([]interface{}, len(in))

// 	for i, v := range in {
// 		c := make(map[string]interface{})
// 		c["type"] = string(v.Type)
// 		c["status"] = string(v.Status)
// 		c["reason"] = v.Reason
// 		c["message"] = v.Message

// 		att[i] = c
// 	}

// 	return att
// }

func flattenVirtualMachineConditionsFromVM(in []*models.V1VMVirtualMachineCondition) []interface{} {
	if len(in) == 0 {
		return nil
	}
	att := make([]interface{}, 0, len(in))
	for _, v := range in {
		if v == nil {
			continue
		}
		c := make(map[string]interface{})
		if v.Type != nil {
			c["type"] = *v.Type
		}
		if v.Status != nil {
			c["status"] = *v.Status
		}
		c["reason"] = v.Reason
		c["message"] = v.Message
		att = append(att, c)
	}
	return att
}

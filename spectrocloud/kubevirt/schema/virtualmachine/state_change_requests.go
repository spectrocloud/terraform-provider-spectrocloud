package virtualmachine

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/utils"
)

func virtualMachineStateChangeRequestFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"action": {
			Type:        schema.TypeString,
			Description: "Indicates the type of action that is requested. e.g. Start or Stop.",
			Optional:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"Start",
				"Stop",
			}, false),
		},
		"data": {
			Type:        schema.TypeMap,
			Description: "Provides additional data in order to perform the Action.",
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"uid": {
			Type:        schema.TypeString,
			Description: "Indicates the UUID of an existing Virtual Machine Instance that this change request applies to -- if applicable.",
			Optional:    true,
		},
	}
}

func virtualMachineStateChangeRequestsSchema() *schema.Schema {
	fields := virtualMachineStateChangeRequestFields()

	return &schema.Schema{
		Type: schema.TypeList,

		Description: "StateChangeRequests indicates a list of actions that should be taken on a VMI.",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: fields,
		},
	}
}

func expandVirtualMachineStateChangeRequests(virtualMachineStateChangeRequests []interface{}) []*models.V1VMVirtualMachineStateChangeRequest {
	if len(virtualMachineStateChangeRequests) == 0 || virtualMachineStateChangeRequests[0] == nil {
		return []*models.V1VMVirtualMachineStateChangeRequest{}
	}

	result := make([]*models.V1VMVirtualMachineStateChangeRequest, len(virtualMachineStateChangeRequests))
	for i, virtualMachineStateChangeRequest := range virtualMachineStateChangeRequests {
		in := virtualMachineStateChangeRequest.(map[string]interface{})
		req := &models.V1VMVirtualMachineStateChangeRequest{}

		if v, ok := in["action"].(string); ok {
			req.Action = utils.PtrToString(v)
		}
		if v, ok := in["data"].(map[string]interface{}); ok && len(v) > 0 {
			req.Data = utils.ExpandStringMap(v)
		}
		if v, ok := in["uid"].(string); ok {
			req.UID = v
		}

		result[i] = req
	}

	return result
}

func flattenVirtualMachineStateChangeRequestsFromVM(in []*models.V1VMVirtualMachineStateChangeRequest) []interface{} {
	if len(in) == 0 {
		return nil
	}
	att := make([]interface{}, 0, len(in))
	for _, v := range in {
		if v == nil {
			continue
		}
		c := make(map[string]interface{})
		if v.Action != nil {
			c["action"] = *v.Action
		}
		c["data"] = v.Data
		if v.UID != "" {
			c["uid"] = v.UID
		}
		att = append(att, c)
	}
	return att
}

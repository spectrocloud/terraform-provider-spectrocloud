package convert

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

// SchemaToHapiVmStatus converts Terraform schema status fields to HAPI VM Status
func SchemaToHapiVmStatus(d *schema.ResourceData) (*models.V1ClusterVirtualMachineStatus, error) {
	status := &models.V1ClusterVirtualMachineStatus{}

	if v, ok := d.GetOk("status"); ok {
		statusList := v.([]interface{})
		if len(statusList) > 0 && statusList[0] != nil {
			statusMap := statusList[0].(map[string]interface{})

			// Created
			if v, ok := statusMap["created"].(bool); ok {
				status.Created = v
			}

			// Ready
			if v, ok := statusMap["ready"].(bool); ok {
				status.Ready = v
			}

			// Conditions
			if v, ok := statusMap["conditions"].([]interface{}); ok {
				conditions, err := SchemaToHapiConditions(v)
				if err != nil {
					return nil, fmt.Errorf("failed to convert conditions: %w", err)
				}
				status.Conditions = conditions
			}

			// StateChangeRequests
			if v, ok := statusMap["state_change_requests"].([]interface{}); ok {
				stateChangeRequests, err := SchemaToHapiStateChangeRequests(v)
				if err != nil {
					return nil, fmt.Errorf("failed to convert state change requests: %w", err)
				}
				status.StateChangeRequests = stateChangeRequests
			}
		}
	}

	return status, nil
}

// SchemaToHapiConditions converts Terraform schema conditions to HAPI Conditions
func SchemaToHapiConditions(conditions []interface{}) ([]*models.V1VMVirtualMachineCondition, error) {
	if len(conditions) == 0 {
		return nil, nil
	}

	result := make([]*models.V1VMVirtualMachineCondition, len(conditions))

	for i, condition := range conditions {
		if condition == nil {
			continue
		}

		condMap := condition.(map[string]interface{})
		cond := &models.V1VMVirtualMachineCondition{}

		if v, ok := condMap["type"].(string); ok {
			cond.Type = &v
		}

		if v, ok := condMap["status"].(string); ok {
			cond.Status = &v
		}

		if v, ok := condMap["reason"].(string); ok {
			cond.Reason = v
		}

		if v, ok := condMap["message"].(string); ok {
			cond.Message = v
		}

		result[i] = cond
	}

	return result, nil
}

// SchemaToHapiStateChangeRequests converts Terraform schema state change requests to HAPI StateChangeRequests
func SchemaToHapiStateChangeRequests(requests []interface{}) ([]*models.V1VMVirtualMachineStateChangeRequest, error) {
	if len(requests) == 0 {
		return nil, nil
	}

	result := make([]*models.V1VMVirtualMachineStateChangeRequest, len(requests))

	for i, request := range requests {
		if request == nil {
			continue
		}

		reqMap := request.(map[string]interface{})
		req := &models.V1VMVirtualMachineStateChangeRequest{}

		if v, ok := reqMap["action"].(string); ok {
			req.Action = &v
		}

		if v, ok := reqMap["data"].(map[string]interface{}); ok && len(v) > 0 {
			dataMap := make(map[string]string)
			for k, val := range v {
				dataMap[k] = val.(string)
			}
			req.Data = dataMap
		}

		if v, ok := reqMap["uid"].(string); ok {
			req.UID = v
		}

		result[i] = req
	}

	return result, nil
}

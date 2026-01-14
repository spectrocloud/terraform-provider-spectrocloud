package convert

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

// HapiVmStatusToSchema converts HAPI VM Status to Terraform schema
func HapiVmStatusToSchema(status *models.V1ClusterVirtualMachineStatus, d *schema.ResourceData) error {
	if status == nil {
		return nil
	}

	statusMap := make(map[string]interface{})

	// Created
	statusMap["created"] = status.Created

	// Ready
	statusMap["ready"] = status.Ready

	// Conditions
	if status.Conditions != nil && len(status.Conditions) > 0 {
		conditions, err := HapiConditionsToSchema(status.Conditions)
		if err != nil {
			return fmt.Errorf("failed to convert conditions: %w", err)
		}
		statusMap["conditions"] = conditions
	}

	// StateChangeRequests
	if status.StateChangeRequests != nil && len(status.StateChangeRequests) > 0 {
		stateChangeRequests, err := HapiStateChangeRequestsToSchema(status.StateChangeRequests)
		if err != nil {
			return fmt.Errorf("failed to convert state change requests: %w", err)
		}
		statusMap["state_change_requests"] = stateChangeRequests
	}

	return d.Set("status", []interface{}{statusMap})
}

// HapiConditionsToSchema converts HAPI Conditions to Terraform schema
func HapiConditionsToSchema(conditions []*models.V1VMVirtualMachineCondition) ([]interface{}, error) {
	if len(conditions) == 0 {
		return nil, nil
	}

	result := make([]interface{}, len(conditions))

	for i, condition := range conditions {
		if condition == nil {
			continue
		}

		condMap := make(map[string]interface{})

		if condition.Type != nil {
			condMap["type"] = *condition.Type
		}

		if condition.Status != nil {
			condMap["status"] = *condition.Status
		}

		condMap["reason"] = condition.Reason
		condMap["message"] = condition.Message

		result[i] = condMap
	}

	return result, nil
}

// HapiStateChangeRequestsToSchema converts HAPI StateChangeRequests to Terraform schema
func HapiStateChangeRequestsToSchema(requests []*models.V1VMVirtualMachineStateChangeRequest) ([]interface{}, error) {
	if len(requests) == 0 {
		return nil, nil
	}

	result := make([]interface{}, len(requests))

	for i, request := range requests {
		if request == nil {
			continue
		}

		reqMap := make(map[string]interface{})

		if request.Action != nil {
			reqMap["action"] = *request.Action
		}

		if request.Data != nil && len(request.Data) > 0 {
			dataMap := make(map[string]interface{})
			for k, v := range request.Data {
				dataMap[k] = v
			}
			reqMap["data"] = dataMap
		}

		if request.UID != "" {
			reqMap["uid"] = request.UID
		}

		result[i] = reqMap
	}

	return result, nil
}

package convert

import (
	"encoding/json"
	"fmt"

	"github.com/spectrocloud/palette-sdk-go/api/models"
)

func ToKubevirtVMStatusM(status *models.V1ClusterVirtualMachineStatus) (models.V1ClusterVirtualMachineStatus, error) {
	var kubevirtVMStatus models.V1ClusterVirtualMachineStatus

	// Marshal the input spec to JSON
	hapiClusterVMSpecJSON, err := json.Marshal(status)
	if err != nil {
		return kubevirtVMStatus, fmt.Errorf("failed to marshal models.V1ClusterVirtualMachineSpec to JSON: %v", err)
	}

	// Unmarshal the JSON to the desired Kubevirt VM spec
	err = json.Unmarshal(hapiClusterVMSpecJSON, &kubevirtVMStatus)
	if err != nil {
		return kubevirtVMStatus, fmt.Errorf("failed to unmarshal JSON to VirtualMachineSpec: %v", err)
	}

	return kubevirtVMStatus, nil
}

func ToKubevirtVMStatus(status *models.V1ClusterVirtualMachineStatus) models.V1ClusterVirtualMachineStatus {
	if status == nil {
		return models.V1ClusterVirtualMachineStatus{}
	}
	return models.V1ClusterVirtualMachineStatus{
		SnapshotInProgress:     status.SnapshotInProgress,
		RestoreInProgress:      status.RestoreInProgress,
		Created:                status.Created,
		Ready:                  status.Ready,
		PrintableStatus:        status.PrintableStatus,
		Conditions:             status.Conditions,
		StateChangeRequests:    status.StateChangeRequests,
		VolumeRequests:         status.VolumeRequests,
		VolumeSnapshotStatuses: status.VolumeSnapshotStatuses,
		StartFailure:           status.StartFailure,
		MemoryDumpRequest:      status.MemoryDumpRequest,
	}
}

func ToKvVmStatusConditions(conditions []*models.V1VMVirtualMachineCondition) []models.V1VMVirtualMachineCondition {
	var kvConditions []models.V1VMVirtualMachineCondition
	for _, condition := range conditions {
		kvConditions = append(kvConditions, ToKvVmStatusCondition(condition))
	}
	return kvConditions
}

func ToKvVmStatusCondition(condition *models.V1VMVirtualMachineCondition) models.V1VMVirtualMachineCondition {
	if condition == nil {
		return models.V1VMVirtualMachineCondition{}
	}
	return models.V1VMVirtualMachineCondition{
		Type:               condition.Type,
		Status:             condition.Status,
		LastProbeTime:      condition.LastProbeTime,
		LastTransitionTime: condition.LastTransitionTime,
		Reason:             condition.Reason,
		Message:            condition.Message,
	}
}

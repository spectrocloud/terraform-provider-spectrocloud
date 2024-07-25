package convert

import (
	"encoding/json"
	"fmt"
	"github.com/spectrocloud/palette-api-go/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
	k8sv1 "k8s.io/api/core/v1"
	kubevirtapiv1 "kubevirt.io/api/core/v1"
)

func ToKubevirtVMStatusM(status *models.V1ClusterVirtualMachineStatus) (kubevirtapiv1.VirtualMachineStatus, error) {
	var kubevirtVMStatus kubevirtapiv1.VirtualMachineStatus

	// Marshal the input spec to JSON
	hapiClusterVMSpecJSON, err := json.Marshal(status)
	if err != nil {
		return kubevirtVMStatus, fmt.Errorf("failed to marshal models.V1ClusterVirtualMachineSpec to JSON: %v", err)
	}

	// Unmarshal the JSON to the desired Kubevirt VM spec
	err = json.Unmarshal(hapiClusterVMSpecJSON, &kubevirtVMStatus)
	if err != nil {
		return kubevirtVMStatus, fmt.Errorf("failed to unmarshal JSON to kubevirtapiv1.VirtualMachineSpec: %v", err)
	}

	return kubevirtVMStatus, nil
}

func ToKubevirtVMStatus(status *models.V1ClusterVirtualMachineStatus) kubevirtapiv1.VirtualMachineStatus {

	var PrintableStatus kubevirtapiv1.VirtualMachinePrintableStatus
	if status.PrintableStatus != "" {
		PrintableStatus = kubevirtapiv1.VirtualMachinePrintableStatus(status.PrintableStatus)
	}

	return kubevirtapiv1.VirtualMachineStatus{
		SnapshotInProgress:     types.Ptr(status.SnapshotInProgress),
		RestoreInProgress:      types.Ptr(status.RestoreInProgress),
		Created:                status.Created,
		Ready:                  status.Ready,
		PrintableStatus:        PrintableStatus,
		Conditions:             ToKvVmStatusConditions(status.Conditions),
		StateChangeRequests:    nil,
		VolumeRequests:         nil,
		VolumeSnapshotStatuses: nil,
		StartFailure:           nil,
		MemoryDumpRequest:      nil,
	}
}

func ToKvVmStatusConditions(conditions []*models.V1VMVirtualMachineCondition) []kubevirtapiv1.VirtualMachineCondition {
	var kvConditions []kubevirtapiv1.VirtualMachineCondition
	for _, condition := range conditions {
		kvConditions = append(kvConditions, ToKvVmStatusCondition(condition))
	}
	return kvConditions
}

func ToKvVmStatusCondition(condition *models.V1VMVirtualMachineCondition) kubevirtapiv1.VirtualMachineCondition {
	if condition == nil {
		return kubevirtapiv1.VirtualMachineCondition{}
	}

	var VirtualMachineConditionType kubevirtapiv1.VirtualMachineConditionType
	if condition.Type != nil {
		VirtualMachineConditionType = kubevirtapiv1.VirtualMachineConditionType(*condition.Type)
	}

	var ConditionStatus k8sv1.ConditionStatus
	if condition.Status != nil {
		ConditionStatus = k8sv1.ConditionStatus(*condition.Status)
	}

	return kubevirtapiv1.VirtualMachineCondition{
		Type:   VirtualMachineConditionType,
		Status: ConditionStatus,
		// TODO: LastProbeTime:      condition.LastProbeTime,
		// TODO: LastTransitionTime: condition.LastTransitionTime,
		Reason:  condition.Reason,
		Message: condition.Message,
	}
}

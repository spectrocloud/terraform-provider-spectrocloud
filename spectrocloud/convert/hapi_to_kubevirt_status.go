package convert

import (
	"github.com/spectrocloud/gomi/pkg/ptr"
	"github.com/spectrocloud/hapi/models"
	k8sv1 "k8s.io/api/core/v1"
	kubevirtapiv1 "kubevirt.io/api/core/v1"
)

func ToKubevirtVMStatus(status *models.V1ClusterVirtualMachineStatus) kubevirtapiv1.VirtualMachineStatus {

	var PrintableStatus kubevirtapiv1.VirtualMachinePrintableStatus
	if status.PrintableStatus != "" {
		PrintableStatus = kubevirtapiv1.VirtualMachinePrintableStatus(status.PrintableStatus)
	}

	return kubevirtapiv1.VirtualMachineStatus{
		SnapshotInProgress:     ptr.StringPtr(status.SnapshotInProgress),
		RestoreInProgress:      ptr.StringPtr(status.RestoreInProgress),
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

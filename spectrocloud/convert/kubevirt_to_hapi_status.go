package convert

import (
	"github.com/spectrocloud/gomi/pkg/ptr"
	"github.com/spectrocloud/hapi/models"
	kubevirtapiv1 "kubevirt.io/api/core/v1"
)

func ToHapiVmStatus(status kubevirtapiv1.VirtualMachineStatus) *models.V1ClusterVirtualMachineStatus {
	var RestoreInProgress string
	if status.RestoreInProgress != nil {
		RestoreInProgress = *status.RestoreInProgress
	}

	var SnapshotInProgress string
	if status.SnapshotInProgress != nil {
		SnapshotInProgress = *status.SnapshotInProgress
	}
	return &models.V1ClusterVirtualMachineStatus{
		Conditions:             ToHapiVmStatusConditions(status.Conditions),
		Created:                status.Created,
		MemoryDumpRequest:      ToHapiVmStatusMemoryDumpRequest(status.MemoryDumpRequest),
		PrintableStatus:        string(status.PrintableStatus),
		Ready:                  status.Ready,
		RestoreInProgress:      RestoreInProgress,
		SnapshotInProgress:     SnapshotInProgress,
		StartFailure:           ToHapiVmStatusStartFailure(status.StartFailure),
		StateChangeRequests:    ToHapiVmStatusStateChangeRequests(status.StateChangeRequests),
		VolumeRequests:         ToHapiVmStatusVolumeRequests(status.VolumeRequests),
		VolumeSnapshotStatuses: ToHapiVmStatusVolumeSnapshotStatuses(status.VolumeSnapshotStatuses),
	}
}

func ToHapiVmStatusVolumeSnapshotStatuses(statuses []kubevirtapiv1.VolumeSnapshotStatus) []*models.V1VMVolumeSnapshotStatus {
	var hapiStatuses []*models.V1VMVolumeSnapshotStatus
	for _, status := range statuses {
		hapiStatuses = append(hapiStatuses, ToHapiVmStatusVolumeSnapshotStatus(&status))
	}
	return hapiStatuses
}

func ToHapiVmStatusVolumeSnapshotStatus(k *kubevirtapiv1.VolumeSnapshotStatus) *models.V1VMVolumeSnapshotStatus {
	if k == nil {
		return nil
	}
	return &models.V1VMVolumeSnapshotStatus{
		Enabled: ptr.BoolPtr(k.Enabled),
		Name:    ptr.StringPtr(k.Name),
		Reason:  k.Reason,
	}
}

func ToHapiVmStatusVolumeRequests(requests []kubevirtapiv1.VirtualMachineVolumeRequest) []*models.V1VMVirtualMachineVolumeRequest {
	var hapiRequests []*models.V1VMVirtualMachineVolumeRequest
	for _, request := range requests {
		hapiRequests = append(hapiRequests, ToHapiVmStatusVolumeRequest(&request))
	}
	return hapiRequests
}

func ToHapiVmStatusVolumeRequest(k *kubevirtapiv1.VirtualMachineVolumeRequest) *models.V1VMVirtualMachineVolumeRequest {
	if k == nil {
		return nil
	}
	return &models.V1VMVirtualMachineVolumeRequest{
		AddVolumeOptions:    ToHapiVmStatusAddVolumeOptions(k.AddVolumeOptions),
		RemoveVolumeOptions: ToHapiVmStatusRemoveVolumeOptions(k.RemoveVolumeOptions),
	}
}

func ToHapiVmStatusRemoveVolumeOptions(options *kubevirtapiv1.RemoveVolumeOptions) *models.V1VMRemoveVolumeOptions {
	if options == nil {
		return nil
	}
	return &models.V1VMRemoveVolumeOptions{
		Name:   ptr.StringPtr(options.Name),
		DryRun: options.DryRun,
	}
}

func ToHapiVmStatusAddVolumeOptions(options *kubevirtapiv1.AddVolumeOptions) *models.V1VMAddVolumeOptions {
	if options == nil {
		return nil
	}
	return &models.V1VMAddVolumeOptions{
		Disk:         ToHapiVmStatusDisk(options.Disk),
		DryRun:       options.DryRun,
		Name:         ptr.StringPtr(options.Name),
		VolumeSource: ToHapiVmStatusVolumeSource(options.VolumeSource),
	}
}

func ToHapiVmStatusVolumeSource(source *kubevirtapiv1.HotplugVolumeSource) *models.V1VMHotplugVolumeSource {
	if source == nil {
		return nil
	}
	return &models.V1VMHotplugVolumeSource{
		DataVolume:            ToHapiVmStatusDataVolume(source.DataVolume),
		PersistentVolumeClaim: ToHapiVmStatusPersistentVolumeClaim(source.PersistentVolumeClaim),
	}
}

func ToHapiVmStatusPersistentVolumeClaim(claim *kubevirtapiv1.PersistentVolumeClaimVolumeSource) *models.V1VMPersistentVolumeClaimVolumeSource {
	if claim == nil {
		return nil
	}
	return &models.V1VMPersistentVolumeClaimVolumeSource{
		ClaimName:    ptr.StringPtr(claim.ClaimName),
		Hotpluggable: claim.Hotpluggable,
		ReadOnly:     claim.ReadOnly,
	}
}

func ToHapiVmStatusDataVolume(volume *kubevirtapiv1.DataVolumeSource) *models.V1VMCoreDataVolumeSource {
	if volume == nil {
		return nil
	}
	return &models.V1VMCoreDataVolumeSource{
		Name:         ptr.StringPtr(volume.Name),
		Hotpluggable: volume.Hotpluggable,
	}
}

func ToHapiVmStatusDisk(disk *kubevirtapiv1.Disk) *models.V1VMDisk {
	if disk == nil {
		return nil
	}

	/*bootOrder := int32(1) // Default value.
	if disk.BootOrder != nil {
		bootOrder = int32(*disk.BootOrder)
	}*/

	DedicatedIOThread := false
	if disk.DedicatedIOThread != nil {
		DedicatedIOThread = *disk.DedicatedIOThread
	}

	return &models.V1VMDisk{
		BlockSize: ToHapiVmBlockSize(disk.BlockSize),
		// TODO : BootOrder:         bootOrder,
		Cache:             string(disk.Cache),
		Cdrom:             ToHapiVmCdRom(disk.CDRom),
		DedicatedIOThread: DedicatedIOThread,
		Disk:              ToHapiVmDiskTarget(disk.Disk),
		Io:                string(disk.IO),
		Lun:               ToHapiVmLunTarget(disk.LUN),
		Name:              ptr.StringPtr(disk.Name),
		Serial:            disk.Serial,
		Shareable:         false,
		Tag:               disk.Tag,
	}
}

func ToHapiVmStatusStateChangeRequests(requests []kubevirtapiv1.VirtualMachineStateChangeRequest) []*models.V1VMVirtualMachineStateChangeRequest {
	var hapiRequests []*models.V1VMVirtualMachineStateChangeRequest
	for _, request := range requests {
		hapiRequests = append(hapiRequests, ToHapiVmStatusStateChangeRequest(&request))
	}
	return hapiRequests
}

func ToHapiVmStatusStateChangeRequest(k *kubevirtapiv1.VirtualMachineStateChangeRequest) *models.V1VMVirtualMachineStateChangeRequest {
	if k == nil {
		return nil
	}

	var uid string
	if k.UID != nil {
		uid = string(*k.UID)
	}
	return &models.V1VMVirtualMachineStateChangeRequest{
		Action: ptr.StringPtr(string(k.Action)),
		Data:   k.Data,
		UID:    uid,
	}
}

func ToHapiVmStatusStartFailure(failure *kubevirtapiv1.VirtualMachineStartFailure) *models.V1VMVirtualMachineStartFailure {
	if failure == nil {
		return nil
	}
	return &models.V1VMVirtualMachineStartFailure{
		ConsecutiveFailCount: int32(failure.ConsecutiveFailCount),
		LastFailedVMIUID:     string(failure.LastFailedVMIUID),
		// TODO: RetryAfterTimestamp:  models.V1Time{},
	}
}

func ToHapiVmStatusMemoryDumpRequest(request *kubevirtapiv1.VirtualMachineMemoryDumpRequest) *models.V1VMVirtualMachineMemoryDumpRequest {
	if request == nil {
		return nil
	}

	var FileName string
	if request.FileName != nil {
		FileName = *request.FileName
	}
	return &models.V1VMVirtualMachineMemoryDumpRequest{
		ClaimName: ptr.StringPtr(request.ClaimName),
		// TODO: EndTimestamp:   models.V1Time{},
		FileName: FileName,
		Message:  request.Message,
		Phase:    ptr.StringPtr(string(request.Phase)),
		Remove:   request.Remove,
		// TODO: StartTimestamp: models.V1Time{},
	}
}

func ToHapiVmStatusConditions(conditions []kubevirtapiv1.VirtualMachineCondition) []*models.V1VMVirtualMachineCondition {
	var hapiConditions []*models.V1VMVirtualMachineCondition
	for _, condition := range conditions {
		hapiConditions = append(hapiConditions, &models.V1VMVirtualMachineCondition{
			// TODO: LastProbeTime:      condition.LastProbeTime,
			// TODO: LastTransitionTime: condition.LastTransitionTime,
			Message: condition.Message,
			Reason:  condition.Reason,
			Status:  ptr.StringPtr(string(condition.Status)),
			Type:    ptr.StringPtr(string(condition.Type)),
		})
	}
	return hapiConditions
}

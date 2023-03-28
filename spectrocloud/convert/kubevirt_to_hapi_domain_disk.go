package convert

import (
	"github.com/spectrocloud/hapi/models"
	kubevirtapiv1 "kubevirt.io/api/core/v1"

	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

func ToHapiVmDisks(disks []kubevirtapiv1.Disk) []*models.V1VMDisk {
	var hapiDisks []*models.V1VMDisk
	for _, disk := range disks {
		hapiDisks = append(hapiDisks, ToHapiVmDisk(disk))
	}
	return hapiDisks
}

func ToHapiVmDisk(disk kubevirtapiv1.Disk) *models.V1VMDisk {
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
		// TODO: BootOrder:         bootOrder,
		Cache:             string(disk.Cache),
		Cdrom:             ToHapiVmCdRom(disk.CDRom),
		DedicatedIOThread: DedicatedIOThread,
		Disk:              ToHapiVmDiskTarget(disk.Disk),
		Io:                string(disk.IO),
		Lun:               ToHapiVmLunTarget(disk.LUN),
		Name:              types.Ptr(disk.Name),
		Serial:            disk.Serial,
		Shareable:         false,
		Tag:               disk.Tag,
	}
}

func ToHapiVmLunTarget(lun *kubevirtapiv1.LunTarget) *models.V1VMLunTarget {
	if lun == nil {
		return nil
	}

	return &models.V1VMLunTarget{
		Bus:      string(lun.Bus),
		Readonly: lun.ReadOnly,
	}
}

func ToHapiVmDiskTarget(disk *kubevirtapiv1.DiskTarget) *models.V1VMDiskTarget {
	if disk == nil {
		return nil
	}

	return &models.V1VMDiskTarget{
		Bus:        string(disk.Bus),
		PciAddress: disk.PciAddress,
		Readonly:   disk.ReadOnly,
	}
}

func ToHapiVmCdRom(cdrom *kubevirtapiv1.CDRomTarget) *models.V1VMCDRomTarget {
	if cdrom == nil {
		return nil
	}

	return &models.V1VMCDRomTarget{
		Bus: string(cdrom.Bus),
	}
}

func ToHapiVmBlockSize(size *kubevirtapiv1.BlockSize) *models.V1VMBlockSize {
	if size == nil {
		return nil
	}

	return &models.V1VMBlockSize{
		Custom:      ToHapiVmCustomBlockSize(size.Custom),
		MatchVolume: nil,
	}
}

func ToHapiVmCustomBlockSize(custom *kubevirtapiv1.CustomBlockSize) *models.V1VMCustomBlockSize {
	if custom == nil {
		return nil
	}

	return &models.V1VMCustomBlockSize{
		Physical: types.Ptr(int32(custom.Physical)),
		Logical:  types.Ptr(int32(custom.Logical)),
	}
}

package convert

import (
	"github.com/spectrocloud/hapi/models"
	kubevirtapiv1 "kubevirt.io/api/core/v1"

	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

func ToHapiVmCPU(cpu *kubevirtapiv1.CPU) *models.V1VMCPU {
	if cpu == nil {
		return nil
	}

	cores := int64(cpu.Cores)
	sockets := int64(cpu.Sockets)
	threads := int64(cpu.Threads)

	features := make([]*models.V1VMCPUFeature, len(cpu.Features))
	for i, f := range cpu.Features {
		features[i] = &models.V1VMCPUFeature{
			Name:   types.Ptr(f.Name),
			Policy: f.Policy,
		}
	}

	return &models.V1VMCPU{
		Cores:                 cores,
		DedicatedCPUPlacement: cpu.DedicatedCPUPlacement,
		Features:              features,
		IsolateEmulatorThread: cpu.IsolateEmulatorThread,
		Model:                 cpu.Model,
		Numa:                  ToHapiVmNUMA(cpu.NUMA),
		Realtime:              ToHapivmRealtime(cpu.Realtime),
		Sockets:               sockets,
		Threads:               threads,
	}
}

func ToHapiVmNUMA(numa *kubevirtapiv1.NUMA) *models.V1VMNUMA {
	if numa == nil {
		return nil
	}

	return &models.V1VMNUMA{
		GuestMappingPassthrough: numa.GuestMappingPassthrough,
	}
}

func ToHapivmRealtime(realtime *kubevirtapiv1.Realtime) *models.V1VMRealtime {
	if realtime == nil {
		return nil
	}

	return &models.V1VMRealtime{
		Mask: realtime.Mask,
	}
}

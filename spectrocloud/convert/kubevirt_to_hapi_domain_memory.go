package convert

import (
	"github.com/spectrocloud/hapi/models"
	"k8s.io/apimachinery/pkg/api/resource"
	kubevirtapiv1 "kubevirt.io/api/core/v1"
)

func ToHapiVmMemory(memory *kubevirtapiv1.Memory) *models.V1VMMemory {
	if memory == nil {
		return nil
	}

	return &models.V1VMMemory{
		Hugepages: ToHapiVmHugepages(memory.Hugepages),
		Guest:     ToHapiVmGuestMemory(memory.Guest),
	}
}

func ToHapiVmGuestMemory(guest *resource.Quantity) models.V1VMQuantity {
	if guest == nil {
		return ""
	}

	return models.V1VMQuantity(guest.String())
}

func ToHapiVmHugepages(hugepages *kubevirtapiv1.Hugepages) *models.V1VMHugepages {
	if hugepages == nil {
		return nil
	}

	return &models.V1VMHugepages{
		PageSize: hugepages.PageSize,
	}
}

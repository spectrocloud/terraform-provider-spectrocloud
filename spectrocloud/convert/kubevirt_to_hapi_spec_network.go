package convert

import (
	"github.com/spectrocloud/gomi/pkg/ptr"
	"github.com/spectrocloud/hapi/models"
	kubevirtapiv1 "kubevirt.io/api/core/v1"
)

func ToHapiVmNetworks(networks []kubevirtapiv1.Network) []*models.V1VMNetwork {
	var hapiNetworks []*models.V1VMNetwork
	for _, network := range networks {
		hapiNetworks = append(hapiNetworks, &models.V1VMNetwork{
			Multus: ToHapiVmMultus(network.Multus),
			Name:   ptr.StringPtr(network.Name),
			Pod:    ToHapiVmPodNetwork(network.Pod),
		})
	}
	return hapiNetworks
}

func ToHapiVmPodNetwork(pod *kubevirtapiv1.PodNetwork) *models.V1VMPodNetwork {
	if pod == nil {
		return nil
	}

	return &models.V1VMPodNetwork{
		VMIPV6NetworkCIDR: pod.VMIPv6NetworkCIDR,
		VMNetworkCIDR:     pod.VMNetworkCIDR,
	}
}

func ToHapiVmMultus(multus *kubevirtapiv1.MultusNetwork) *models.V1VMMultusNetwork {
	if multus == nil {
		return nil
	}

	return &models.V1VMMultusNetwork{
		Default:     multus.Default,
		NetworkName: ptr.StringPtr(multus.NetworkName),
	}
}

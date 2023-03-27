package convert

import (
	"github.com/spectrocloud/hapi/models"
	v1 "k8s.io/api/core/v1"
)

func ToHapiVmDNSConfig(config *v1.PodDNSConfig) *models.V1VMPodDNSConfig {
	var Nameservers []string
	if config != nil {
		Nameservers = config.Nameservers
	}

	return &models.V1VMPodDNSConfig{
		Nameservers: Nameservers,
		Options:     ToHapiVmPodDNSConfigOption(config.Options),
		Searches:    config.Searches,
	}
}

func ToHapiVmPodDNSConfigOption(options []v1.PodDNSConfigOption) []*models.V1VMPodDNSConfigOption {
	var result []*models.V1VMPodDNSConfigOption
	for _, option := range options {
		result = append(result, ToHapiVmPodDNSConfigOptionItem(option))
	}

	return result
}

func ToHapiVmPodDNSConfigOptionItem(option v1.PodDNSConfigOption) *models.V1VMPodDNSConfigOption {
	return &models.V1VMPodDNSConfigOption{
		Name:  option.Name,
		Value: *option.Value,
	}
}

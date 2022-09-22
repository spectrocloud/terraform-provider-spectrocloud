package spectrocloud

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"
)

func toClusterHostConfigs(d *schema.ResourceData) *models.V1HostClusterConfig {
	if d.Get("host_config") != nil {
		for _, hostConfig := range d.Get("host_config").([]interface{}) {
			return toClusterHostConfig(hostConfig.(map[string]interface{}))
		}
	}

	isHostCluster := false
	return &models.V1HostClusterConfig{
		ClusterEndpoint: nil,
		IsHostCluster:   &isHostCluster,
	}
}

func toClusterHostConfig(config map[string]interface{}) *models.V1HostClusterConfig {
	isHostCluster := true
	return &models.V1HostClusterConfig{
		ClusterEndpoint: toClusterEndpoint(config),
		IsHostCluster:   &isHostCluster,
	}
}

func toClusterEndpoint(config map[string]interface{}) *models.V1HostClusterEndpoint {
	hostType := "Ingress"
	if config["host_endpoint_type"] != nil {
		hostType = config["host_endpoint_type"].(string)
	}
	return &models.V1HostClusterEndpoint{
		Config: toClusterEndpointConfig(config),
		Type:   hostType,
	}
}

func toClusterEndpointConfig(config map[string]interface{}) *models.V1HostClusterEndpointConfig {
	return &models.V1HostClusterEndpointConfig{
		IngressConfig:      toIngressConfig(config),
		LoadBalancerConfig: toLoadBalancerConfig(config),
	}
}

func toIngressConfig(config map[string]interface{}) *models.V1IngressConfig {
	ingressHost := ""
	if config["ingress_host"] != nil {
		ingressHost = config["ingress_host"].(string)
	}
	return &models.V1IngressConfig{
		Host: ingressHost,
	}
}

func toLoadBalancerConfig(config map[string]interface{}) *models.V1LoadBalancerConfig {

	loadBalancerConfig := &models.V1LoadBalancerConfig{}

	if config["external_traffic_policy"] != nil {
		loadBalancerConfig.ExternalTrafficPolicy = config["external_traffic_policy"].(string)
	}

	if config["load_balancer_source_ranges"] != nil {
		loadBalancerConfig.LoadBalancerSourceRanges = strings.Split(config["load_balancer_source_ranges"].(string), ",")
	}

	return loadBalancerConfig
}

func flattenHostConfig(hostConfig *models.V1HostClusterConfig) []interface{} {
	result := make(map[string]interface{})
	configs := make([]interface{}, 0)

	if hostConfig != nil && hostConfig.ClusterEndpoint != nil {
		if hostConfig.ClusterEndpoint != nil {
			result["host_endpoint_type"] = hostConfig.ClusterEndpoint.Type
		}
		if hostConfig.ClusterEndpoint.Config != nil {
			if hostConfig.ClusterEndpoint.Config.IngressConfig != nil {
				result["ingress_host"] = hostConfig.ClusterEndpoint.Config.IngressConfig.Host

			}
			if hostConfig.ClusterEndpoint.Config.LoadBalancerConfig != nil {
				result["external_traffic_policy"] = hostConfig.ClusterEndpoint.Config.LoadBalancerConfig.ExternalTrafficPolicy
				result["load_balancer_source_ranges"] = flattenSourceRanges(hostConfig)
			}
		}
		configs = append(configs, result)
	}

	return configs
}

func flattenSourceRanges(hostConfig *models.V1HostClusterConfig) string {
	sourceRanges := make([]string, 0)
	for _, r := range hostConfig.ClusterEndpoint.Config.LoadBalancerConfig.LoadBalancerSourceRanges {
		sourceRanges = append(sourceRanges, r)
	}
	return strings.Join(sourceRanges, ",")
}

func flattenExternalIPs(hostConfig *models.V1HostClusterConfig) string {
	externalIPs := make([]string, 0)
	for _, ip := range hostConfig.ClusterEndpoint.Config.LoadBalancerConfig.ExternalIPs {
		externalIPs = append(externalIPs, ip)
	}
	return strings.Join(externalIPs, ",")
}

func updateHostConfig(c *client.V1Client, d *schema.ResourceData) error {
	if hostConfigs := toClusterHostConfigs(d); hostConfigs != nil {
		return c.ApplyClusterHostConfig(d.Id(), &models.V1HostClusterConfigEntity{
			HostClusterConfig: hostConfigs,
		})
	}
	return nil
}

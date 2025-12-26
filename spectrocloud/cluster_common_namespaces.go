package spectrocloud

import (
	"math"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
)

func toClusterNamespaces(d *schema.ResourceData) []*models.V1ClusterNamespaceResourceInputEntity {
	clusterNamespaces := make([]*models.V1ClusterNamespaceResourceInputEntity, 0)
	if d.Get("namespaces") == nil {
		return nil
	}
	for _, clusterNamespace := range d.Get("namespaces").([]interface{}) {
		ns := toClusterNamespace(clusterNamespace)
		clusterNamespaces = append(clusterNamespaces, ns)
	}

	return clusterNamespaces
}

func toClusterNamespace(clusterRbacBinding interface{}) *models.V1ClusterNamespaceResourceInputEntity {
	m := clusterRbacBinding.(map[string]interface{})

	resourceAllocation, _ := m["resource_allocation"].(map[string]interface{})

	cpu_cores, err := strconv.ParseFloat(resourceAllocation["cpu_cores"].(string), 64)
	if err != nil {
		return nil
	}

	memory_MiB, err := strconv.ParseFloat(resourceAllocation["memory_MiB"].(string), 64)
	if err != nil {
		return nil
	}

	var gpuConfig *models.V1GpuConfig
	if gpuLimitVal, exists := resourceAllocation["gpu_limit"]; exists && gpuLimitVal != nil {
		gpu_limit, err := strconv.ParseInt(gpuLimitVal.(string), 10, 32)
		if err != nil {
			return nil
		}

		gpu_provider := "nvidia"
		if provider, exists := resourceAllocation["gpu_provider"]; exists && provider != nil {
			gpu_provider = provider.(string)
		}

		gpuConfig = &models.V1GpuConfig{
			Limit:    int32(gpu_limit),
			Provider: &gpu_provider,
		}
	}

	resource_alloc := &models.V1ClusterNamespaceResourceAllocation{
		CPUCores:  cpu_cores,
		MemoryMiB: memory_MiB,
		GpuConfig: gpuConfig,
	}

	ns := &models.V1ClusterNamespaceResourceInputEntity{
		Metadata: &models.V1ObjectMetaUpdateEntity{
			Name: m["name"].(string),
		},
		Spec: &models.V1ClusterNamespaceSpec{
			IsRegex:            IsRegex(m["name"].(string)),
			ResourceAllocation: resource_alloc,
		},
	}

	return ns
}

func flattenClusterNamespaces(items []*models.V1ClusterNamespaceResource) []interface{} {
	result := make([]interface{}, 0)
	for _, namespace := range items {
		flattenNamespace := make(map[string]interface{})
		flattenNamespace["name"] = namespace.Metadata.Name

		flattenResourceAllocation := make(map[string]interface{})
		flattenResourceAllocation["cpu_cores"] = strconv.Itoa(int(math.Round(namespace.Spec.ResourceAllocation.CPUCores)))
		flattenResourceAllocation["memory_MiB"] = strconv.Itoa(int(math.Round(namespace.Spec.ResourceAllocation.MemoryMiB)))

		// Only set GPU fields if GpuConfig exists and has meaningful values
		if namespace.Spec.ResourceAllocation.GpuConfig != nil && namespace.Spec.ResourceAllocation.GpuConfig.Limit > 0 {
			flattenResourceAllocation["gpu_limit"] = strconv.Itoa(int(namespace.Spec.ResourceAllocation.GpuConfig.Limit))
			if namespace.Spec.ResourceAllocation.GpuConfig.Provider != nil {
				flattenResourceAllocation["gpu_provider"] = *namespace.Spec.ResourceAllocation.GpuConfig.Provider
			} else {
				flattenResourceAllocation["gpu_provider"] = "nvidia"
			}
		}

		flattenNamespace["resource_allocation"] = flattenResourceAllocation
		result = append(result, flattenNamespace)
	}
	return result
}

func updateClusterNamespaces(c *client.V1Client, d *schema.ResourceData) error {
	if namespaces := toClusterNamespaces(d); namespaces != nil {
		clusterContext := d.Get("context").(string)
		err := ValidateContext(clusterContext)
		if err != nil {
			return err
		}
		return c.ApplyClusterNamespaceConfig(d.Id(), namespaces)
	}
	return nil
}

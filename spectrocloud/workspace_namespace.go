package spectrocloud

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/constants"
)

// Helper function to create V1WorkspaceResourceAllocation from resource allocation map
func toWorkspaceResourceAllocation(resourceAllocation map[string]interface{}) (*models.V1WorkspaceResourceAllocation, error) {
	cpu_cores, err := strconv.ParseFloat(resourceAllocation["cpu_cores"].(string), 64)
	if err != nil {
		return nil, err
	}

	memory_MiB, err := strconv.ParseFloat(resourceAllocation["memory_MiB"].(string), 64)
	if err != nil {
		return nil, err
	}

	resource_alloc := &models.V1WorkspaceResourceAllocation{
		CPUCores:  cpu_cores,
		MemoryMiB: memory_MiB,
	}

	// Handle GPU configuration if specified
	if gpuVal, exists := resourceAllocation["gpu"]; exists && gpuVal.(int) > 0 {
		gpuInt := gpuVal.(int)
		if gpuInt > constants.Int32MaxValue {
			return nil, fmt.Errorf("gpu value %d is out of range for int32", gpuInt)
		}
		provider := "nvidia" // Default provider for cluster allocations
		// gpu_provider is optional - mainly used for default resource allocations
		if gpuProvider, providerExists := resourceAllocation["gpu_provider"]; providerExists && gpuProvider.(string) != "" {
			provider = gpuProvider.(string)
		}
		resource_alloc.GpuConfig = &models.V1GpuConfig{
			Limit:    SafeInt32(gpuInt),
			Provider: &provider,
		}
	}

	return resource_alloc, nil
}

func toWorkspaceNamespaces(d *schema.ResourceData) []*models.V1WorkspaceClusterNamespace {
	workspaceNamespaces := make([]*models.V1WorkspaceClusterNamespace, 0)
	if d.Get("namespaces") == nil {
		return nil
	}
	for _, clusterNamespace := range d.Get("namespaces").([]interface{}) {
		ns := toWorkspaceNamespace(clusterNamespace)
		workspaceNamespaces = append(workspaceNamespaces, ns)
	}

	return workspaceNamespaces
}

func toWorkspaceNamespace(clusterNamespaceConfig interface{}) *models.V1WorkspaceClusterNamespace {
	m := clusterNamespaceConfig.(map[string]interface{})

	// Handle default resource allocation
	resourceAllocation, _ := m["resource_allocation"].(map[string]interface{})
	defaultResourceAlloc, err := toWorkspaceResourceAllocation(resourceAllocation)
	if err != nil {
		return nil
	}

	// Handle cluster resource allocations
	var clusterResourceAllocations []*models.V1ClusterResourceAllocation
	if clusterAllocationsData, exists := m["cluster_resource_allocations"]; exists {
		clusterAllocations := clusterAllocationsData.([]interface{})
		for _, clusterAlloc := range clusterAllocations {
			clusterAllocMap := clusterAlloc.(map[string]interface{})
			uid := clusterAllocMap["uid"].(string)
			clusterResourceAllocation := clusterAllocMap["resource_allocation"].(map[string]interface{})

			resourceAlloc, err := toWorkspaceResourceAllocation(clusterResourceAllocation)
			if err != nil {
				continue // Skip invalid allocations
			}

			clusterResourceAllocations = append(clusterResourceAllocations, &models.V1ClusterResourceAllocation{
				ClusterUID:         uid,
				ResourceAllocation: resourceAlloc,
			})
		}
	}

	// Handle images blacklist
	images, _ := m["images_blacklist"].([]interface{})
	blacklist := make([]string, 0)
	for _, image := range images {
		blacklist = append(blacklist, image.(string))
	}

	name := m["name"].(string)
	IsRegex := IsRegex(name)

	ns := &models.V1WorkspaceClusterNamespace{
		Image: &models.V1WorkspaceNamespaceImage{
			BlackListedImages: blacklist,
		},
		Name:    name,
		IsRegex: IsRegex,
		NamespaceResourceAllocation: &models.V1WorkspaceNamespaceResourceAllocation{
			ClusterResourceAllocations: clusterResourceAllocations,
			DefaultResourceAllocation:  defaultResourceAlloc,
		},
	}

	return ns
}

func IsRegex(name string) bool {
	last := string(name[len(name)-1])

	if !((strings.HasPrefix(name, "~/") || strings.HasPrefix(name, "/")) && last == "/") {
		return false // not a regular expression since it doesn't start with ~/ / or end with /
	}

	exp := name
	if strings.HasPrefix(name, "~/") && len(name) > 3 {
		exp = name[2 : len(name)-2]
	}
	if strings.HasPrefix(name, "/") && len(name) > 2 {
		exp = name[1 : len(name)-1]
	}

	_, err := regexp.Compile(exp)
	if err == nil {
		return true
	} else {
		return false // not a valid regex doesn't compile
	}
}

func toUpdateWorkspaceNamespaces(d *schema.ResourceData, c *client.V1Client) (*models.V1WorkspaceClusterNamespacesEntity, error) {
	quota, err := toQuota(d)
	if err != nil {
		return nil, err
	}

	return &models.V1WorkspaceClusterNamespacesEntity{
		ClusterNamespaces: toWorkspaceNamespaces(d),
		ClusterRefs:       toClusterRefs(d, c),
		Quota:             quota,
	}, nil
}

// Helper function to flatten V1WorkspaceResourceAllocation to resource allocation map
// includeProvider controls whether to include gpu_provider field (true for default allocations, false for cluster allocations)
func flattenWorkspaceResourceAllocation(resourceAlloc *models.V1WorkspaceResourceAllocation, includeProvider bool) map[string]interface{} {
	result := make(map[string]interface{})

	// Convert CPU cores with bounds checking to prevent integer overflow
	cpuCoresRounded := math.Round(resourceAlloc.CPUCores)
	if cpuCoresRounded > math.MaxInt || cpuCoresRounded < math.MinInt {
		// Fallback to string representation if out of int range
		result["cpu_cores"] = fmt.Sprintf("%.0f", cpuCoresRounded)
	} else {
		result["cpu_cores"] = strconv.Itoa(int(cpuCoresRounded))
	}

	// Convert memory with bounds checking to prevent integer overflow
	memoryMiBRounded := math.Round(resourceAlloc.MemoryMiB)
	if memoryMiBRounded > math.MaxInt || memoryMiBRounded < math.MinInt {
		// Fallback to string representation if out of int range
		result["memory_MiB"] = fmt.Sprintf("%.0f", memoryMiBRounded)
	} else {
		result["memory_MiB"] = strconv.Itoa(int(memoryMiBRounded))
	}

	// Handle GPU configuration if present
	if resourceAlloc.GpuConfig != nil {
		// Convert GPU limit with bounds checking to prevent integer overflow
		gpuLimit := int64(resourceAlloc.GpuConfig.Limit)
		if gpuLimit > math.MaxInt || gpuLimit < math.MinInt {
			// Fallback to string representation if out of int range
			result["gpu_limit"] = fmt.Sprintf("%d", gpuLimit)
		} else {
			result["gpu_limit"] = strconv.Itoa(int(gpuLimit))
		}
		// Only include gpu_provider for default resource allocations, not cluster-specific ones
		if includeProvider {
			if resourceAlloc.GpuConfig.Provider != nil {
				result["gpu_provider"] = *resourceAlloc.GpuConfig.Provider
			} else {
				result["gpu_provider"] = "nvidia" // Default provider
			}
		}
	} else {
		result["gpu_limit"] = "0"
		if includeProvider {
			result["gpu_provider"] = ""
		}
	}

	return result
}

func flattenWorkspaceClusterNamespaces(items []*models.V1WorkspaceClusterNamespace) []interface{} {
	result := make([]interface{}, 0)
	for _, namespace := range items {
		flattenNamespace := make(map[string]interface{})
		flattenNamespace["name"] = namespace.Name

		// Flatten default resource allocation using helper (include gpu_provider)
		if namespace.NamespaceResourceAllocation != nil && namespace.NamespaceResourceAllocation.DefaultResourceAllocation != nil {
			flattenNamespace["resource_allocation"] = flattenWorkspaceResourceAllocation(namespace.NamespaceResourceAllocation.DefaultResourceAllocation, true)
		}

		// Flatten cluster resource allocations (exclude gpu_provider)
		if namespace.NamespaceResourceAllocation != nil && len(namespace.NamespaceResourceAllocation.ClusterResourceAllocations) > 0 {
			clusterAllocations := make([]interface{}, 0)
			for _, clusterAlloc := range namespace.NamespaceResourceAllocation.ClusterResourceAllocations {
				clusterAllocMap := map[string]interface{}{
					"uid": clusterAlloc.ClusterUID,
				}
				if clusterAlloc.ResourceAllocation != nil {
					clusterAllocMap["resource_allocation"] = flattenWorkspaceResourceAllocation(clusterAlloc.ResourceAllocation, false)
				}
				clusterAllocations = append(clusterAllocations, clusterAllocMap)
			}
			flattenNamespace["cluster_resource_allocations"] = clusterAllocations
		}

		// Handle images blacklist
		if namespace.Image != nil {
			flattenNamespace["images_blacklist"] = namespace.Image.BlackListedImages
		}

		result = append(result, flattenNamespace)
	}
	return result
}

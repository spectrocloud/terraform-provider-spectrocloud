package spectrocloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/hapi/models"
	"math"
	"strconv"
)

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

func toWorkspaceNamespace(clusterRbacBinding interface{}) *models.V1WorkspaceClusterNamespace {
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

	resource_alloc := &models.V1WorkspaceResourceAllocation{
		CPUCores:  cpu_cores,
		MemoryMiB: memory_MiB,
	}

	images, _ := m["images_blacklist"].([]interface{})
	blacklist := make([]string, 0)
	for _, image := range images {
		blacklist = append(blacklist, image.(string))
	}

	name := m["name"].(string)
	IsRegex := false

	first := string(name[0])
	last := string(name[len(name)-1])

	if first == "/" && last == "/" {
		IsRegex = true
	}

	ns := &models.V1WorkspaceClusterNamespace{
		Image: &models.V1WorkspaceNamespaceImage{
			BlackListedImages: blacklist,
		},
		Name:    name,
		IsRegex: IsRegex,
		NamespaceResourceAllocation: &models.V1WorkspaceNamespaceResourceAllocation{
			ClusterResourceAllocations: nil,
			DefaultResourceAllocation:  resource_alloc,
		},
	}

	return ns
}

func toUpdateWorkspaceNamespaces(d *schema.ResourceData) *models.V1WorkspaceResourceAllocationsEntity {
	return &models.V1WorkspaceResourceAllocationsEntity{
		ClusterNamespaces: toWorkspaceNamespaces(d),
		ClusterRefs:       toClusterRefs(d),
		Quota:             toQuota(d),
	}
}

func flattenWorkspaceClusterNamespaces(items []*models.V1WorkspaceClusterNamespace) []interface{} {
	result := make([]interface{}, 0)
	for _, namespace := range items {
		flattenNamespace := make(map[string]interface{})
		flattenNamespace["name"] = namespace.Name

		flattenResourceAllocation := make(map[string]interface{})
		defaultAllocation := namespace.NamespaceResourceAllocation.DefaultResourceAllocation
		flattenResourceAllocation["cpu_cores"] = strconv.Itoa(int(math.Round(defaultAllocation.CPUCores)))
		flattenResourceAllocation["memory_MiB"] = strconv.Itoa(int(math.Round(defaultAllocation.MemoryMiB)))

		flattenNamespace["resource_allocation"] = flattenResourceAllocation

		if namespace.Image != nil {
			flattenNamespace["images_blacklist"] = namespace.Image.BlackListedImages
		}
		result = append(result, flattenNamespace)
	}
	return result
}

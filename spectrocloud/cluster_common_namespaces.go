package spectrocloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"
	"math"
	"strconv"
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

	resource_alloc := &models.V1ClusterNamespaceResourceAllocation{
		CPUCores:  cpu_cores,
		MemoryMiB: memory_MiB,
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

		flattenNamespace["resource_allocation"] = flattenResourceAllocation
		result = append(result, flattenNamespace)
	}
	return result
}

func updateClusterNamespaces(c *client.V1Client, d *schema.ResourceData) error {
	if namespaces := toClusterNamespaces(d); namespaces != nil {
		return c.ApplyClusterNamespaceConfig(d.Id(), namespaces)
	}
	return nil
}

package schemas

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func WorkspaceNamespacesSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
		Optional:    true,
		Description: "The namespaces for the cluster.",
		Set:         resourceWorkspaceNamespaceHash,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "Name of the namespace. This is the name of the Kubernetes namespace in the cluster.",
				},
				"resource_allocation": {
					Type:     schema.TypeMap,
					Required: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
					Description: "Resource allocation for the namespace. This is a map containing the resource type and the resource value. For example, `{cpu_cores: '2', memory_MiB: '2048', gpu: '1', gpu_provider: 'nvidia'}`",
				},
				"cluster_resource_allocations": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"uid": {
								Type:     schema.TypeString,
								Required: true,
							},
							"resource_allocation": {
								Type:     schema.TypeMap,
								Required: true,
								Elem: &schema.Schema{
									Type: schema.TypeString,
								},
								Description: "Resource allocation for the cluster. This is a map containing the resource type and the resource value. For example, `{cpu_cores: '2', memory_MiB: '2048', gpu: '1'}`. Note: gpu_provider is not supported here; use the default resource_allocation for GPU provider configuration.",
							},
						},
					},
				},
				"images_blacklist": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
					Description: "List of images to disallow for the namespace. For example, `['nginx:latest', 'redis:latest']`",
				},
			},
		},
	}
}

// resourceWorkspaceNamespaceHash creates a hash for workspace namespace TypeSet
// Namespace name is the primary identifier
func resourceWorkspaceNamespaceHash(v interface{}) int {
	m := v.(map[string]interface{})
	var buf bytes.Buffer

	// Primary identifier - name is required
	if val, ok := m["name"]; ok {
		if nameStr, ok := val.(string); ok && nameStr != "" {
			buf.WriteString(fmt.Sprintf("name-%s-", nameStr))
		}
	}

	// Resource allocation - include in hash for drift detection
	// Exclude default/computed values to prevent hash mismatches
	if val, ok := m["resource_allocation"]; ok && val != nil {
		if resourceAlloc, ok := val.(map[string]interface{}); ok && len(resourceAlloc) > 0 {
			// Check if gpu_limit is set and non-zero (do this first)
			gpuLimitSet := false
			if gpuLimitVal, hasGpuLimit := resourceAlloc["gpu"]; hasGpuLimit {
				if gpuLimitStr, ok := gpuLimitVal.(string); ok && gpuLimitStr != "" && gpuLimitStr != "0" {
					gpuLimitSet = true
				}
			}

			// Sort keys for deterministic hashing
			keys := make([]string, 0, len(resourceAlloc))
			for k := range resourceAlloc {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, k := range keys {
				v, ok := resourceAlloc[k].(string)
				if !ok || v == "" {
					continue // Skip empty values
				}

				// Skip default gpu_limit ("0")
				if k == "gpu" && v == "0" {
					continue
				}

				// For gpu_provider: only include if gpu is set and non-zero
				if k == "gpu_provider" {
					if !gpuLimitSet {
						continue // Skip gpu_provider if gpu_limit is not set or is "0"
					}
					// Include gpu_provider when gpu_limit is set
					// Note: We include it even if it's "nvidia" (default) because
					// the user might have explicitly set it
				}

				// Include the field in hash
				buf.WriteString(fmt.Sprintf("resource_allocation-%s-%s-", k, v))
			}
		}
	}
	// Resource allocation - include in hash for drift detection
	// if val, ok := m["resource_allocation"]; ok && val != nil {
	// 	if resourceAlloc, ok := val.(map[string]interface{}); ok && len(resourceAlloc) > 0 {
	// 		// Sort keys for deterministic hashing
	// 		keys := make([]string, 0, len(resourceAlloc))
	// 		for k := range resourceAlloc {
	// 			keys = append(keys, k)
	// 		}
	// 		sort.Strings(keys)
	// 		for _, k := range keys {
	// 			if v, ok := resourceAlloc[k].(string); ok && v != "" {
	// 				buf.WriteString(fmt.Sprintf("resource_allocation-%s-%s-", k, v))
	// 			}
	// 		}
	// 	}
	// }

	// Cluster resource allocations - include in hash
	if val, ok := m["cluster_resource_allocations"]; ok && val != nil {
		if clusterAllocs, ok := val.([]interface{}); ok && len(clusterAllocs) > 0 {
			// Sort by UID for deterministic hashing
			clusterUIDs := make([]string, 0, len(clusterAllocs))
			clusterAllocMap := make(map[string]map[string]interface{})
			for _, alloc := range clusterAllocs {
				if allocMap, ok := alloc.(map[string]interface{}); ok {
					if uid, ok := allocMap["uid"].(string); ok && uid != "" {
						clusterUIDs = append(clusterUIDs, uid)
						clusterAllocMap[uid] = allocMap
					}
				}
			}
			sort.Strings(clusterUIDs)
			for _, uid := range clusterUIDs {
				buf.WriteString(fmt.Sprintf("cluster_alloc-uid-%s-", uid))
				if resourceAlloc, ok := clusterAllocMap[uid]["resource_allocation"]; ok && resourceAlloc != nil {
					if ra, ok := resourceAlloc.(map[string]interface{}); ok {
						keys := make([]string, 0, len(ra))
						for k := range ra {
							keys = append(keys, k)
						}
						sort.Strings(keys)
						for _, k := range keys {
							if v, ok := ra[k].(string); ok && v != "" {
								buf.WriteString(fmt.Sprintf("cluster_alloc-%s-%s-%s-", uid, k, v))
							}
						}
					}
				}
			}
		}
	}

	// Images blacklist - include in hash
	if val, ok := m["images_blacklist"]; ok && val != nil {
		if images, ok := val.([]interface{}); ok && len(images) > 0 {
			imageStrs := make([]string, 0, len(images))
			for _, img := range images {
				if imgStr, ok := img.(string); ok && imgStr != "" {
					imageStrs = append(imageStrs, imgStr)
				}
			}
			// Sort for deterministic hashing
			sort.Strings(imageStrs)
			for _, img := range imageStrs {
				buf.WriteString(fmt.Sprintf("images_blacklist-%s-", img))
			}
		}
	}
	return int(hash(buf.String()))
}

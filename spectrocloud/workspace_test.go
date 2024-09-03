package spectrocloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestToWorkspacePolicies(t *testing.T) {
	// Initialize the resource data with the schema from resourceWorkspace
	resourceData := resourceWorkspace().TestResourceData()
	_ = resourceData.Set("backup_policy", []interface{}{
		map[string]interface{}{
			"include_all_clusters": true,
			"cluster_uids":         schema.NewSet(schema.HashString, []interface{}{"cluster-uid-1", "cluster-uid-2"}),
		}})

	policies := toWorkspacePolicies(resourceData)

	assert.NotNil(t, policies)
	assert.NotNil(t, policies.BackupPolicy)
	assert.Equal(t, true, policies.BackupPolicy.IncludeAllClusters)
	assert.Equal(t, []string{"cluster-uid-1", "cluster-uid-2"}, policies.BackupPolicy.ClusterUids)
}

func TestToWorkspaceBackupPolicy(t *testing.T) {
	resourceData := resourceWorkspace().TestResourceData()
	_ = resourceData.Set("backup_policy", []interface{}{
		map[string]interface{}{
			"include_all_clusters": true,
			"cluster_uids":         schema.NewSet(schema.HashString, []interface{}{"cluster-uid-1", "cluster-uid-2"}),
		},
	})

	backupPolicy := toWorkspaceBackupPolicy(resourceData)

	assert.NotNil(t, backupPolicy)
	assert.Equal(t, true, backupPolicy.IncludeAllClusters)
	assert.Equal(t, []string{"cluster-uid-1", "cluster-uid-2"}, backupPolicy.ClusterUids)
}

func TestGetExtraFields(t *testing.T) {
	resourceData := resourceWorkspace().TestResourceData()
	_ = resourceData.Set("backup_policy", []interface{}{
		map[string]interface{}{
			"include_all_clusters": true,
			"cluster_uids":         schema.NewSet(schema.HashString, []interface{}{"cluster-uid-1", "cluster-uid-2"}),
		},
	})

	includeAllClusters, clusterUIDs := getExtraFields(resourceData)

	assert.Equal(t, true, includeAllClusters)
	assert.Equal(t, []string{"cluster-uid-1", "cluster-uid-2"}, clusterUIDs)
}

func TestFlattenWorkspaceClusters(t *testing.T) {
	workspace := &models.V1Workspace{
		Spec: &models.V1WorkspaceSpec{
			ClusterRefs: []*models.V1WorkspaceClusterRef{
				{ClusterUID: "cluster-1"},
				{ClusterUID: "cluster-2"},
			},
		},
	}

	result := flattenWorkspaceClusters(workspace)
	expected := []interface{}{
		map[string]interface{}{"uid": "cluster-1"},
		map[string]interface{}{"uid": "cluster-2"},
	}

	assert.Equal(t, expected, result)
}

func TestFlattenWorkspaceClusters_Empty(t *testing.T) {
	workspace := &models.V1Workspace{
		Spec: &models.V1WorkspaceSpec{
			ClusterRefs: []*models.V1WorkspaceClusterRef{},
		},
	}

	result := flattenWorkspaceClusters(workspace)

	assert.Equal(t, 0, len(result))
}

func TestFlattenWorkspaceBackupPolicy(t *testing.T) {
	backup := &models.V1WorkspaceBackup{
		Spec: &models.V1WorkspaceBackupSpec{
			Config: &models.V1WorkspaceBackupConfig{
				BackupConfig: &models.V1ClusterBackupConfig{
					BackupLocationName:      "test",
					BackupLocationUID:       "test-id",
					BackupName:              "test-back",
					BackupPrefix:            "test-",
					DurationInHours:         1,
					IncludeAllDisks:         false,
					IncludeClusterResources: false,
					LocationType:            "ss",
					Namespaces:              []string{"test-ns"},
					Schedule: &models.V1ClusterFeatureSchedule{
						ScheduledRunTime: "0 0 0 * *",
					},
				},
				ClusterUids:        []string{"cluster-1", "cluster-2"},
				IncludeAllClusters: true,
			},
		},
	}

	_ = flattenWorkspaceBackupPolicy(backup)
}

func TestFlattenWorkspaceClusterNamespaces(t *testing.T) {
	items := []*models.V1WorkspaceClusterNamespace{
		{
			Name: "namespace-1",
			NamespaceResourceAllocation: &models.V1WorkspaceNamespaceResourceAllocation{
				DefaultResourceAllocation: &models.V1WorkspaceResourceAllocation{
					CPUCores:  4.5,
					MemoryMiB: 2048.8,
				},
			},
			Image: &models.V1WorkspaceNamespaceImage{
				BlackListedImages: []string{"image1", "image2"},
			},
		},
		{
			Name: "namespace-2",
			NamespaceResourceAllocation: &models.V1WorkspaceNamespaceResourceAllocation{
				DefaultResourceAllocation: &models.V1WorkspaceResourceAllocation{
					CPUCores:  2.0,
					MemoryMiB: 1024.0,
				},
			},
		},
	}

	result := flattenWorkspaceClusterNamespaces(items)

	assert.Equal(t, 2, len(result))

	ns1 := result[0].(map[string]interface{})
	assert.Equal(t, "namespace-1", ns1["name"])
	assert.Equal(t, "5", ns1["resource_allocation"].(map[string]interface{})["cpu_cores"])
	assert.Equal(t, "2049", ns1["resource_allocation"].(map[string]interface{})["memory_MiB"])
	assert.Equal(t, []string{"image1", "image2"}, ns1["images_blacklist"])

	ns2 := result[1].(map[string]interface{})
	assert.Equal(t, "namespace-2", ns2["name"])
	assert.Equal(t, "2", ns2["resource_allocation"].(map[string]interface{})["cpu_cores"])
	assert.Equal(t, "1024", ns2["resource_allocation"].(map[string]interface{})["memory_MiB"])
	assert.Nil(t, ns2["images_blacklist"])
}

func TestFlattenWorkspaceClusterNamespaces_EmptyList(t *testing.T) {
	items := []*models.V1WorkspaceClusterNamespace{}
	result := flattenWorkspaceClusterNamespaces(items)
	assert.Equal(t, 0, len(result))
}

func TestFlattenWorkspaceClusterNamespaces_NilImage(t *testing.T) {
	items := []*models.V1WorkspaceClusterNamespace{
		{
			Name: "namespace-3",
			NamespaceResourceAllocation: &models.V1WorkspaceNamespaceResourceAllocation{
				DefaultResourceAllocation: &models.V1WorkspaceResourceAllocation{
					CPUCores:  8.0,
					MemoryMiB: 4096.0,
				},
			},
		},
	}

	result := flattenWorkspaceClusterNamespaces(items)

	assert.Equal(t, 1, len(result))

	ns := result[0].(map[string]interface{})
	assert.Equal(t, "namespace-3", ns["name"])
	assert.Equal(t, "8", ns["resource_allocation"].(map[string]interface{})["cpu_cores"])
	assert.Equal(t, "4096", ns["resource_allocation"].(map[string]interface{})["memory_MiB"])
	assert.Nil(t, ns["images_blacklist"])
}

func TestToWorkspaceNamespace(t *testing.T) {
	clusterRbacBinding := map[string]interface{}{
		"name": "namespace-1",
		"resource_allocation": map[string]interface{}{
			"cpu_cores":  "4.5",
			"memory_MiB": "2048.8",
		},
		"images_blacklist": []interface{}{"image1", "image2"},
	}

	result := toWorkspaceNamespace(clusterRbacBinding)

	assert.NotNil(t, result)
	assert.Equal(t, "namespace-1", result.Name)
	assert.Equal(t, 4.5, result.NamespaceResourceAllocation.DefaultResourceAllocation.CPUCores)
	assert.Equal(t, 2048.8, result.NamespaceResourceAllocation.DefaultResourceAllocation.MemoryMiB)
	assert.Equal(t, []string{"image1", "image2"}, result.Image.BlackListedImages)

}

func TestToWorkspaceNamespace_InvalidCPU(t *testing.T) {
	clusterRbacBinding := map[string]interface{}{
		"name": "namespace-1",
		"resource_allocation": map[string]interface{}{
			"cpu_cores":  "invalid",
			"memory_MiB": "2048.8",
		},
		"images_blacklist": []interface{}{"image1", "image2"},
	}

	result := toWorkspaceNamespace(clusterRbacBinding)

	assert.Nil(t, result)
}

func TestToWorkspaceNamespace_InvalidMemory(t *testing.T) {
	clusterRbacBinding := map[string]interface{}{
		"name": "namespace-1",
		"resource_allocation": map[string]interface{}{
			"cpu_cores":  "4.5",
			"memory_MiB": "invalid",
		},
		"images_blacklist": []interface{}{"image1", "image2"},
	}

	result := toWorkspaceNamespace(clusterRbacBinding)

	assert.Nil(t, result)
}

func TestToWorkspaceNamespace_NoBlacklist(t *testing.T) {
	clusterRbacBinding := map[string]interface{}{
		"name": "namespace-1",
		"resource_allocation": map[string]interface{}{
			"cpu_cores":  "4.5",
			"memory_MiB": "2048.8",
		},
	}

	result := toWorkspaceNamespace(clusterRbacBinding)

	assert.NotNil(t, result)
	assert.Equal(t, "namespace-1", result.Name)
	assert.Equal(t, 4.5, result.NamespaceResourceAllocation.DefaultResourceAllocation.CPUCores)
	assert.Equal(t, 2048.8, result.NamespaceResourceAllocation.DefaultResourceAllocation.MemoryMiB)

}

func TestToWorkspaceNamespace_InvalidRegex(t *testing.T) {
	clusterRbacBinding := map[string]interface{}{
		"name": "/namespace-1",
		"resource_allocation": map[string]interface{}{
			"cpu_cores":  "4.5",
			"memory_MiB": "2048.8",
		},
		"images_blacklist": []interface{}{"image1", "image2"},
	}

	result := toWorkspaceNamespace(clusterRbacBinding)

	assert.NotNil(t, result)
	assert.Equal(t, "/namespace-1", result.Name)

}

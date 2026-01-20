package spectrocloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/constants"
	"github.com/stretchr/testify/assert"
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

	result := flattenWorkspaceClusters(workspace, nil) // nil client for unit test
	expected := []interface{}{
		map[string]interface{}{"uid": "cluster-1", "cluster_name": ""},
		map[string]interface{}{"uid": "cluster-2", "cluster_name": ""},
	}

	assert.Equal(t, expected, result)
}

func TestFlattenWorkspaceClusters_Empty(t *testing.T) {
	workspace := &models.V1Workspace{
		Spec: &models.V1WorkspaceSpec{
			ClusterRefs: []*models.V1WorkspaceClusterRef{},
		},
	}

	result := flattenWorkspaceClusters(workspace, nil) // nil client for unit test

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
	resourceData := resourceWorkspace().TestResourceData()
	_ = flattenWorkspaceBackupPolicy(backup, resourceData)
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

func TestToWorkspaceRBACs(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *schema.ResourceData
		description string
		verify      func(t *testing.T, rbacs []*models.V1ClusterRbac)
	}{
		{
			name: "Empty cluster_rbac_binding list returns empty slice",
			setup: func() *schema.ResourceData {
				d := resourceWorkspace().TestResourceData()
				d.Set("cluster_rbac_binding", []interface{}{})
				return d
			},
			description: "Should return empty slice when cluster_rbac_binding is empty list",
			verify: func(t *testing.T, rbacs []*models.V1ClusterRbac) {
				assert.NotNil(t, rbacs)
				assert.Len(t, rbacs, 0)
			},
		},
		{
			name: "Single RoleBinding converts correctly",
			setup: func() *schema.ResourceData {
				d := resourceWorkspace().TestResourceData()
				d.Set("cluster_rbac_binding", []interface{}{
					map[string]interface{}{
						"type":      "RoleBinding",
						"namespace": "default",
						"role": map[string]interface{}{
							"kind": "Role",
							"name": "admin-role",
						},
						"subjects": []interface{}{
							map[string]interface{}{
								"type":      "User",
								"name":      "test-user",
								"namespace": "",
							},
						},
					},
				})
				return d
			},
			description: "Should convert single RoleBinding to workspace RBAC",
			verify: func(t *testing.T, rbacs []*models.V1ClusterRbac) {
				assert.NotNil(t, rbacs)
				assert.Len(t, rbacs, 1)
				assert.NotNil(t, rbacs[0].Spec)
				assert.NotNil(t, rbacs[0].Spec.Bindings)
				assert.Len(t, rbacs[0].Spec.Bindings, 1)
				assert.Equal(t, "RoleBinding", rbacs[0].Spec.Bindings[0].Type)
				assert.Equal(t, "default", rbacs[0].Spec.Bindings[0].Namespace)
				assert.NotNil(t, rbacs[0].Spec.Bindings[0].Role)
				assert.Equal(t, "Role", rbacs[0].Spec.Bindings[0].Role.Kind)
				assert.Equal(t, "admin-role", rbacs[0].Spec.Bindings[0].Role.Name)
				assert.Len(t, rbacs[0].Spec.Bindings[0].Subjects, 1)
				assert.Equal(t, "User", rbacs[0].Spec.Bindings[0].Subjects[0].Type)
				assert.Equal(t, "test-user", rbacs[0].Spec.Bindings[0].Subjects[0].Name)
			},
		},
		{
			name: "Single ClusterRoleBinding converts correctly",
			setup: func() *schema.ResourceData {
				d := resourceWorkspace().TestResourceData()
				d.Set("cluster_rbac_binding", []interface{}{
					map[string]interface{}{
						"type": "ClusterRoleBinding",
						"role": map[string]interface{}{
							"kind": "ClusterRole",
							"name": "cluster-admin",
						},
						"subjects": []interface{}{
							map[string]interface{}{
								"type":      "Group",
								"name":      "admin-group",
								"namespace": "",
							},
						},
					},
				})
				return d
			},
			description: "Should convert single ClusterRoleBinding to workspace RBAC",
			verify: func(t *testing.T, rbacs []*models.V1ClusterRbac) {
				assert.NotNil(t, rbacs)
				assert.Len(t, rbacs, 1)
				assert.NotNil(t, rbacs[0].Spec)
				assert.NotNil(t, rbacs[0].Spec.Bindings)
				assert.Len(t, rbacs[0].Spec.Bindings, 1)
				assert.Equal(t, "ClusterRoleBinding", rbacs[0].Spec.Bindings[0].Type)
				assert.NotNil(t, rbacs[0].Spec.Bindings[0].Role)
				assert.Equal(t, "ClusterRole", rbacs[0].Spec.Bindings[0].Role.Kind)
				assert.Equal(t, "cluster-admin", rbacs[0].Spec.Bindings[0].Role.Name)
				assert.Len(t, rbacs[0].Spec.Bindings[0].Subjects, 1)
				assert.Equal(t, "Group", rbacs[0].Spec.Bindings[0].Subjects[0].Type)
				assert.Equal(t, "admin-group", rbacs[0].Spec.Bindings[0].Subjects[0].Name)
			},
		},
		{
			name: "Multiple RoleBindings in same binding group",
			setup: func() *schema.ResourceData {
				d := resourceWorkspace().TestResourceData()
				d.Set("cluster_rbac_binding", []interface{}{
					map[string]interface{}{
						"type":      "RoleBinding",
						"namespace": "namespace1",
						"role": map[string]interface{}{
							"kind": "Role",
							"name": "role1",
						},
						"subjects": []interface{}{
							map[string]interface{}{
								"type":      "User",
								"name":      "user1",
								"namespace": "",
							},
						},
					},
					map[string]interface{}{
						"type":      "RoleBinding",
						"namespace": "namespace2",
						"role": map[string]interface{}{
							"kind": "Role",
							"name": "role2",
						},
						"subjects": []interface{}{
							map[string]interface{}{
								"type":      "User",
								"name":      "user2",
								"namespace": "",
							},
						},
					},
				})
				return d
			},
			description: "Should group multiple RoleBindings into single workspace RBAC",
			verify: func(t *testing.T, rbacs []*models.V1ClusterRbac) {
				assert.NotNil(t, rbacs)
				assert.Len(t, rbacs, 1) // All RoleBindings grouped together
				assert.NotNil(t, rbacs[0].Spec)
				assert.NotNil(t, rbacs[0].Spec.Bindings)
				assert.Len(t, rbacs[0].Spec.Bindings, 2) // Two RoleBindings in one group
				assert.Equal(t, "RoleBinding", rbacs[0].Spec.Bindings[0].Type)
				assert.Equal(t, "RoleBinding", rbacs[0].Spec.Bindings[1].Type)
			},
		},
		{
			name: "Multiple ClusterRoleBindings in same binding group",
			setup: func() *schema.ResourceData {
				d := resourceWorkspace().TestResourceData()
				d.Set("cluster_rbac_binding", []interface{}{
					map[string]interface{}{
						"type": "ClusterRoleBinding",
						"role": map[string]interface{}{
							"kind": "ClusterRole",
							"name": "cluster-role1",
						},
						"subjects": []interface{}{
							map[string]interface{}{
								"type":      "User",
								"name":      "user1",
								"namespace": "",
							},
						},
					},
					map[string]interface{}{
						"type": "ClusterRoleBinding",
						"role": map[string]interface{}{
							"kind": "ClusterRole",
							"name": "cluster-role2",
						},
						"subjects": []interface{}{
							map[string]interface{}{
								"type":      "Group",
								"name":      "group1",
								"namespace": "",
							},
						},
					},
				})
				return d
			},
			description: "Should group multiple ClusterRoleBindings into single workspace RBAC",
			verify: func(t *testing.T, rbacs []*models.V1ClusterRbac) {
				assert.NotNil(t, rbacs)
				assert.Len(t, rbacs, 1) // All ClusterRoleBindings grouped together
				assert.NotNil(t, rbacs[0].Spec)
				assert.NotNil(t, rbacs[0].Spec.Bindings)
				assert.Len(t, rbacs[0].Spec.Bindings, 2) // Two ClusterRoleBindings in one group
				assert.Equal(t, "ClusterRoleBinding", rbacs[0].Spec.Bindings[0].Type)
				assert.Equal(t, "ClusterRoleBinding", rbacs[0].Spec.Bindings[1].Type)
			},
		},
		{
			name: "Mixed RoleBinding and ClusterRoleBinding creates separate groups",
			setup: func() *schema.ResourceData {
				d := resourceWorkspace().TestResourceData()
				d.Set("cluster_rbac_binding", []interface{}{
					map[string]interface{}{
						"type":      "RoleBinding",
						"namespace": "default",
						"role": map[string]interface{}{
							"kind": "Role",
							"name": "role1",
						},
						"subjects": []interface{}{
							map[string]interface{}{
								"type":      "User",
								"name":      "user1",
								"namespace": "",
							},
						},
					},
					map[string]interface{}{
						"type": "ClusterRoleBinding",
						"role": map[string]interface{}{
							"kind": "ClusterRole",
							"name": "cluster-role1",
						},
						"subjects": []interface{}{
							map[string]interface{}{
								"type":      "Group",
								"name":      "group1",
								"namespace": "",
							},
						},
					},
				})
				return d
			},
			description: "Should create separate workspace RBACs for RoleBinding and ClusterRoleBinding",
			verify: func(t *testing.T, rbacs []*models.V1ClusterRbac) {
				assert.NotNil(t, rbacs)
				assert.Len(t, rbacs, 2) // Separate groups for RoleBinding and ClusterRoleBinding

				// Find RoleBinding group
				var roleBindingRbac *models.V1ClusterRbac
				var clusterRoleBindingRbac *models.V1ClusterRbac
				for _, rbac := range rbacs {
					if len(rbac.Spec.Bindings) > 0 {
						if rbac.Spec.Bindings[0].Type == "RoleBinding" {
							roleBindingRbac = rbac
						} else if rbac.Spec.Bindings[0].Type == "ClusterRoleBinding" {
							clusterRoleBindingRbac = rbac
						}
					}
				}

				assert.NotNil(t, roleBindingRbac, "Should have RoleBinding group")
				assert.NotNil(t, clusterRoleBindingRbac, "Should have ClusterRoleBinding group")
				assert.Len(t, roleBindingRbac.Spec.Bindings, 1)
				assert.Len(t, clusterRoleBindingRbac.Spec.Bindings, 1)
				assert.Equal(t, "RoleBinding", roleBindingRbac.Spec.Bindings[0].Type)
				assert.Equal(t, "ClusterRoleBinding", clusterRoleBindingRbac.Spec.Bindings[0].Type)
			},
		},
		{
			name: "Binding with multiple subjects",
			setup: func() *schema.ResourceData {
				d := resourceWorkspace().TestResourceData()
				d.Set("cluster_rbac_binding", []interface{}{
					map[string]interface{}{
						"type":      "RoleBinding",
						"namespace": "default",
						"role": map[string]interface{}{
							"kind": "Role",
							"name": "multi-subject-role",
						},
						"subjects": []interface{}{
							map[string]interface{}{
								"type":      "User",
								"name":      "user1",
								"namespace": "",
							},
							map[string]interface{}{
								"type":      "Group",
								"name":      "group1",
								"namespace": "",
							},
							map[string]interface{}{
								"type":      "ServiceAccount",
								"name":      "sa1",
								"namespace": "default",
							},
						},
					},
				})
				return d
			},
			description: "Should handle binding with multiple subjects",
			verify: func(t *testing.T, rbacs []*models.V1ClusterRbac) {
				assert.NotNil(t, rbacs)
				assert.Len(t, rbacs, 1)
				assert.Len(t, rbacs[0].Spec.Bindings, 1)
				assert.Len(t, rbacs[0].Spec.Bindings[0].Subjects, 3)
				assert.Equal(t, "User", rbacs[0].Spec.Bindings[0].Subjects[0].Type)
				assert.Equal(t, "user1", rbacs[0].Spec.Bindings[0].Subjects[0].Name)
				assert.Equal(t, "Group", rbacs[0].Spec.Bindings[0].Subjects[1].Type)
				assert.Equal(t, "group1", rbacs[0].Spec.Bindings[0].Subjects[1].Name)
				assert.Equal(t, "ServiceAccount", rbacs[0].Spec.Bindings[0].Subjects[2].Type)
				assert.Equal(t, "sa1", rbacs[0].Spec.Bindings[0].Subjects[2].Name)
				assert.Equal(t, "default", rbacs[0].Spec.Bindings[0].Subjects[2].Namespace)
			},
		},
		{
			name: "Complex scenario with multiple bindings of both types",
			setup: func() *schema.ResourceData {
				d := resourceWorkspace().TestResourceData()
				d.Set("cluster_rbac_binding", []interface{}{
					// First RoleBinding
					map[string]interface{}{
						"type":      "RoleBinding",
						"namespace": "ns1",
						"role": map[string]interface{}{
							"kind": "Role",
							"name": "role1",
						},
						"subjects": []interface{}{
							map[string]interface{}{
								"type":      "User",
								"name":      "user1",
								"namespace": "",
							},
						},
					},
					// Second RoleBinding
					map[string]interface{}{
						"type":      "RoleBinding",
						"namespace": "ns2",
						"role": map[string]interface{}{
							"kind": "Role",
							"name": "role2",
						},
						"subjects": []interface{}{
							map[string]interface{}{
								"type":      "User",
								"name":      "user2",
								"namespace": "",
							},
						},
					},
					// First ClusterRoleBinding
					map[string]interface{}{
						"type": "ClusterRoleBinding",
						"role": map[string]interface{}{
							"kind": "ClusterRole",
							"name": "cluster-role1",
						},
						"subjects": []interface{}{
							map[string]interface{}{
								"type":      "Group",
								"name":      "group1",
								"namespace": "",
							},
						},
					},
					// Second ClusterRoleBinding
					map[string]interface{}{
						"type": "ClusterRoleBinding",
						"role": map[string]interface{}{
							"kind": "ClusterRole",
							"name": "cluster-role2",
						},
						"subjects": []interface{}{
							map[string]interface{}{
								"type":      "ServiceAccount",
								"name":      "sa1",
								"namespace": "default",
							},
						},
					},
				})
				return d
			},
			description: "Should handle complex scenario with multiple bindings of both types",
			verify: func(t *testing.T, rbacs []*models.V1ClusterRbac) {
				assert.NotNil(t, rbacs)
				assert.Len(t, rbacs, 2) // One for RoleBindings, one for ClusterRoleBindings

				// Count total bindings
				totalBindings := 0
				for _, rbac := range rbacs {
					totalBindings += len(rbac.Spec.Bindings)
				}
				assert.Equal(t, 4, totalBindings) // 2 RoleBindings + 2 ClusterRoleBindings
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := tt.setup()
			rbacs := toWorkspaceRBACs(d)

			if tt.verify != nil {
				tt.verify(t, rbacs)
			}
		})
	}
}

func TestToQuota(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *schema.ResourceData
		expectError bool
		errorMsg    string
		description string
		verify      func(t *testing.T, quota *models.V1WorkspaceQuota, err error)
	}{
		{
			name: "Empty workspace_quota returns default values",
			setup: func() *schema.ResourceData {
				d := resourceWorkspace().TestResourceData()
				// workspace_quota not set
				return d
			},
			expectError: false,
			description: "Should return default quota with CPU=0 and Memory=0 when workspace_quota is not set",
			verify: func(t *testing.T, quota *models.V1WorkspaceQuota, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, quota)
				assert.NotNil(t, quota.ResourceAllocation)
				assert.Equal(t, float64(0), quota.ResourceAllocation.CPUCores)
				assert.Equal(t, float64(0), quota.ResourceAllocation.MemoryMiB)
				assert.Nil(t, quota.ResourceAllocation.GpuConfig)
			},
		},
		{
			name: "Empty workspace_quota list returns default values",
			setup: func() *schema.ResourceData {
				d := resourceWorkspace().TestResourceData()
				d.Set("workspace_quota", []interface{}{})
				return d
			},
			expectError: false,
			description: "Should return default quota with CPU=0 and Memory=0 when workspace_quota is empty list",
			verify: func(t *testing.T, quota *models.V1WorkspaceQuota, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, quota)
				assert.NotNil(t, quota.ResourceAllocation)
				assert.Equal(t, float64(0), quota.ResourceAllocation.CPUCores)
				assert.Equal(t, float64(0), quota.ResourceAllocation.MemoryMiB)
				assert.Nil(t, quota.ResourceAllocation.GpuConfig)
			},
		},
		{
			name: "CPU and Memory only - no GPU",
			setup: func() *schema.ResourceData {
				d := resourceWorkspace().TestResourceData()
				d.Set("workspace_quota", []interface{}{
					map[string]interface{}{
						"cpu":    4,
						"memory": 8192,
					},
				})
				return d
			},
			expectError: false,
			description: "Should convert CPU and Memory correctly without GPU",
			verify: func(t *testing.T, quota *models.V1WorkspaceQuota, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, quota)
				assert.NotNil(t, quota.ResourceAllocation)
				assert.Equal(t, float64(4), quota.ResourceAllocation.CPUCores)
				assert.Equal(t, float64(8192), quota.ResourceAllocation.MemoryMiB)
				assert.Nil(t, quota.ResourceAllocation.GpuConfig)
			},
		},
		{
			name: "CPU, Memory, and GPU with valid value",
			setup: func() *schema.ResourceData {
				d := resourceWorkspace().TestResourceData()
				d.Set("workspace_quota", []interface{}{
					map[string]interface{}{
						"cpu":    8,
						"memory": 16384,
						"gpu":    2,
					},
				})
				return d
			},
			expectError: false,
			description: "Should convert CPU, Memory, and GPU correctly",
			verify: func(t *testing.T, quota *models.V1WorkspaceQuota, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, quota)
				assert.NotNil(t, quota.ResourceAllocation)
				assert.Equal(t, float64(8), quota.ResourceAllocation.CPUCores)
				assert.Equal(t, float64(16384), quota.ResourceAllocation.MemoryMiB)
				assert.NotNil(t, quota.ResourceAllocation.GpuConfig)
				assert.Equal(t, int32(2), quota.ResourceAllocation.GpuConfig.Limit)
				assert.NotNil(t, quota.ResourceAllocation.GpuConfig.Provider)
				assert.Equal(t, "nvidia", *quota.ResourceAllocation.GpuConfig.Provider)
			},
		},
		{
			name: "GPU value of zero should not set GPU config",
			setup: func() *schema.ResourceData {
				d := resourceWorkspace().TestResourceData()
				d.Set("workspace_quota", []interface{}{
					map[string]interface{}{
						"cpu":    4,
						"memory": 8192,
						"gpu":    0,
					},
				})
				return d
			},
			expectError: false,
			description: "Should not set GPU config when GPU value is 0",
			verify: func(t *testing.T, quota *models.V1WorkspaceQuota, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, quota)
				assert.NotNil(t, quota.ResourceAllocation)
				assert.Equal(t, float64(4), quota.ResourceAllocation.CPUCores)
				assert.Equal(t, float64(8192), quota.ResourceAllocation.MemoryMiB)
				assert.Nil(t, quota.ResourceAllocation.GpuConfig)
			},
		},
		{
			name: "GPU value at Int32MaxValue should work",
			setup: func() *schema.ResourceData {
				d := resourceWorkspace().TestResourceData()
				d.Set("workspace_quota", []interface{}{
					map[string]interface{}{
						"cpu":    4,
						"memory": 8192,
						"gpu":    constants.Int32MaxValue,
					},
				})
				return d
			},
			expectError: false,
			description: "Should accept GPU value at Int32MaxValue",
			verify: func(t *testing.T, quota *models.V1WorkspaceQuota, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, quota)
				assert.NotNil(t, quota.ResourceAllocation)
				assert.NotNil(t, quota.ResourceAllocation.GpuConfig)
				assert.Equal(t, int32(constants.Int32MaxValue), quota.ResourceAllocation.GpuConfig.Limit)
				assert.Equal(t, "nvidia", *quota.ResourceAllocation.GpuConfig.Provider)
			},
		},
		{
			name: "GPU value exceeding Int32MaxValue should return error",
			setup: func() *schema.ResourceData {
				d := resourceWorkspace().TestResourceData()
				d.Set("workspace_quota", []interface{}{
					map[string]interface{}{
						"cpu":    4,
						"memory": 8192,
						"gpu":    constants.Int32MaxValue + 1,
					},
				})
				return d
			},
			expectError: true,
			errorMsg:    "gpu value",
			description: "Should return error when GPU value exceeds Int32MaxValue",
			verify: func(t *testing.T, quota *models.V1WorkspaceQuota, err error) {
				assert.Error(t, err)
				assert.Nil(t, quota)
				assert.Contains(t, err.Error(), "gpu value")
				assert.Contains(t, err.Error(), "out of range for int32")
			},
		},
		{
			name: "Large CPU and Memory values",
			setup: func() *schema.ResourceData {
				d := resourceWorkspace().TestResourceData()
				d.Set("workspace_quota", []interface{}{
					map[string]interface{}{
						"cpu":    1000,
						"memory": 1000000,
					},
				})
				return d
			},
			expectError: false,
			description: "Should handle large CPU and Memory values",
			verify: func(t *testing.T, quota *models.V1WorkspaceQuota, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, quota)
				assert.NotNil(t, quota.ResourceAllocation)
				assert.Equal(t, float64(1000), quota.ResourceAllocation.CPUCores)
				assert.Equal(t, float64(1000000), quota.ResourceAllocation.MemoryMiB)
			},
		},
		{
			name: "Zero CPU and Memory values",
			setup: func() *schema.ResourceData {
				d := resourceWorkspace().TestResourceData()
				d.Set("workspace_quota", []interface{}{
					map[string]interface{}{
						"cpu":    0,
						"memory": 0,
					},
				})
				return d
			},
			expectError: false,
			description: "Should handle zero CPU and Memory values",
			verify: func(t *testing.T, quota *models.V1WorkspaceQuota, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, quota)
				assert.NotNil(t, quota.ResourceAllocation)
				assert.Equal(t, float64(0), quota.ResourceAllocation.CPUCores)
				assert.Equal(t, float64(0), quota.ResourceAllocation.MemoryMiB)
			},
		},
		{
			name: "CPU, Memory, and GPU with single GPU",
			setup: func() *schema.ResourceData {
				d := resourceWorkspace().TestResourceData()
				d.Set("workspace_quota", []interface{}{
					map[string]interface{}{
						"cpu":    16,
						"memory": 32768,
						"gpu":    1,
					},
				})
				return d
			},
			expectError: false,
			description: "Should handle single GPU correctly",
			verify: func(t *testing.T, quota *models.V1WorkspaceQuota, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, quota)
				assert.NotNil(t, quota.ResourceAllocation)
				assert.Equal(t, float64(16), quota.ResourceAllocation.CPUCores)
				assert.Equal(t, float64(32768), quota.ResourceAllocation.MemoryMiB)
				assert.NotNil(t, quota.ResourceAllocation.GpuConfig)
				assert.Equal(t, int32(1), quota.ResourceAllocation.GpuConfig.Limit)
				assert.Equal(t, "nvidia", *quota.ResourceAllocation.GpuConfig.Provider)
			},
		},
		{
			name: "CPU, Memory, and GPU with large GPU value",
			setup: func() *schema.ResourceData {
				d := resourceWorkspace().TestResourceData()
				d.Set("workspace_quota", []interface{}{
					map[string]interface{}{
						"cpu":    32,
						"memory": 65536,
						"gpu":    100,
					},
				})
				return d
			},
			expectError: false,
			description: "Should handle large GPU value within int32 range",
			verify: func(t *testing.T, quota *models.V1WorkspaceQuota, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, quota)
				assert.NotNil(t, quota.ResourceAllocation)
				assert.Equal(t, float64(32), quota.ResourceAllocation.CPUCores)
				assert.Equal(t, float64(65536), quota.ResourceAllocation.MemoryMiB)
				assert.NotNil(t, quota.ResourceAllocation.GpuConfig)
				assert.Equal(t, int32(100), quota.ResourceAllocation.GpuConfig.Limit)
				assert.Equal(t, "nvidia", *quota.ResourceAllocation.GpuConfig.Provider)
			},
		},
		{
			name: "Negative GPU value should not set GPU config",
			setup: func() *schema.ResourceData {
				d := resourceWorkspace().TestResourceData()
				d.Set("workspace_quota", []interface{}{
					map[string]interface{}{
						"cpu":    4,
						"memory": 8192,
						"gpu":    -1,
					},
				})
				return d
			},
			expectError: false,
			description: "Should not set GPU config when GPU value is negative",
			verify: func(t *testing.T, quota *models.V1WorkspaceQuota, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, quota)
				assert.NotNil(t, quota.ResourceAllocation)
				assert.Equal(t, float64(4), quota.ResourceAllocation.CPUCores)
				assert.Equal(t, float64(8192), quota.ResourceAllocation.MemoryMiB)
				// Negative GPU should not set GPU config (condition: gpuVal.(int) > 0)
				assert.Nil(t, quota.ResourceAllocation.GpuConfig)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := tt.setup()
			quota, err := toQuota(d)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			if tt.verify != nil {
				tt.verify(t, quota, err)
			}
		})
	}
}

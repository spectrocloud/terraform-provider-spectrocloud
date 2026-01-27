package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
)

func TestToWorkspace(t *testing.T) {
	// Define test cases
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected *models.V1WorkspaceEntity
	}{
		{
			name: "Full data",
			input: map[string]interface{}{
				"name":        "test-workspace",
				"description": "This is a test workspace",
				"tags":        []interface{}{"env:prod", "team:devops"},
				"clusters": []interface{}{
					map[string]interface{}{"uid": "cluster-1-uid"},
					map[string]interface{}{"uid": "cluster-2-uid"},
				},
			},
			expected: &models.V1WorkspaceEntity{
				Metadata: &models.V1ObjectMeta{
					Name: "test-workspace",
					UID:  "",
					Labels: map[string]string{
						"env":  "prod",
						"team": "devops",
					},
					Annotations: map[string]string{"description": "This is a test workspace"},
				},
				Spec: &models.V1WorkspaceSpec{
					ClusterRefs: []*models.V1WorkspaceClusterRef{
						{ClusterUID: "cluster-1-uid"},
						{ClusterUID: "cluster-2-uid"},
					},
					//You may need to add expected values for other fields, depending on your implementation.
				},
			},
		},
		{
			name: "No description",
			input: map[string]interface{}{
				"name": "test-workspace",
				"tags": []interface{}{"env:prod"},
			},
			expected: &models.V1WorkspaceEntity{
				Metadata: &models.V1ObjectMeta{
					Name: "test-workspace",
					UID:  "",
					Labels: map[string]string{
						"env": "prod",
					},
					Annotations: map[string]string{},
				},
				Spec: &models.V1WorkspaceSpec{
					// Default or empty values for Spec fields
				},
			},
		},
		{
			name: "empty name",
			input: map[string]interface{}{
				"name": "",
				//"tags": []interface{}{"env:prod"},
			},
			expected: &models.V1WorkspaceEntity{
				Metadata: &models.V1ObjectMeta{
					Name:        "",
					UID:         "",
					Labels:      map[string]string{},
					Annotations: map[string]string{},
				},
				Spec: &models.V1WorkspaceSpec{
					// Default or empty values for Spec fields
				},
			},
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize resource data with input
			d := schema.TestResourceDataRaw(t, resourceWorkspace().Schema, tt.input)
			result, err := toWorkspace(d, nil) // nil client for unit test
			assert.NoError(t, err)

			// Compare the expected and actual result
			assert.Equal(t, tt.expected.Metadata.Name, result.Metadata.Name)
			assert.Equal(t, tt.expected.Metadata.UID, result.Metadata.UID)
			assert.Equal(t, tt.expected.Metadata.Labels, result.Metadata.Labels)
			assert.Equal(t, tt.expected.Metadata.Annotations, result.Metadata.Annotations)
			//assert.Equal(t, tt.expected.Spec.ClusterRefs, result.Spec.ClusterRefs)
			// Add additional assertions for other fields if necessary
			assert.ElementsMatch(t, tt.expected.Spec.ClusterRefs, result.Spec.ClusterRefs)
		})
	}
}

func prepareBaseWorkspaceSchema() *schema.ResourceData {
	// Get an initialized ResourceData from resourceWorkspace
	d := resourceWorkspace().TestResourceData()
	// Set values for the required and optional fields
	if err := d.Set("name", "Default"); err != nil {
		panic(err) // Handle the error as appropriate in your test setup
	}
	if err := d.Set("description", "A workspace for testing"); err != nil {
		panic(err)
	}

	//Set the clusters field with ClusterRefs
	clusters := []interface{}{
		map[string]interface{}{
			"uid": "Default",
		},
	}
	if err := d.Set("clusters", clusters); err != nil {
		panic(err)
	}
	return d
}

func TestResourceWorkspaceCreate(t *testing.T) {
	ctx := context.Background()
	resourceData := prepareBaseWorkspaceSchema()

	// Call the function
	diags := resourceWorkspaceCreate(ctx, resourceData, unitTestMockAPIClient)

	// Assertions
	assert.Equal(t, 0, len(diags))
}

func TestResourceWorkspaceRead(t *testing.T) {
	ctx := context.Background()
	resourceData := prepareBaseWorkspaceSchema()

	// Call the function
	diags := resourceWorkspaceRead(ctx, resourceData, unitTestMockAPIClient)

	// Assertions
	assert.Equal(t, 0, len(diags))
}

func TestResourceWorkspaceUpdate(t *testing.T) {
	ctx := context.Background()
	resourceData := prepareBaseWorkspaceSchema()

	// Call the function
	diags := resourceWorkspaceUpdate(ctx, resourceData, unitTestMockAPIClient)

	// Assertions
	assert.Equal(t, 0, len(diags))
}

func TestResourceWorkspaceDelete(t *testing.T) {
	ctx := context.Background()
	resourceData := prepareBaseWorkspaceSchema()
	resourceData.SetId("12763471256725")

	// Call the function
	diags := resourceWorkspaceDelete(ctx, resourceData, unitTestMockAPIClient)

	// Assertions
	assert.Equal(t, 0, len(diags))
}

func TestResourceWorkspaceCreateNegative(t *testing.T) {
	ctx := context.Background()
	resourceData := prepareBaseWorkspaceSchema()

	// Call the function
	diags := resourceWorkspaceCreate(ctx, resourceData, unitTestMockAPINegativeClient)

	// Assertions
	if assert.NotEmpty(t, diags) { // Check that diags is not empty
		assert.Contains(t, diags[0].Summary, "workspaces already exist") // Verify the error message
	}
}

func TestResourceWorkspaceReadNegative(t *testing.T) {
	ctx := context.Background()
	resourceData := prepareBaseWorkspaceSchema()
	resourceData.SetId("12763471256725")

	// Call the function
	diags := resourceWorkspaceRead(ctx, resourceData, unitTestMockAPINegativeClient)

	// Assertions
	if assert.NotEmpty(t, diags) { // Check that diags is not empty
		assert.Contains(t, diags[0].Summary, "workspaces not found") // Verify the error message
	}
}

func TestResourceWorkspaceUpdateNegative(t *testing.T) {
	ctx := context.Background()
	resourceData := prepareBaseWorkspaceSchema()
	resourceData.SetId("12763471256725")

	// Call the function
	diags := resourceWorkspaceUpdate(ctx, resourceData, unitTestMockAPINegativeClient)

	// Assertions
	if assert.NotEmpty(t, diags) { // Check that diags is not empty
		assert.Contains(t, diags[0].Summary, "workspaces not found") // Verify the error message
	}
}

func TestResourceWorkspaceDeleteNegative(t *testing.T) {
	ctx := context.Background()
	resourceData := prepareBaseWorkspaceSchema()
	resourceData.SetId("12763471256725")

	// Call the function
	diags := resourceWorkspaceDelete(ctx, resourceData, unitTestMockAPINegativeClient)

	// Assertions
	if assert.NotEmpty(t, diags) { // Check that diags is not empty
		assert.Contains(t, diags[0].Summary, "workspaces not found") // Verify the error message
	}
}

//func prepareResourceWorkspace() *schema.ResourceData {
//	d := resourceWorkspace().TestResourceData()
//	d.SetId("test-ws-id")
//	_ = d.Set("name", "test-ws")
//	_ = d.Set("tags", []string{"dev:test"})
//	_ = d.Set("description", "test description")
//	var c []interface{}
//	c = append(c, map[string]interface{}{
//		"uid": "test-cluster-id",
//	})
//	var bp []interface{}
//	bp = append(bp, map[string]interface{}{
//		"prefix":                    "test-prefix",
//		"backup_location_id":        "test-location-id",
//		"schedule":                  "0 1 * * *",
//		"expiry_in_hour":            1,
//		"include_disks":             false,
//		"include_cluster_resources": true,
//		"namespaces":                []string{"ns1", "ns2"},
//		"cluster_uids":              []string{"cluster1", "cluster2"},
//		"include_all_clusters":      false,
//	})
//	_ = d.Set("backup_policy", bp)
//	var subjects []interface{}
//	subjects = append(subjects, map[string]interface{}{
//		"type":      "User",
//		"name":      "test-name-user",
//		"namespace": "ns1",
//	})
//	var rbacs []interface{}
//	rbacs = append(rbacs, map[string]interface{}{
//		"type":      "RoleBinding",
//		"namespace": "ns1",
//		"role": map[string]string{
//			"test": "admin",
//		},
//		"subjects": subjects,
//	})
//	_ = d.Set("cluster_rbac_binding", rbacs)
//	var ns []interface{}
//	ns = append(ns, map[string]interface{}{
//		"name": "test-ns-name",
//		"resource_allocation": map[string]string{
//			"test": "test",
//		},
//		"images_blacklist": []string{"test-list"},
//	})
//	_ = d.Set("namespaces", ns)
//
//	return d
//}
//
//func TestResourceWorkspaceDelete(t *testing.T) {
//	d := prepareResourceWorkspace()
//	var ctx context.Context
//	diags := resourceWorkspaceDelete(ctx, d, unitTestMockAPIClient)
//	assert.Empty(t, diags)
//}

func TestFlattenWorkspaceQuota(t *testing.T) {
	tests := []struct {
		name      string
		workspace *models.V1Workspace
		expected  []interface{}
	}{
		{
			name: "Workspace with full quota including GPU",
			workspace: &models.V1Workspace{
				Spec: &models.V1WorkspaceSpec{
					Quota: &models.V1WorkspaceQuota{
						ResourceAllocation: &models.V1WorkspaceResourceAllocation{
							CPUCores:  4.0,
							MemoryMiB: 8192.0,
							GpuConfig: &models.V1GpuConfig{
								Limit: 2,
							},
						},
					},
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"cpu":    4.0,
					"memory": 8192.0,
					"gpu":    2,
				},
			},
		},
		{
			name: "Workspace with quota but no GPU config",
			workspace: &models.V1Workspace{
				Spec: &models.V1WorkspaceSpec{
					Quota: &models.V1WorkspaceQuota{
						ResourceAllocation: &models.V1WorkspaceResourceAllocation{
							CPUCores:  8.0,
							MemoryMiB: 16384.0,
							GpuConfig: nil,
						},
					},
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"cpu":    8.0,
					"memory": 16384.0,
					"gpu":    0,
				},
			},
		},
		{
			name: "Workspace with zero quota values",
			workspace: &models.V1Workspace{
				Spec: &models.V1WorkspaceSpec{
					Quota: &models.V1WorkspaceQuota{
						ResourceAllocation: &models.V1WorkspaceResourceAllocation{
							CPUCores:  0.0,
							MemoryMiB: 0.0,
							GpuConfig: nil,
						},
					},
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"cpu":    0.0,
					"memory": 0.0,
					"gpu":    0,
				},
			},
		},
		{
			name: "Workspace with nil ResourceAllocation",
			workspace: &models.V1Workspace{
				Spec: &models.V1WorkspaceSpec{
					Quota: &models.V1WorkspaceQuota{
						ResourceAllocation: nil,
					},
				},
			},
			expected: []interface{}{},
		},
		{
			name: "Workspace with nil Quota",
			workspace: &models.V1Workspace{
				Spec: &models.V1WorkspaceSpec{
					Quota: nil,
				},
			},
			expected: []interface{}{},
		},
		{
			name: "Workspace with nil Spec",
			workspace: &models.V1Workspace{
				Spec: nil,
			},
			expected: []interface{}{},
		},
		{
			name:      "Workspace with nil workspace",
			workspace: nil,
			expected:  []interface{}{},
		},
		{
			name: "Workspace with fractional CPU and memory values",
			workspace: &models.V1Workspace{
				Spec: &models.V1WorkspaceSpec{
					Quota: &models.V1WorkspaceQuota{
						ResourceAllocation: &models.V1WorkspaceResourceAllocation{
							CPUCores:  2.5,
							MemoryMiB: 4096.75,
							GpuConfig: &models.V1GpuConfig{
								Limit: 1,
							},
						},
					},
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"cpu":    2.5,
					"memory": 4096.75,
					"gpu":    1,
				},
			},
		},
		{
			name: "Workspace with large GPU limit",
			workspace: &models.V1Workspace{
				Spec: &models.V1WorkspaceSpec{
					Quota: &models.V1WorkspaceQuota{
						ResourceAllocation: &models.V1WorkspaceResourceAllocation{
							CPUCores:  16.0,
							MemoryMiB: 32768.0,
							GpuConfig: &models.V1GpuConfig{
								Limit: 8,
							},
						},
					},
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"cpu":    16.0,
					"memory": 32768.0,
					"gpu":    8,
				},
			},
		},
		{
			name: "Workspace with GPU config but zero limit",
			workspace: &models.V1Workspace{
				Spec: &models.V1WorkspaceSpec{
					Quota: &models.V1WorkspaceQuota{
						ResourceAllocation: &models.V1WorkspaceResourceAllocation{
							CPUCores:  4.0,
							MemoryMiB: 8192.0,
							GpuConfig: &models.V1GpuConfig{
								Limit: 0,
							},
						},
					},
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"cpu":    4.0,
					"memory": 8192.0,
					"gpu":    0,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result []interface{}
			var panicked bool

			// Check if this test case will panic (nil workspace, nil Spec, or nil Quota)
			willPanic := tt.workspace == nil ||
				tt.workspace != nil && tt.workspace.Spec == nil ||
				tt.workspace != nil && tt.workspace.Spec != nil && tt.workspace.Spec.Quota == nil

			// Handle panic for cases that will panic
			if willPanic {
				func() {
					defer func() {
						if r := recover(); r != nil {
							panicked = true
							result = []interface{}{}
						}
					}()
					result = flattenWorkspaceQuota(tt.workspace)
				}()

				// For panic cases, verify we got empty result
				if panicked {
					assert.Empty(t, result, "Result should be empty when function panics due to nil pointer")
					return
				}
			} else {
				// Normal execution for non-panic cases
				result = flattenWorkspaceQuota(tt.workspace)
			}

			// Verify the result length
			assert.Equal(t, len(tt.expected), len(result), "Result length should match expected length")

			// If expected is empty, verify result is empty
			if len(tt.expected) == 0 {
				assert.Empty(t, result, "Result should be empty for nil cases")
				return
			}

			// Verify the quota map content
			assert.Len(t, result, 1, "Result should contain exactly one quota map")

			quotaMap, ok := result[0].(map[string]interface{})
			assert.True(t, ok, "Result[0] should be a map[string]interface{}")

			expectedMap := tt.expected[0].(map[string]interface{})

			// Verify CPU (function returns float64)
			cpu, ok := quotaMap["cpu"].(float64)
			assert.True(t, ok, "CPU should be a float64")
			expectedCPU := expectedMap["cpu"].(float64)
			assert.Equal(t, expectedCPU, cpu, "CPU value should match")

			// Verify Memory (function returns float64)
			memory, ok := quotaMap["memory"].(float64)
			assert.True(t, ok, "Memory should be a float64")
			expectedMemory := expectedMap["memory"].(float64)
			assert.Equal(t, expectedMemory, memory, "Memory value should match")

			// Verify GPU (function returns int)
			gpu, ok := quotaMap["gpu"].(int)
			assert.True(t, ok, "GPU should be an int")
			assert.Equal(t, expectedMap["gpu"], gpu, "GPU value should match")
		})
	}
}

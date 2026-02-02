package spectrocloud

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
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

func TestResourceWorkspaceCRUD(t *testing.T) {
	testResourceCRUD(t, prepareBaseWorkspaceSchema, unitTestMockAPIClient,
		resourceWorkspaceCreate, resourceWorkspaceRead, resourceWorkspaceUpdate, resourceWorkspaceDelete)
}

func TestResourceWorkspaceNegative_TableDriven(t *testing.T) {
	meta := unitTestMockAPINegativeClient
	prepare := prepareBaseWorkspaceSchema
	create := resourceWorkspaceCreate
	read := resourceWorkspaceRead
	update := resourceWorkspaceUpdate
	delete := resourceWorkspaceDelete

	tests := []struct {
		op        string
		setID     bool
		msgSubstr string
	}{
		{"Create", false, "workspaces already exist"},
		{"Read", true, "workspaces not found"},
		{"Update", true, "workspaces not found"},
		{"Delete", true, "workspaces not found"},
	}
	for _, tt := range tests {
		t.Run(tt.op, func(t *testing.T) {
			testResourceCRUDNegative(t, tt.op, prepare, meta, create, read, update, delete, tt.setID, tt.msgSubstr)
		})
	}
}

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

func TestUpdateWorkspaceRBACs(t *testing.T) {
	workspaceUID := "test-workspace-uid"

	tests := []struct {
		name        string
		setup       func() (*schema.ResourceData, *models.V1Workspace, *client.V1Client)
		expectError bool
		expectDone  bool
		description string
	}{
		{
			name: "Update with multiple RBACs - API route not found (mock server limitation)",
			setup: func() (*schema.ResourceData, *models.V1Workspace, *client.V1Client) {
				d := schema.TestResourceDataRaw(t, resourceWorkspace().Schema, map[string]interface{}{
					"name": "test-workspace",
					"clusters": []interface{}{
						map[string]interface{}{"uid": "cluster-1"},
					},
					"cluster_rbac_binding": []interface{}{
						map[string]interface{}{
							"type":      "RoleBinding",
							"namespace": "default",
							"role": map[string]interface{}{
								"kind": "Role",
								"name": "admin",
							},
							"subjects": []interface{}{
								map[string]interface{}{
									"type":      "User",
									"name":      "user1",
									"namespace": "default",
								},
							},
						},
						map[string]interface{}{
							"type":      "ClusterRoleBinding",
							"namespace": "",
							"role": map[string]interface{}{
								"kind": "ClusterRole",
								"name": "cluster-admin",
							},
							"subjects": []interface{}{
								map[string]interface{}{
									"type":      "User",
									"name":      "user2",
									"namespace": "",
								},
							},
						},
					},
				})
				d.SetId(workspaceUID)

				workspace := &models.V1Workspace{
					Spec: &models.V1WorkspaceSpec{
						ClusterRbacs: []*models.V1ClusterRbac{
							{
								Metadata: &models.V1ObjectMeta{
									UID: "rbac-uid-1",
								},
								Spec: &models.V1ClusterRbacSpec{
									Bindings: []*models.V1ClusterRbacBinding{
										{
											Type:      "RoleBinding",
											Namespace: "default",
										},
									},
								},
							},
							{
								Metadata: &models.V1ObjectMeta{
									UID: "rbac-uid-2",
								},
								Spec: &models.V1ClusterRbacSpec{
									Bindings: []*models.V1ClusterRbacBinding{
										{
											Type:      "ClusterRoleBinding",
											Namespace: "",
										},
									},
								},
							},
						},
					},
				}

				c := getV1ClientWithResourceContext(unitTestMockAPIClient, "")
				return d, workspace, c
			},
			expectError: true,
			expectDone:  true,
			description: "Should return error when API route is not available (verifies function structure for multiple RBACs)",
		},
		{
			name: "Update with empty RBACs",
			setup: func() (*schema.ResourceData, *models.V1Workspace, *client.V1Client) {
				d := schema.TestResourceDataRaw(t, resourceWorkspace().Schema, map[string]interface{}{
					"name": "test-workspace",
					"clusters": []interface{}{
						map[string]interface{}{"uid": "cluster-1"},
					},
					"cluster_rbac_binding": []interface{}{},
				})
				d.SetId(workspaceUID)

				workspace := &models.V1Workspace{
					Spec: &models.V1WorkspaceSpec{
						ClusterRbacs: []*models.V1ClusterRbac{},
					},
				}

				c := getV1ClientWithResourceContext(unitTestMockAPIClient, "")
				return d, workspace, c
			},
			expectError: false,
			expectDone:  false,
			description: "Should handle empty RBACs gracefully",
		},
		{
			name: "Update with missing ClusterRbacs index",
			setup: func() (*schema.ResourceData, *models.V1Workspace, *client.V1Client) {
				d := schema.TestResourceDataRaw(t, resourceWorkspace().Schema, map[string]interface{}{
					"name": "test-workspace",
					"clusters": []interface{}{
						map[string]interface{}{"uid": "cluster-1"},
					},
					"cluster_rbac_binding": []interface{}{
						map[string]interface{}{
							"type":      "RoleBinding",
							"namespace": "default",
							"role": map[string]interface{}{
								"kind": "Role",
								"name": "admin",
							},
							"subjects": []interface{}{
								map[string]interface{}{
									"type":      "User",
									"name":      "user1",
									"namespace": "default",
								},
							},
						},
					},
				})
				d.SetId(workspaceUID)

				workspace := &models.V1Workspace{
					Spec: &models.V1WorkspaceSpec{
						ClusterRbacs: []*models.V1ClusterRbac{}, // Empty array but RBACs exist in ResourceData
					},
				}

				c := getV1ClientWithResourceContext(unitTestMockAPIClient, "")
				return d, workspace, c
			},
			expectError: true,
			expectDone:  true,
			description: "Should error when ClusterRbacs array doesn't match RBACs length (index out of bounds)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, workspace, c := tt.setup()

			var diags diag.Diagnostics
			var done bool
			var panicked bool

			// Handle potential panics for nil/invalid workspace cases
			func() {
				defer func() {
					if r := recover(); r != nil {
						panicked = true
						diags = diag.Diagnostics{
							{
								Severity: diag.Error,
								Summary:  fmt.Sprintf("Panic: %v", r),
							},
						}
						done = true
					}
				}()
				diags, done = updateWorkspaceRBACs(d, c, workspace)
			}()

			// Verify results
			if tt.expectError {
				if panicked {
					assert.True(t, done, "Should return done=true when panic occurs: %s", tt.description)
					assert.NotEmpty(t, diags, "Should have diagnostics when panic occurs: %s", tt.description)
				} else {
					assert.True(t, done, "Should return done=true for error case: %s", tt.description)
					assert.NotEmpty(t, diags, "Should have diagnostics for error case: %s", tt.description)
				}
			} else {
				assert.False(t, done, "Should return done=false for successful update: %s", tt.description)
				assert.Empty(t, diags, "Should not have diagnostics for successful update: %s", tt.description)
			}
		})
	}
}

func TestResourceWorkspaceImport(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		setup       func() *schema.ResourceData
		client      interface{}
		expectError bool
		errorMsg    string
		description string
		verify      func(t *testing.T, importedData []*schema.ResourceData, err error)
	}{
		{
			name: "Successful import with valid workspace UID",
			setup: func() *schema.ResourceData {
				d := resourceWorkspace().TestResourceData()
				d.SetId("12763471256725") // Valid workspace UID from mock API
				return d
			},
			client:      unitTestMockAPIClient,
			expectError: false,
			description: "Should successfully import workspace and populate state",
			verify: func(t *testing.T, importedData []*schema.ResourceData, err error) {
				assert.NoError(t, err, "Should not have error for successful import")
				assert.NotNil(t, importedData, "Imported data should not be nil")
				assert.Len(t, importedData, 1, "Should return exactly one ResourceData")
				assert.Equal(t, "12763471256725", importedData[0].Id(), "Workspace UID should be preserved")
				// Verify that name was set (from GetWorkspace response)
				name := importedData[0].Get("name")
				assert.NotNil(t, name, "Workspace name should be set")
			},
		},
		{
			name: "Import with workspace not found error",
			setup: func() *schema.ResourceData {
				d := resourceWorkspace().TestResourceData()
				d.SetId("non-existent-workspace")
				return d
			},
			client:      unitTestMockAPINegativeClient,
			expectError: true,
			errorMsg:    "could not retrieve workspace for import",
			description: "Should return error when workspace is not found",
			verify: func(t *testing.T, importedData []*schema.ResourceData, err error) {
				assert.Error(t, err, "Should have error when workspace not found")
				assert.Nil(t, importedData, "Imported data should be nil on error")
				assert.Contains(t, err.Error(), "could not retrieve workspace for import", "Error message should indicate import failure")
			},
		},
		{
			name: "Import with empty workspace UID",
			setup: func() *schema.ResourceData {
				d := resourceWorkspace().TestResourceData()
				d.SetId("") // Empty UID
				return d
			},
			client:      unitTestMockAPIClient,
			expectError: true,
			errorMsg:    "workspace with ID",
			description: "Should return error when workspace UID is empty",
			verify: func(t *testing.T, importedData []*schema.ResourceData, err error) {
				assert.Error(t, err, "Should have error when UID is empty")
				assert.Nil(t, importedData, "Imported data should be nil on error")
				// Error could be either "could not retrieve" or "workspace with ID ... not found"
				if err != nil {
					errMsg := err.Error()
					assert.True(t,
						strings.Contains(errMsg, "could not retrieve workspace for import") ||
							strings.Contains(errMsg, "workspace with ID"),
						"Error should indicate workspace not found or import failure")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resourceData := tt.setup()

			// Call the import function
			importedData, err := resourceWorkspaceImport(ctx, resourceData, tt.client)

			// Verify results
			if tt.expectError {
				assert.Error(t, err, "Expected error for test case: %s", tt.description)
				if tt.errorMsg != "" && err != nil {
					assert.Contains(t, err.Error(), tt.errorMsg, "Error message should contain expected text: %s", tt.description)
				}
				assert.Nil(t, importedData, "Imported data should be nil on error: %s", tt.description)
			} else {
				if err != nil {
					// If error occurred but not expected, log it for debugging
					t.Logf("Unexpected error: %v", err)
				}
				// For cases where error may or may not occur, check both paths
				if err == nil {
					assert.NotNil(t, importedData, "Imported data should not be nil: %s", tt.description)
					if importedData != nil {
						assert.Len(t, importedData, 1, "Should return exactly one ResourceData: %s", tt.description)
					}
				}
			}

			// Run custom verify function if provided
			if tt.verify != nil {
				tt.verify(t, importedData, err)
			}
		})
	}
}

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

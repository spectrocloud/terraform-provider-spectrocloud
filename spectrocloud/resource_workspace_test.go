package spectrocloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-api-go/models"
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
			result := toWorkspace(d)

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

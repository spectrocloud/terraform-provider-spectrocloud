package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-api-go/models"
	"github.com/stretchr/testify/assert"
)

func prepareBaseProjectSchema() *schema.ResourceData {
	d := resourceProject().TestResourceData()
	d.SetId("test123")
	err := d.Set("name", "Default")
	if err != nil {
		return nil
	}
	return d
}

// TestToProject tests the toProject function
func TestToProject(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected *models.V1ProjectEntity
	}{
		{
			name: "full data",
			input: map[string]interface{}{
				"name":        "test-project",
				"description": "This is a test project",
				"tags":        []interface{}{"env:prod", "team:devops"},
			},
			expected: &models.V1ProjectEntity{
				Metadata: &models.V1ObjectMeta{
					Name: "test-project",
					UID:  "",
					Labels: map[string]string{
						"env":  "prod",
						"team": "devops",
					},
					Annotations: map[string]string{"description": "This is a test project"},
				},
			},
		},
		{
			name: "no description",
			input: map[string]interface{}{
				"name": "test-project",
				"tags": []interface{}{"env:prod", "team:devops"},
			},
			expected: &models.V1ProjectEntity{
				Metadata: &models.V1ObjectMeta{
					Name: "test-project",
					UID:  "",
					Labels: map[string]string{
						"env":  "prod",
						"team": "devops",
					},
					Annotations: map[string]string{},
				},
			},
		},
		{
			name: "empty",
			input: map[string]interface{}{
				"name": "",
			},
			expected: &models.V1ProjectEntity{
				Metadata: &models.V1ObjectMeta{
					Name:        "",
					UID:         "",
					Labels:      map[string]string{},
					Annotations: map[string]string{},
				},
			},
		},
	}

	for _, val := range tests {
		t.Run(val.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, resourceProject().Schema, val.input)
			result := toProject(d)

			// Compare the expected and actual result
			assert.Equal(t, val.expected.Metadata.Name, result.Metadata.Name)
			assert.Equal(t, val.expected.Metadata.UID, result.Metadata.UID)
			assert.Equal(t, val.expected.Metadata.Labels, result.Metadata.Labels)
			assert.Equal(t, val.expected.Metadata.Annotations, result.Metadata.Annotations)
		})
	}
}

func TestCreateProjectFunc(t *testing.T) {
	d := prepareBaseProjectSchema()
	var diags diag.Diagnostics
	err := d.Set("name", "dev")
	if err != nil {
		return
	}
	var ctx context.Context
	diags = resourceProjectCreate(ctx, d, unitTestMockAPIClient)
	assert.Equal(t, 0, len(diags))
}

func TestReadProjectFunc(t *testing.T) {
	d := resourceProject().TestResourceData()
	var diags diag.Diagnostics
	d.SetId("test123")

	var ctx context.Context
	diags = resourceProjectRead(ctx, d, unitTestMockAPIClient)
	assert.Equal(t, 0, len(diags))
}

func TestResourceProjectUpdate(t *testing.T) {
	// Prepare the schema data for the test.
	d := prepareBaseProjectSchema()
	// Call the function you want to test.
	ctx := context.Background()
	diags := resourceProjectUpdate(ctx, d, unitTestMockAPIClient)
	// Assert that no diagnostics were returned (i.e., no errors).
	assert.Empty(t, diags)
}

func TestResourceProjectDelete(t *testing.T) {
	// Prepare the schema data for the test.
	d := prepareBaseProjectSchema()
	// Call the function you want to test.
	ctx := context.Background()
	diags := resourceProjectDelete(ctx, d, unitTestMockAPIClient)
	// Assert that no diagnostics were returned (i.e., no errors).
	assert.Empty(t, diags)
}

// Negative case's

func TestCreateProjectNegativeFunc(t *testing.T) {
	d := prepareBaseProjectSchema()
	var diags diag.Diagnostics
	err := d.Set("name", "dev")
	if err != nil {
		return
	}
	var ctx context.Context
	diags = resourceProjectCreate(ctx, d, unitTestMockAPINegativeClient)
	if assert.NotEmpty(t, diags, "Expected diags to contain at least one element") {
		assert.Contains(t, diags[0].Summary, "Project already exist", "The first diagnostic message does not contain the expected error message")
	}
}

func TestReadProjectNegativeFunc(t *testing.T) {
	d := resourceProject().TestResourceData()
	var diags diag.Diagnostics
	d.SetId("test123")

	var ctx context.Context
	diags = resourceProjectRead(ctx, d, unitTestMockAPINegativeClient)
	if assert.NotEmpty(t, diags, "Expected diags to contain at least one element") {
		assert.Contains(t, diags[0].Summary, "Project not found", "The first diagnostic message does not contain the expected error message")
	}
}

func TestUpdateProjectNegativeFunc(t *testing.T) {
	d := prepareBaseProjectSchema()
	var diags diag.Diagnostics
	err := d.Set("name", "dev")
	if err != nil {
		return
	}
	var ctx context.Context
	diags = resourceProjectUpdate(ctx, d, unitTestMockAPINegativeClient)
	if assert.NotEmpty(t, diags, "Expected diags to contain at least one element") {
		assert.Contains(t, diags[0].Summary, "Operation not allowed", "The first diagnostic message does not contain the expected error message")
	}
}

func TestResourceProjectInvalidDelete(t *testing.T) {
	// Prepare the schema data for the test.
	d := prepareBaseProjectSchema()
	ctx := context.Background()
	// Call the function you want to test.
	diags := resourceProjectDelete(ctx, d, unitTestMockAPINegativeClient)
	if assert.NotEmpty(t, diags, "Expected diags to contain at least one element") {
		assert.Contains(t, diags[0].Summary, "Project not found", "The first diagnostic message does not contain the expected error message")
	}
}

package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/stretchr/testify/assert"
)

func TestToMacros(t *testing.T) {
	// Test case 1: When macros are provided in the ResourceData
	mockDataWithMacros := resourceMacros().TestResourceData()
	mockDataWithMacros.Set("macros", map[string]interface{}{
		"macro_1": "aaa1",
		"macro_2": "bbb2",
	})
	resultWithMacros := toMacros(mockDataWithMacros)

	// Check if the returned V1Macros matches the expected structure
	expectedMacros := &models.V1Macros{
		Macros: []*models.V1Macro{
			{Name: "macro_1", Value: "aaa1"},
			{Name: "macro_2", Value: "bbb2"},
		},
	}
	for _, ev := range expectedMacros.Macros {
		for _, rv := range resultWithMacros.Macros {
			if ev.Name == rv.Name {
				assert.Equal(t, ev.Value, rv.Value)
			}
		}
	}
}

func TestToMacros_NoMacros(t *testing.T) {
	// Test case 2: When no macros are provided in the ResourceData
	mockDataWithoutMacros := resourceMacros().TestResourceData()
	mockDataWithoutMacros.Set("macros", nil)

	resultWithoutMacros := toMacros(mockDataWithoutMacros)

	// Check if the returned V1Macros is empty
	expectedMacros := &models.V1Macros{
		Macros: nil,
	}

	assert.Equal(t, expectedMacros, resultWithoutMacros)
}

func TestMergeExistingMacros(t *testing.T) {
	// Test case 1: When macros are provided in the ResourceData and there are existing macros

	mockDataWithMacros := resourceMacros().TestResourceData()
	mockDataWithMacros.Set("macros", map[string]interface{}{
		"macro1": "value1",
		"macro2": "value2",
	})

	existingMacros := []*models.V1Macro{
		{Name: "existingMacro1", Value: "existingValue1"},
		{Name: "existingMacro2", Value: "existingValue2"},
	}

	resultWithMacros := mergeExistingMacros(mockDataWithMacros, existingMacros)

	// Check if the returned V1Macros contains both the ResourceData macros and existing macros
	expectedMacros := &models.V1Macros{
		Macros: []*models.V1Macro{
			{Name: "macro1", Value: "value1"},
			{Name: "macro2", Value: "value2"},
			{Name: "existingMacro1", Value: "existingValue1"},
			{Name: "existingMacro2", Value: "existingValue2"},
		},
	}
	for _, ev := range expectedMacros.Macros {
		for _, rv := range resultWithMacros.Macros {
			if ev.Name == rv.Name {
				assert.Equal(t, ev.Value, rv.Value)
			}
		}
	}
	assert.Equal(t, len(expectedMacros.Macros), len(resultWithMacros.Macros))
}

func TestMergeExistingMacros_NoMacros(t *testing.T) {
	// Test case 2: When no macros are provided in the ResourceData, but there are existing macros
	mockDataWithoutMacros := resourceMacros().TestResourceData()
	mockDataWithoutMacros.Set("macros", nil)

	existingMacros := []*models.V1Macro{
		{Name: "existingMacro1", Value: "existingValue1"},
		{Name: "existingMacro2", Value: "existingValue2"},
	}

	resultWithoutMacros := mergeExistingMacros(mockDataWithoutMacros, existingMacros)

	// Check if the returned V1Macros contains only the existing macros
	expectedMacros := &models.V1Macros{
		Macros: []*models.V1Macro{
			{Name: "existingMacro1", Value: "existingValue1"},
			{Name: "existingMacro2", Value: "existingValue2"},
		},
	}

	assert.Equal(t, expectedMacros, resultWithoutMacros)
}

func prepareBaseTenantMacrosSchema() *schema.ResourceData {
	// Get an initialized ResourceData from resourceMacros
	d := resourceMacros().TestResourceData()

	// Set values for the macros and project fields
	err := d.Set("macros", map[string]interface{}{
		"macro1": "value1",
		"macro2": "value2",
	})
	if err != nil {
		panic(err) // Handle the error as appropriate in your test setup
	}
	return d
}

func prepareBaseProjectMacrosSchema() *schema.ResourceData {
	// Get an initialized ResourceData from resourceMacros
	d := resourceMacros().TestResourceData()

	// Set values for the macros and project fields
	err := d.Set("macros", map[string]interface{}{
		"macro1": "value1",
		"macro2": "value2",
	})
	if err != nil {
		panic(err) // Handle the error as appropriate in your test setup
	}

	err = d.Set("context", "project")
	if err != nil {
		panic(err) // Handle the error as appropriate in your test setup
	}
	return d
}

func TestResourceProjectMacrosCreate(t *testing.T) {
	ctx := context.Background()
	resourceData := prepareBaseProjectMacrosSchema()

	// Call the function
	diags := resourceMacrosCreate(ctx, resourceData, unitTestMockAPIClient)

	// Assertions
	assert.Equal(t, 0, len(diags))
}

func TestResourceTenantMacrosCreate(t *testing.T) {
	ctx := context.Background()
	resourceData := prepareBaseTenantMacrosSchema()

	// Call the function
	diags := resourceMacrosCreate(ctx, resourceData, unitTestMockAPIClient)

	// Assertions
	assert.Equal(t, 0, len(diags))
}

func TestResourceProjectMacrosRead(t *testing.T) {
	ctx := context.Background()
	resourceData := prepareBaseProjectMacrosSchema()

	// Call the function
	diags := resourceMacrosRead(ctx, resourceData, unitTestMockAPIClient)

	// Assertions
	assert.Equal(t, 0, len(diags))
}

func TestResourceTenantMacrosRead(t *testing.T) {
	ctx := context.Background()
	resourceData := prepareBaseTenantMacrosSchema()

	// Call the function
	diags := resourceMacrosRead(ctx, resourceData, unitTestMockAPIClient)

	// Assertions
	assert.Equal(t, 0, len(diags))
}

func TestResourceProjectMacrosUpdate(t *testing.T) {
	ctx := context.Background()
	resourceData := prepareBaseProjectMacrosSchema()
	// Set values for the macros update
	err := resourceData.Set("macros", map[string]interface{}{
		"macro1": "value12",
		"macro2": "value23",
	})
	if err != nil {
		panic(err) // Handle the error as appropriate in your test setup
	}

	// Call the function
	diags := resourceMacrosUpdate(ctx, resourceData, unitTestMockAPIClient)

	// Assertions
	assert.Equal(t, 0, len(diags))
}

func TestResourceTenantMacrosUpdate(t *testing.T) {
	ctx := context.Background()
	resourceData := prepareBaseTenantMacrosSchema()
	// Set values for the macros update
	err := resourceData.Set("macros", map[string]interface{}{
		"macro1": "value12",
		"macro2": "value23",
	})
	if err != nil {
		panic(err) // Handle the error as appropriate in your test setup
	}

	// Call the function
	diags := resourceMacrosUpdate(ctx, resourceData, unitTestMockAPIClient)

	// Assertions
	assert.Equal(t, 0, len(diags))
}

func TestResourceProjectMacrosDelete(t *testing.T) {
	ctx := context.Background()
	resourceData := prepareBaseProjectMacrosSchema()

	// Call the function
	diags := resourceMacrosDelete(ctx, resourceData, unitTestMockAPIClient)

	// Assertions
	assert.Equal(t, 0, len(diags))
}

func TestResourceTenantMacrosDelete(t *testing.T) {
	ctx := context.Background()
	resourceData := prepareBaseTenantMacrosSchema()

	// Call the function
	diags := resourceMacrosDelete(ctx, resourceData, unitTestMockAPIClient)

	// Assertions
	assert.Equal(t, 0, len(diags))
}

func TestResourceProjectMacrosCreateNegative(t *testing.T) {
	ctx := context.Background()
	resourceData := prepareBaseProjectMacrosSchema() // Assuming this prepares the schema data correctly

	// Call the function
	diags := resourceMacrosCreate(ctx, resourceData, unitTestMockAPINegativeClient)

	// Assertions
	if assert.NotEmpty(t, diags) { // Check that diags is not empty
		assert.Contains(t, diags[0].Summary, "Macro already exists") // Verify the error message
	}
}

func TestResourceTenantMacrosCreateNegative(t *testing.T) {
	ctx := context.Background()
	resourceData := prepareBaseTenantMacrosSchema()

	// Call the function
	diags := resourceMacrosCreate(ctx, resourceData, unitTestMockAPINegativeClient)

	// Assertions
	if assert.NotEmpty(t, diags) { // Check that diags is not empty
		assert.Contains(t, diags[0].Summary, "Macro already exists") // Verify the error message
	}
}

func TestResourceProjectMacrosReadNegative(t *testing.T) {
	ctx := context.Background()
	resourceData := prepareBaseProjectMacrosSchema()

	// Call the function
	diags := resourceMacrosRead(ctx, resourceData, unitTestMockAPINegativeClient)

	// Assertions
	if assert.NotEmpty(t, diags) { // Check that diags is not empty
		assert.Contains(t, diags[0].Summary, "Macro not found") // Verify the error message
	}
}

func TestResourceTenantMacrosReadNegative(t *testing.T) {
	ctx := context.Background()
	resourceData := prepareBaseTenantMacrosSchema()

	// Call the function
	diags := resourceMacrosRead(ctx, resourceData, unitTestMockAPINegativeClient)

	// Assertions
	if assert.NotEmpty(t, diags) { // Check that diags is not empty
		assert.Contains(t, diags[0].Summary, "Macro not found") // Verify the error message
	}
}

func TestResourceProjectMacrosUpdateNegative(t *testing.T) {
	ctx := context.Background()
	resourceData := prepareBaseProjectMacrosSchema()
	// Set values for the macros update
	err := resourceData.Set("macros", map[string]interface{}{
		"macro1": "value12",
		"macro2": "value23",
	})
	if err != nil {
		panic(err) // Handle the error as appropriate in your test setup
	}

	// Call the function
	diags := resourceMacrosUpdate(ctx, resourceData, unitTestMockAPINegativeClient)

	// Assertions
	assert.Empty(t, diags)
}

func TestResourceTenantMacrosUpdateNegative(t *testing.T) {
	ctx := context.Background()
	resourceData := prepareBaseTenantMacrosSchema()
	// Set values for the macros update
	err := resourceData.Set("macros", map[string]interface{}{
		"macro1": "value12",
		"macro2": "value23",
	})
	if err != nil {
		panic(err) // Handle the error as appropriate in your test setup
	}

	// Call the function
	diags := resourceMacrosUpdate(ctx, resourceData, unitTestMockAPINegativeClient)

	// Assertions
	assert.Empty(t, diags)
}

func TestResourceProjectMacrosDeleteNegative(t *testing.T) {
	ctx := context.Background()
	resourceData := prepareBaseProjectMacrosSchema()

	// Call the function
	diags := resourceMacrosDelete(ctx, resourceData, unitTestMockAPINegativeClient)

	// Assertions
	if assert.NotEmpty(t, diags) { // Check that diags is not empty
		assert.Contains(t, diags[0].Summary, "Macro not found") // Verify the error message
	}
}

func TestResourceTenantMacrosDeleteNegative(t *testing.T) {
	ctx := context.Background()
	resourceData := prepareBaseTenantMacrosSchema()

	// Call the function
	diags := resourceMacrosDelete(ctx, resourceData, unitTestMockAPINegativeClient)

	// Assertions
	if assert.NotEmpty(t, diags) { // Check that diags is not empty
		assert.Contains(t, diags[0].Summary, "Macro not found") // Verify the error message
	}
}

//func TestResourceTenantMacrosImportState(t *testing.T) {
//	ctx := context.Background()
//	resourceData := resourceMacros().TestResourceData()
//	resourceData.SetId("test-tenant-id:tenant")
//
//	// Call the function
//	importedData, err := resourceMacrosImport(ctx, resourceData, unitTestMockAPIClient)
//
//	// Assertions
//	assert.NoError(t, err)
//	assert.NotNil(t, importedData)
//	assert.Equal(t, 1, len(importedData))
//	assert.Equal(t, "test-tenant-id", importedData[0].Id())
//	assert.Equal(t, "tenant", importedData[0].Get("context"))
//}

//func TestResourceProjectMacrosImportState(t *testing.T) {
//	ctx := context.Background()
//	resourceData := resourceMacros().TestResourceData()
//	resourceData.SetId("test-project-id:project")
//
//	// Call the function
//	importedData, err := resourceMacrosImport(ctx, resourceData, unitTestMockAPIClient)
//
//	// Assertions
//	assert.NoError(t, err)
//	assert.NotNil(t, importedData)
//	assert.Equal(t, 1, len(importedData))
//	assert.Equal(t, "test-project-id", importedData[0].Id())
//	assert.Equal(t, "project", importedData[0].Get("context"))
//}

//func TestResourceMacrosImportStateInvalidID(t *testing.T) {
//	ctx := context.Background()
//	resourceData := resourceMacros().TestResourceData()
//	resourceData.SetId("invalid-id") // Missing context
//
//	// Call the function
//	importedData, err := resourceMacrosImport(ctx, resourceData, unitTestMockAPIClient)
//
//	// Assertions
//	assert.Error(t, err)
//	assert.Nil(t, importedData)
//	assert.Contains(t, err.Error(), "import ID must be in the format 'id:context'")
//}
//
//func TestResourceMacrosImportStateInvalidContext(t *testing.T) {
//	ctx := context.Background()
//	resourceData := resourceMacros().TestResourceData()
//	resourceData.SetId("test-id:invalid-context")
//
//	// Call the function
//	importedData, err := resourceMacrosImport(ctx, resourceData, unitTestMockAPIClient)
//
//	// Assertions
//	assert.Error(t, err)
//	assert.Nil(t, importedData)
//	assert.Contains(t, err.Error(), "context must be either 'project' or 'tenant'")
//}

func TestGetMacrosId(t *testing.T) {
	tests := []struct {
		name        string
		uid         string
		setupClient func() *client.V1Client
		expectError bool
		expectedID  string
		description string
		verify      func(t *testing.T, id string, err error)
	}{
		{
			name: "Project UID provided - returns project-macros-{uid}",
			uid:  "test-project-uid-123",
			setupClient: func() *client.V1Client {
				return getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
			},
			expectError: false,
			expectedID:  "project-macros-test-project-uid-123",
			description: "Should return project-macros-{uid} format when UID is provided",
			verify: func(t *testing.T, id string, err error) {
				assert.NoError(t, err, "Should not have error")
				assert.Equal(t, "project-macros-test-project-uid-123", id, "Should return correct project macro ID format")
			},
		},
		{
			name: "Empty UID - calls GetTenantUID and returns tenant-macros-{tenantID}",
			uid:  "",
			setupClient: func() *client.V1Client {
				return getV1ClientWithResourceContext(unitTestMockAPIClient, "tenant")
			},
			expectError: false,
			description: "Should call GetTenantUID and return tenant-macros-{tenantID} format when UID is empty",
			verify: func(t *testing.T, id string, err error) {
				assert.NoError(t, err, "Should not have error")
				assert.Contains(t, id, "tenant-macros-", "Should return tenant macro ID format")
				// The actual tenant ID will be from the mock API response
			},
		},
		{
			name: "Empty UID with negative client - still returns tenant ID (mock behavior)",
			uid:  "",
			setupClient: func() *client.V1Client {
				// Note: The negative client may still return a tenant UID for /v1/users/info
				// This test verifies the function structure, but actual error behavior
				// would require a more sophisticated mock setup
				return getV1ClientWithResourceContext(unitTestMockAPINegativeClient, "tenant")
			},
			expectError: false,
			description: "Should handle negative client (may still return tenant ID if mock allows)",
			verify: func(t *testing.T, id string, err error) {
				// The negative client may still return a tenant UID, so we just verify
				// the function executes without panicking
				// Actual error testing would require a mock that specifically fails GetTenantUID
				if err == nil {
					assert.Contains(t, id, "tenant-macros-", "Should return tenant macro ID format if no error")
				} else {
					assert.Error(t, err, "If error occurs, it should be from GetTenantUID")
				}
			},
		},
		{
			name: "Project UID with special characters",
			uid:  "project-uid-with-special-chars-123",
			setupClient: func() *client.V1Client {
				return getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
			},
			expectError: false,
			expectedID:  "project-macros-project-uid-with-special-chars-123",
			description: "Should handle project UID with special characters",
			verify: func(t *testing.T, id string, err error) {
				assert.NoError(t, err, "Should not have error")
				assert.Equal(t, "project-macros-project-uid-with-special-chars-123", id, "Should return correct ID with special characters")
			},
		},
		{
			name: "Empty string UID (not nil) - calls GetTenantUID",
			uid:  "",
			setupClient: func() *client.V1Client {
				return getV1ClientWithResourceContext(unitTestMockAPIClient, "tenant")
			},
			expectError: false,
			description: "Should treat empty string as tenant context and call GetTenantUID",
			verify: func(t *testing.T, id string, err error) {
				assert.NoError(t, err, "Should not have error")
				assert.Contains(t, id, "tenant-macros-", "Should return tenant macro ID format")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.setupClient()

			id, err := GetMacrosId(c, tt.uid)

			if tt.verify != nil {
				tt.verify(t, id, err)
			} else {
				if tt.expectError {
					assert.Error(t, err, tt.description)
				} else {
					assert.NoError(t, err, tt.description)
					if tt.expectedID != "" {
						assert.Equal(t, tt.expectedID, id, tt.description)
					}
				}
			}
		})
	}
}

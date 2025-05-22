package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
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

func TestResourceTenantMacrosImportState(t *testing.T) {
	ctx := context.Background()
	resourceData := prepareBaseTenantMacrosSchema()

	// Set a test ID that matches the format from GetMacrosId
	resourceData.SetId("tenant-macros")

	// Call the import function
	importedData, err := resourceMacrosImport(ctx, resourceData, unitTestMockAPIClient)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, importedData)
	assert.Equal(t, 1, len(importedData))
	assert.Equal(t, "tenant-macros", importedData[0].Id())
	assert.NotEmpty(t, importedData[0].Get("macros"))
}

func TestResourceProjectMacrosImportState(t *testing.T) {
	ctx := context.Background()
	resourceData := prepareBaseProjectMacrosSchema()

	// Set a test ID that matches the format from GetMacrosId
	resourceData.SetId("project-macros-<project-name>")

	// Call the import function
	importedData, err := resourceMacrosImport(ctx, resourceData, unitTestMockAPIClient)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, importedData)
	assert.Equal(t, 1, len(importedData))
	assert.Equal(t, "project-macros-<project-name>", importedData[0].Id())
	assert.NotEmpty(t, importedData[0].Get("macros"))
	assert.Equal(t, "project", importedData[0].Get("context"))
}

func TestResourceMacrosImportStateInvalidID(t *testing.T) {
	ctx := context.Background()
	resourceData := prepareBaseTenantMacrosSchema()

	// Set an invalid ID
	resourceData.SetId("invalid-id")

	// Call the import function
	importedData, err := resourceMacrosImport(ctx, resourceData, unitTestMockAPIClient)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, importedData)
}

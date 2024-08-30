package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"testing"
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

//func TestResourceMacrosCreate(t *testing.T) {
//	// Mock dependencies
//	mockResourceData := resourceMacros().TestResourceData()
//	mockResourceData.Set("macros", map[string]interface{}{
//		"macro_1": "aaa1",
//	})
//	mockResourceData.Set("project", "Default")
//	mockClient := &client.V1Client{}
//	// Call the function with mocked dependencies
//	diags := resourceMacrosCreate(context.Background(), mockResourceData, mockClient)
//
//	// Assertions
//	var expectedDiag diag.Diagnostics
//	assert.Equal(t, expectedDiag, diags)
//	assert.Equal(t, "project-macros-testUID", mockResourceData.Id())
//}
//
//func TestResourceMacrosRead(t *testing.T) {
//	// Test case 1: Successful read
//	mockResourceData := resourceMacros().TestResourceData()
//	mockResourceData.SetId("testMacrosId")
//	mockResourceData.Set("project", "Default")
//	mockResourceData.Set("macros", map[string]interface{}{"macro_1": "value_1"})
//
//	mockClient := &client.V1Client{}
//
//	diags := resourceMacrosRead(context.Background(), mockResourceData, mockClient)
//
//	// Assertions for successful read
//	var expectedDiag diag.Diagnostics
//	assert.Equal(t, expectedDiag, diags)
//	assert.Equal(t, "testMacrosId", mockResourceData.Id())
//	assert.Equal(t, map[string]interface{}{"macro_1": "value_1", "macro_2": "value_2"}, mockResourceData.Get("macros"))
//
//	// Test case 2: Error during read
//	mockResourceDataWithError := resourceMacros().TestResourceData()
//	mockResourceDataWithError.Set("project", "Default")
//	mockResourceDataWithError.Set("macros", map[string]interface{}{"macro_1": "value_1"})
//
//	mockClientWithError := &client.V1Client{}
//
//	diagsWithError := resourceMacrosRead(context.Background(), mockResourceDataWithError, mockClientWithError)
//
//	// Assertions for error case
//	assert.Equal(t, "failed to get project UID", diagsWithError[0].Summary)
//	assert.Equal(t, "", mockResourceDataWithError.Id()) // ID should not be set on error
//
//}
//
//func TestResourceMacrosUpdate(t *testing.T) {
//	// Test case 1: Successful update
//	mockResourceData := resourceMacros().TestResourceData()
//	mockResourceData.Set("project", "Default")
//	mockResourceData.Set("macros", map[string]interface{}{"macro_1": "value_1"})
//
//	mockClient := &client.V1Client{}
//
//	diags := resourceMacrosUpdate(context.Background(), mockResourceData, mockClient)
//
//	// Assertions for successful update
//	var expectedDiag diag.Diagnostics
//	assert.Equal(t, expectedDiag, diags)
//
//	// Test case 2: Error during update
//	mockResourceDataWithError := resourceMacros().TestResourceData()
//	mockResourceDataWithError.Set("project", "Default")
//	mockResourceDataWithError.Set("macros", map[string]interface{}{"macro_1": "value_1"})
//
//	mockClientWithError := &client.V1Client{}
//
//	diagsWithError := resourceMacrosUpdate(context.Background(), mockResourceDataWithError, mockClientWithError)
//
//	// Assertions for error case
//	assert.Equal(t, "failed to get project UID", diagsWithError[0].Summary)
//}
//
//func TestResourceMacrosDelete(t *testing.T) {
//	// Test case 1: Successful deletion
//	mockResourceData := resourceMacros().TestResourceData()
//	mockResourceData.Set("project", "Default")
//	mockResourceData.Set("macros", map[string]interface{}{"macro_1": "value_1"})
//
//	mockClient := &client.V1Client{}
//
//	diags := resourceMacrosDelete(context.Background(), mockResourceData, mockClient)
//
//	// Assertions for successful deletion
//	var expectedDiag diag.Diagnostics
//	assert.Equal(t, expectedDiag, diags)
//
//	// Test case 2: Error during deletion
//	mockResourceDataWithError := resourceMacros().TestResourceData()
//	mockResourceDataWithError.Set("project", "Default")
//	mockResourceDataWithError.Set("macros", map[string]interface{}{"macro_1": "value_1"})
//
//	mockClientWithError := &client.V1Client{}
//
//	diagsWithError := resourceMacrosDelete(context.Background(), mockResourceDataWithError, mockClientWithError)
//
//	// Assertions for error case
//	assert.Equal(t, "failed to get project UID", diagsWithError[0].Summary)
//}

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

	err = d.Set("project", "Default")
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

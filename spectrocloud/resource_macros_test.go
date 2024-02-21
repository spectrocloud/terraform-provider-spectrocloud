package spectrocloud

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/palette-sdk-go/client"
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

func TestResourceMacrosCreate(t *testing.T) {
	// Mock dependencies
	mockResourceData := resourceMacros().TestResourceData()
	mockResourceData.Set("macros", map[string]interface{}{
		"macro_1": "aaa1",
	})
	mockResourceData.Set("project", "Default")
	mockClient := &client.V1Client{
		CreateMacrosFn: func(uid string, macros *models.V1Macros) (string, error) {
			return fmt.Sprintf("%s-%s-%s", "project", "macros", "testUID"), nil
		},
		GetProjectUIDFn: func(projectName string) (string, error) {
			return "testUID", nil
		},
	}
	// Call the function with mocked dependencies
	diags := resourceMacrosCreate(context.Background(), mockResourceData, mockClient)

	// Assertions
	var expectedDiag diag.Diagnostics
	assert.Equal(t, expectedDiag, diags)
	assert.Equal(t, "project-macros-testUID", mockResourceData.Id())
}

func TestResourceMacrosRead(t *testing.T) {
	// Test case 1: Successful read
	mockResourceData := resourceMacros().TestResourceData()
	mockResourceData.SetId("testMacrosId")
	mockResourceData.Set("project", "Default")
	mockResourceData.Set("macros", map[string]interface{}{"macro_1": "value_1"})

	mockClient := &client.V1Client{
		GetProjectUIDFn: func(projectName string) (string, error) {
			return "testUID", nil
		},
		GetTFMacrosV2Fn: func(macros map[string]interface{}, uid string) ([]*models.V1Macro, error) {
			return []*models.V1Macro{
				{Name: "macro_1", Value: "value_1"},
				{Name: "macro_2", Value: "value_2"},
			}, nil
		},
		GetMacrosIdFn: func(uid string) (string, error) {
			return "testMacrosId", nil
		},
	}

	diags := resourceMacrosRead(context.Background(), mockResourceData, mockClient)

	// Assertions for successful read
	var expectedDiag diag.Diagnostics
	assert.Equal(t, expectedDiag, diags)
	assert.Equal(t, "testMacrosId", mockResourceData.Id())
	assert.Equal(t, map[string]interface{}{"macro_1": "value_1", "macro_2": "value_2"}, mockResourceData.Get("macros"))

	// Test case 2: Error during read
	mockResourceDataWithError := resourceMacros().TestResourceData()
	mockResourceDataWithError.Set("project", "Default")
	mockResourceDataWithError.Set("macros", map[string]interface{}{"macro_1": "value_1"})

	mockClientWithError := &client.V1Client{
		GetProjectUIDFn: func(projectName string) (string, error) {
			return "", errors.New("failed to get project UID")
		},
	}

	diagsWithError := resourceMacrosRead(context.Background(), mockResourceDataWithError, mockClientWithError)

	// Assertions for error case
	assert.Equal(t, "failed to get project UID", diagsWithError[0].Summary)
	assert.Equal(t, "", mockResourceDataWithError.Id()) // ID should not be set on error

}

func TestResourceMacrosUpdate(t *testing.T) {
	// Test case 1: Successful update
	mockResourceData := resourceMacros().TestResourceData()
	mockResourceData.Set("project", "Default")
	mockResourceData.Set("macros", map[string]interface{}{"macro_1": "value_1"})

	mockClient := &client.V1Client{
		GetProjectUIDFn: func(projectName string) (string, error) {
			return "testUID", nil
		},
		GetExistMacrosFn: func(macros map[string]interface{}, uid string) ([]*models.V1Macro, error) {
			return []*models.V1Macro{
				{Name: "macro_1", Value: "value_1"},
				{Name: "macro_2", Value: "value_2"},
			}, nil
		},
		UpdateMacrosFn: func(uid string, updatedMacros *models.V1Macros) error {
			return nil
		},
	}

	diags := resourceMacrosUpdate(context.Background(), mockResourceData, mockClient)

	// Assertions for successful update
	var expectedDiag diag.Diagnostics
	assert.Equal(t, expectedDiag, diags)

	// Test case 2: Error during update
	mockResourceDataWithError := resourceMacros().TestResourceData()
	mockResourceDataWithError.Set("project", "Default")
	mockResourceDataWithError.Set("macros", map[string]interface{}{"macro_1": "value_1"})

	mockClientWithError := &client.V1Client{
		GetProjectUIDFn: func(projectName string) (string, error) {
			return "", errors.New("failed to get project UID")
		},
	}

	diagsWithError := resourceMacrosUpdate(context.Background(), mockResourceDataWithError, mockClientWithError)

	// Assertions for error case
	assert.Equal(t, "failed to get project UID", diagsWithError[0].Summary)
}

func TestResourceMacrosDelete(t *testing.T) {
	// Test case 1: Successful deletion
	mockResourceData := resourceMacros().TestResourceData()
	mockResourceData.Set("project", "Default")
	mockResourceData.Set("macros", map[string]interface{}{"macro_1": "value_1"})

	mockClient := &client.V1Client{
		GetProjectUIDFn: func(projectName string) (string, error) {
			return "testUID", nil
		},
		DeleteMacrosFn: func(uid string, macros *models.V1Macros) error {
			return nil
		},
	}

	diags := resourceMacrosDelete(context.Background(), mockResourceData, mockClient)

	// Assertions for successful deletion
	var expectedDiag diag.Diagnostics
	assert.Equal(t, expectedDiag, diags)

	// Test case 2: Error during deletion
	mockResourceDataWithError := resourceMacros().TestResourceData()
	mockResourceDataWithError.Set("project", "Default")
	mockResourceDataWithError.Set("macros", map[string]interface{}{"macro_1": "value_1"})

	mockClientWithError := &client.V1Client{
		GetProjectUIDFn: func(projectName string) (string, error) {
			return "", errors.New("failed to get project UID")
		},
	}

	diagsWithError := resourceMacrosDelete(context.Background(), mockResourceDataWithError, mockClientWithError)

	// Assertions for error case
	assert.Equal(t, "failed to get project UID", diagsWithError[0].Summary)
}

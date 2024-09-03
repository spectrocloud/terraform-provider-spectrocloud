package spectrocloud

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDataSourcePCGRead(t *testing.T) {
	resourceData := dataSourcePCG().TestResourceData()
	_ = resourceData.Set("name", "test-pcg-name")
	diags := dataSourcePCGRead(context.Background(), resourceData, unitTestMockAPIClient)

	assert.Len(t, diags, 0)
	assert.Equal(t, "test-pcg-name", resourceData.Get("name").(string))
	assert.NotEmpty(t, resourceData.Id())
}

func TestDataSourcePCGRead_MissingName(t *testing.T) {
	resourceData := dataSourcePCG().TestResourceData()

	diags := dataSourcePCGRead(context.Background(), resourceData, unitTestMockAPIClient)

	assert.Empty(t, diags)
}

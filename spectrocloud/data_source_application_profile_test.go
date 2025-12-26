package spectrocloud

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDataSourceApplicationProfileRead(t *testing.T) {
	resourceData := dataSourceApplicationProfile().TestResourceData()
	_ = resourceData.Set("name", "test-application-profile")

	diags := dataSourceApplicationProfileRead(context.Background(), resourceData, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
	assert.Equal(t, "test-application-profile", resourceData.Get("name").(string))
}

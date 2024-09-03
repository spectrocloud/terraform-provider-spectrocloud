package spectrocloud

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDataSourceIpPoolRead(t *testing.T) {
	resourceData := dataSourcePrivateCloudGatewayIpPool().TestResourceData()
	_ = resourceData.Set("private_cloud_gateway_id", "test-pcg-id")
	_ = resourceData.Set("name", "test-name")

	diags := dataSourceIpPoolRead(context.Background(), resourceData, unitTestMockAPIClient)

	assert.Len(t, diags, 0)
	assert.Equal(t, "test-pcg-id", resourceData.Get("private_cloud_gateway_id").(string))
	assert.Equal(t, "test-name", resourceData.Get("name").(string))
}

func TestDataSourceIpPoolRead_MissingFields(t *testing.T) {
	resourceData := dataSourcePrivateCloudGatewayIpPool().TestResourceData()
	diags := dataSourceIpPoolRead(context.Background(), resourceData, unitTestMockAPIClient)

	assert.NotEmpty(t, diags)
}

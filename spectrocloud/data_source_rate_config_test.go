package spectrocloud

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDataSourceRateConfigRead(t *testing.T) {
	d := dataSourceRateConfig().TestResourceData()

	diags := dataSourceRateConfigRead(context.Background(), d, unitTestMockAPIClient)

	assert.Empty(t, diags)
	assert.Equal(t, "test-tenant-uid", d.Id())

	aws := d.Get("aws").([]interface{})
	require.Len(t, aws, 1)
	computeOptimized := aws[0].(map[string]interface{})["compute_optimized"].([]interface{})
	require.Len(t, computeOptimized, 1)
	assert.Equal(t, 65.0, computeOptimized[0].(map[string]interface{})["compute_rate_proportion"])

	vsphere := d.Get("vsphere").([]interface{})
	require.Len(t, vsphere, 1)
	assert.Equal(t, 0.021811, vsphere[0].(map[string]interface{})["cpu_unit_price_per_hour"])

	custom := d.Get("custom").([]interface{})
	require.Len(t, custom, 1)
	assert.Equal(t, "test-custom-cloud", custom[0].(map[string]interface{})["cloud_type"])
}

func TestDataSourceRateConfigReadNegative(t *testing.T) {
	d := dataSourceRateConfig().TestResourceData()

	diags := dataSourceRateConfigRead(context.Background(), d, unitTestMockAPINegativeClient)

	require.NotEmpty(t, diags)
	assert.Contains(t, diags[0].Summary, "Invalid rate config")
	assert.Empty(t, d.Id())
}

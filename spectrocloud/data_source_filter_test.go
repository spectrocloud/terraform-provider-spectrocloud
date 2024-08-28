package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"testing"
)

func prepareBaseFilterResourceData() *schema.ResourceData {
	d := dataSourceFilter().TestResourceData()
	meta := make([]map[string]interface{}, 0)
	meta = append(meta, map[string]interface{}{
		"name": "test-filter-1",
		"annotations": map[string]string{
			"tag": "unit-test",
		},
		"labels": map[string]string{
			"label": "unit-test",
		},
	})
	err := d.Set("name", "test-filter-1")
	if err != nil {
		return nil
	}
	err = d.Set("metadata", meta)
	if err != nil {
		return nil
	}
	return d
}

func TestDataSourceFilterRead(t *testing.T) {
	d := prepareBaseFilterResourceData()
	ctx := context.Background()
	diags := dataSourceFilterRead(ctx, d, unitTestMockAPIClient)
	assert.Empty(t, diags)
}

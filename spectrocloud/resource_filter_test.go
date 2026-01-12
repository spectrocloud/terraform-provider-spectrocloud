package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"testing"
)

func prepareBaseFilterTestData() *schema.ResourceData {
	d := resourceFilter().TestResourceData()
	_ = d.Set("metadata", []interface{}{
		map[string]interface{}{
			"name": "test-filter-name",
		},
	})
	_ = d.Set("spec", []interface{}{
		map[string]interface{}{
			"filter_group": []interface{}{
				map[string]interface{}{
					"conjunction": "AND",
					"filters": []interface{}{
						map[string]interface{}{
							"key":      "test-key",
							"negation": false,
							"operator": "eq",
							"values":   []string{"test-value"},
						},
					},
				},
			},
		},
	})
	d.SetId("test-filter-id")
	return d
}

func TestResourceFilterCreate(t *testing.T) {
	d := prepareBaseFilterTestData()
	var ctx context.Context
	diags := resourceFilterCreate(ctx, d, unitTestMockAPIClient)
	assert.Empty(t, diags)
	assert.Equal(t, "test-filter-id", d.Id())
}

func TestResourceFilterRead(t *testing.T) {
	d := prepareBaseFilterTestData()
	var ctx context.Context
	diags := resourceFilterRead(ctx, d, unitTestMockAPIClient)
	assert.Empty(t, diags)
	assert.Equal(t, "test-filter-id", d.Id())
}

func TestResourceFilterUpdate(t *testing.T) {
	d := prepareBaseFilterTestData()
	var ctx context.Context
	diags := resourceFilterUpdate(ctx, d, unitTestMockAPIClient)
	assert.Empty(t, diags)
	assert.Equal(t, "test-filter-id", d.Id())
}

func TestResourceFilterDelete(t *testing.T) {
	d := prepareBaseFilterTestData()
	var ctx context.Context
	diags := resourceFilterDelete(ctx, d, unitTestMockAPIClient)
	assert.Empty(t, diags)
}
